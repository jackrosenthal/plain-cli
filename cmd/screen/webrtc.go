package screen

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

const (
	startScreenMirrorMutation = `mutation startScreenMirror($audio: Boolean!) {
  startScreenMirror(audio: $audio)
}`

	stopScreenMirrorMutation = `mutation {
  stopScreenMirror
}`

	sendWebRTCSignalingMutation = `mutation sendWebRtcSignaling($payload: WebRtcSignalingMessage!) {
  sendWebRtcSignaling(payload: $payload)
}`
)

type signalingPayload struct {
	Type          string `json:"type"`
	SDP           string `json:"sdp,omitempty"`
	Candidate     string `json:"candidate,omitempty"`
	SDPMid        string `json:"sdpMid,omitempty"`
	SDPMLineIndex *int   `json:"sdpMLineIndex,omitempty"`
	PhoneIP       string `json:"phoneIp,omitempty"`
}

type webRTCSession struct {
	pc     *webrtc.PeerConnection
	cancel context.CancelFunc
}

func (s *webRTCSession) close(apiClient *client.Client) {
	s.cancel()
	_ = s.pc.Close()
	var resp struct {
		Data struct {
			StopScreenMirror bool `json:"stopScreenMirror"`
		} `json:"data"`
	}
	_ = apiClient.GraphQL(context.Background(), stopScreenMirrorMutation, nil, &resp)
}

// newWebRTCSession starts screen mirroring and performs WebRTC negotiation,
// returning the session and first video track when ICE is connected.
func newWebRTCSession(ctx context.Context, apiClient *client.Client, audio bool) (*webRTCSession, *webrtc.TrackRemote, error) {
	sessCtx, sessCancel := context.WithCancel(ctx)

	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		sessCancel()
		return nil, nil, fmt.Errorf("register codecs: %w", err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		sessCancel()
		return nil, nil, fmt.Errorf("create peer connection: %w", err)
	}

	trackCh := make(chan *webrtc.TrackRemote, 1)

	pc.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		if track.Kind() == webrtc.RTPCodecTypeVideo {
			select {
			case trackCh <- track:
			default:
			}
		}
	})

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		init := c.ToJSON()
		mlineIdx := 0
		if init.SDPMLineIndex != nil {
			mlineIdx = int(*init.SDPMLineIndex)
		}
		sdpMid := ""
		if init.SDPMid != nil {
			sdpMid = *init.SDPMid
		}
		_ = sendWebRTCSignal(sessCtx, apiClient, signalingPayload{
			Type:          "ice_candidate",
			Candidate:     init.Candidate,
			SDPMid:        sdpMid,
			SDPMLineIndex: &mlineIdx,
		})
	})

	sigCh := make(chan client.WebRTCSignalingEvent, 32)
	evErrCh := make(chan error, 1)
	wsReadyCh := make(chan struct{}, 1)
	mirrorReadyCh := make(chan struct{}, 1)

	go func() {
		err := client.ConnectEventsOnce(sessCtx, apiClient, client.EventHandler{
			OnConnected: func() {
				select {
				case wsReadyCh <- struct{}{}:
				default:
				}
			},
			ScreenMirroring: func(_ client.ScreenMirroringEvent) {
				select {
				case mirrorReadyCh <- struct{}{}:
				default:
				}
			},
			WebRTCSignaling: func(e client.WebRTCSignalingEvent) {
				select {
				case sigCh <- e:
				default:
				}
			},
		})
		if err != nil && sessCtx.Err() == nil {
			evErrCh <- err
		}
	}()

	select {
	case <-wsReadyCh:
	case err := <-evErrCh:
		sessCancel()
		_ = pc.Close()
		return nil, nil, fmt.Errorf("event stream: %w", err)
	case <-sessCtx.Done():
		sessCancel()
		_ = pc.Close()
		return nil, nil, sessCtx.Err()
	}

	var startResp struct {
		Data struct {
			StartScreenMirror bool `json:"startScreenMirror"`
		} `json:"data"`
	}
	if err := apiClient.GraphQL(sessCtx, startScreenMirrorMutation, map[string]any{"audio": audio}, &startResp); err != nil {
		sessCancel()
		_ = pc.Close()
		return nil, nil, fmt.Errorf("start screen mirror: %w", err)
	}
	if !startResp.Data.StartScreenMirror {
		sessCancel()
		_ = pc.Close()
		return nil, nil, fmt.Errorf("start screen mirror: mutation returned false")
	}

	fmt.Fprintln(os.Stderr, "grant permission on phone...")
	select {
	case <-mirrorReadyCh:
	case err := <-evErrCh:
		sessCancel()
		_ = pc.Close()
		return nil, nil, fmt.Errorf("event stream: %w", err)
	case <-sessCtx.Done():
		sessCancel()
		_ = pc.Close()
		return nil, nil, sessCtx.Err()
	}

	phoneIP, err := extractHostname(apiClient.Host)
	if err != nil {
		sessCancel()
		_ = pc.Close()
		return nil, nil, fmt.Errorf("extract phone IP: %w", err)
	}
	if err := sendWebRTCSignal(sessCtx, apiClient, signalingPayload{
		Type:    "ready",
		PhoneIP: phoneIP,
	}); err != nil {
		sessCancel()
		_ = pc.Close()
		return nil, nil, fmt.Errorf("send ready signal: %w", err)
	}

	var remoteDescSet bool
	for {
		select {
		case <-sessCtx.Done():
			sessCancel()
			_ = pc.Close()
			return nil, nil, sessCtx.Err()
		case err := <-evErrCh:
			sessCancel()
			_ = pc.Close()
			return nil, nil, fmt.Errorf("event stream: %w", err)
		case track := <-trackCh:
			return &webRTCSession{pc: pc, cancel: sessCancel}, track, nil
		case sig := <-sigCh:
			switch sig.Type {
			case "offer":
				if err := pc.SetRemoteDescription(webrtc.SessionDescription{
					Type: webrtc.SDPTypeOffer,
					SDP:  sig.SDP,
				}); err != nil {
					sessCancel()
					_ = pc.Close()
					return nil, nil, fmt.Errorf("set remote description: %w", err)
				}
				remoteDescSet = true
				answer, err := pc.CreateAnswer(nil)
				if err != nil {
					sessCancel()
					_ = pc.Close()
					return nil, nil, fmt.Errorf("create answer: %w", err)
				}
				if err := pc.SetLocalDescription(answer); err != nil {
					sessCancel()
					_ = pc.Close()
					return nil, nil, fmt.Errorf("set local description: %w", err)
				}
				if err := sendWebRTCSignal(sessCtx, apiClient, signalingPayload{
					Type: "answer",
					SDP:  answer.SDP,
				}); err != nil {
					sessCancel()
					_ = pc.Close()
					return nil, nil, fmt.Errorf("send answer: %w", err)
				}
			case "ice_candidate":
				if !remoteDescSet {
					break
				}
				sdpMid := sig.SDPMid
				mlineIdx := uint16(sig.SDPMLineIndex)
				if err := pc.AddICECandidate(webrtc.ICECandidateInit{
					Candidate:     sig.Candidate,
					SDPMid:        &sdpMid,
					SDPMLineIndex: &mlineIdx,
				}); err != nil {
					// non-fatal: some candidates may be incompatible
					_ = err
				}
			}
		}
	}
}

func extractHostname(host string) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}

func sendWebRTCSignal(ctx context.Context, apiClient *client.Client, payload signalingPayload) error {
	return apiClient.GraphQL(ctx, sendWebRTCSignalingMutation, map[string]any{
		"payload": payload,
	}, nil)
}

// resolvePlayer returns the player binary and args for the given player preference.
// "auto" tries mpv first, then ffplay.
func resolvePlayer(pref, sdpPath string) (string, []string, error) {
	playerArgs := map[string][]string{
		"mpv": {
			"--no-terminal",
			"--profile=low-latency",
			"--cache=no",
			"--demuxer-readahead-secs=0",
			"--demuxer-lavf-analyzeduration=0.1",
			"--demuxer-lavf-probesize=32768",
			sdpPath,
		},
		"ffplay": {
			"-protocol_whitelist", "file,rtp,udp",
			"-fflags", "nobuffer",
			"-flags", "low_delay",
			"-analyzeduration", "100000",
			"-probesize", "32768",
			"-i", sdpPath,
			"-autoexit",
		},
	}
	candidates := []string{pref}
	if pref == "auto" {
		candidates = []string{"mpv", "ffplay"}
	}
	for _, name := range candidates {
		if _, err := exec.LookPath(name); err == nil {
			return name, playerArgs[name], nil
		}
	}
	if pref == "auto" {
		return "", nil, fmt.Errorf("no player found: install mpv or ffplay")
	}
	return "", nil, fmt.Errorf("player %q not found in PATH", pref)
}

// setupRTPForwarding picks a free UDP port and writes a minimal SDP temp file
// describing the track's codec. Returns the port, SDP file path, and a cleanup func.
func setupRTPForwarding(codec webrtc.RTPCodecParameters) (port int, sdpPath string, cleanup func(), err error) {
	ln, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		return 0, "", nil, fmt.Errorf("find free port: %w", err)
	}
	port = ln.LocalAddr().(*net.UDPAddr).Port
	_ = ln.Close()

	codecName := strings.ToUpper(strings.TrimPrefix(codec.MimeType, "video/"))
	sdpContent := fmt.Sprintf(
		"v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=Plain Screen\r\nc=IN IP4 127.0.0.1\r\nt=0 0\r\nm=video %d RTP/AVP %d\r\na=rtpmap:%d %s/%d\r\n",
		port, codec.PayloadType, codec.PayloadType, codecName, codec.ClockRate,
	)

	f, err := os.CreateTemp("", "plain-screen-*.sdp")
	if err != nil {
		return 0, "", nil, fmt.Errorf("create SDP file: %w", err)
	}
	if _, err := f.WriteString(sdpContent); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return 0, "", nil, fmt.Errorf("write SDP: %w", err)
	}
	_ = f.Close()

	return port, f.Name(), func() { _ = os.Remove(f.Name()) }, nil
}

// forwardRTP reads RTP packets from track and sends them as raw UDP to addr
// until the context is done or a write error occurs. It sends periodic RTCP
// PLI requests so the sender emits fresh keyframes.
func forwardRTP(ctx context.Context, pc *webrtc.PeerConnection, track *webrtc.TrackRemote, addr string) {
	conn, err := net.Dial("udp4", addr)
	if err != nil {
		slog.Warn("forwardRTP dial error", "err", err)
		return
	}
	defer func() { _ = conn.Close() }()

	// Request keyframe immediately (burst) and then every 3 s so the decoder can start.
	sendPLI := func() {
		_ = pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{
			MediaSSRC: uint32(track.SSRC()),
		}})
	}
	// Burst to maximise chance phone responds before next natural keyframe.
	for i := 0; i < 3; i++ {
		sendPLI()
		time.Sleep(50 * time.Millisecond)
	}
	pliTicker := time.NewTicker(3 * time.Second)
	defer pliTicker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pliTicker.C:
				sendPLI()
			}
		}
	}()

	first := true
	for {
		pkt, _, err := track.ReadRTP()
		if err != nil {
			slog.Debug("forwardRTP read stopped", "err", err)
			return
		}
		if first {
			slog.Debug("first RTP packet", "seq", pkt.SequenceNumber, "ssrc", pkt.SSRC)
			first = false
		}
		raw, err := pkt.Marshal()
		if err != nil {
			continue
		}
		if _, err := conn.Write(raw); err != nil {
			return
		}
	}
}
