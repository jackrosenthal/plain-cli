package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

const (
	eventTypeMessageCreated  byte = 1
	eventTypeMessageDeleted  byte = 2
	eventTypeMessageUpdated  byte = 3
	eventTypeFeedsFetched    byte = 4
	eventTypeScreenMirroring byte = 5
	eventTypeWebRTCSignaling byte = 6

	eventTypeNotificationCreated      byte = 7
	eventTypeNotificationUpdated      byte = 8
	eventTypeNotificationDeleted      byte = 9
	eventTypeNotificationRefreshed    byte = 10
	eventTypePomodoroAction           byte = 11
	eventTypePomodoroSettingsUpdate   byte = 12
	eventTypeMessageCleared           byte = 13
	eventTypeScreenMirrorAudioGranted byte = 14
	eventTypeBookmarkUpdated          byte = 15
	eventTypeDownloadProgress         byte = 16
	eventTypeReserved17               byte = 17
	eventTypeChannelsUpdated          byte = 18
	eventTypeImageSearchUpdated       byte = 19
)

type EventHandler struct {
	MessageCreated           func(MessageCreatedEvent)
	MessageDeleted           func(MessageDeletedEvent)
	MessageUpdated           func(MessageUpdatedEvent)
	FeedsFetched             func(FeedsFetchedEvent)
	ScreenMirroring          func(ScreenMirroringEvent)
	WebRTCSignaling          func(WebRTCSignalingEvent)
	NotificationCreated      func(NotificationCreatedEvent)
	NotificationUpdated      func(NotificationUpdatedEvent)
	NotificationDeleted      func(NotificationDeletedEvent)
	NotificationRefreshed    func(NotificationRefreshedEvent)
	PomodoroAction           func(PomodoroActionEvent)
	PomodoroSettingsUpdate   func(PomodoroSettingsUpdateEvent)
	MessageCleared           func(MessageClearedEvent)
	ScreenMirrorAudioGranted func(ScreenMirrorAudioGrantedEvent)
	BookmarkUpdated          func(BookmarkUpdatedEvent)
	DownloadProgress         func(DownloadProgressEvent)
	Reserved17               func(Reserved17Event)
	ChannelsUpdated          func(ChannelsUpdatedEvent)
	ImageSearchUpdated       func(ImageSearchUpdatedEvent)
}

type TagPayload struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type MessageAttachmentPayload struct {
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
	Name        string `json:"name"`
}

type MessagePayload struct {
	ID             string                     `json:"id"`
	Body           string                     `json:"body"`
	Address        string                     `json:"address"`
	ServiceCenter  string                     `json:"serviceCenter"`
	Date           string                     `json:"date"`
	Type           int                        `json:"type"`
	ThreadID       string                     `json:"threadId"`
	SubscriptionID string                     `json:"subscriptionId"`
	IsMMS          bool                       `json:"isMms"`
	Attachments    []MessageAttachmentPayload `json:"attachments"`
	Tags           []TagPayload               `json:"tags"`
}

type MessageCreatedEvent struct {
	MessagePayload
}

type MessageDeletedEvent struct {
	ID string `json:"id"`
}

type MessageUpdatedEvent struct {
	MessagePayload
}

type FeedsFetchedEvent struct{}

type ScreenMirroringEvent struct {
	State string `json:"state"`
}

type WebRTCSignalingEvent struct {
	Type          string `json:"type"`
	SDP           string `json:"sdp"`
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
}

type NotificationPayload struct {
	ID           string   `json:"id"`
	OnlyOnce     bool     `json:"onlyOnce"`
	IsClearable  bool     `json:"isClearable"`
	AppID        string   `json:"appId"`
	AppName      string   `json:"appName"`
	Time         string   `json:"time"`
	Silent       bool     `json:"silent"`
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	Actions      []string `json:"actions"`
	ReplyActions []string `json:"replyActions"`
}

type NotificationCreatedEvent struct {
	NotificationPayload
}

type NotificationUpdatedEvent struct {
	NotificationPayload
}

type NotificationDeletedEvent struct {
	ID string `json:"id"`
}

type NotificationRefreshedEvent struct{}

type PomodoroActionEvent struct {
	Date           string `json:"date"`
	CompletedCount int    `json:"completedCount"`
	CurrentRound   int    `json:"currentRound"`
	TimeLeft       int    `json:"timeLeft"`
	TotalTime      int    `json:"totalTime"`
	IsRunning      bool   `json:"isRunning"`
	IsPause        bool   `json:"isPause"`
	State          string `json:"state"`
}

type PomodoroSettingsUpdateEvent struct {
	WorkDuration             int  `json:"workDuration"`
	ShortBreakDuration       int  `json:"shortBreakDuration"`
	LongBreakDuration        int  `json:"longBreakDuration"`
	PomodorosBeforeLongBreak int  `json:"pomodorosBeforeLongBreak"`
	ShowNotification         bool `json:"showNotification"`
	PlaySoundOnComplete      bool `json:"playSoundOnComplete"`
}

type MessageClearedEvent struct {
	ThreadID string `json:"threadId"`
}

type ScreenMirrorAudioGrantedEvent struct{}

type BookmarkUpdatedEvent struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	Title         string `json:"title"`
	FaviconPath   string `json:"faviconPath"`
	GroupID       string `json:"groupId"`
	Pinned        bool   `json:"pinned"`
	ClickCount    int    `json:"clickCount"`
	LastClickedAt string `json:"lastClickedAt"`
	SortOrder     int    `json:"sortOrder"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type DownloadProgressEvent struct {
	ID       string  `json:"id"`
	Progress float64 `json:"progress"`
	Done     bool    `json:"done"`
}

type Reserved17Event struct {
	Payload json.RawMessage `json:"-"`
}

type ChatChannelMemberPayload struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ChatChannelPayload struct {
	ID        string                     `json:"id"`
	Name      string                     `json:"name"`
	Owner     string                     `json:"owner"`
	Members   []ChatChannelMemberPayload `json:"members"`
	Version   int                        `json:"version"`
	Status    string                     `json:"status"`
	CreatedAt string                     `json:"createdAt"`
	UpdatedAt string                     `json:"updatedAt"`
}

type ChannelsUpdatedEvent struct {
	Channels []ChatChannelPayload
}

type ImageSearchUpdatedEvent struct {
	Status           string  `json:"status"`
	DownloadProgress float64 `json:"downloadProgress"`
	ErrorMessage     string  `json:"errorMessage"`
	ModelSize        int     `json:"modelSize"`
	ModelDir         string  `json:"modelDir"`
	IsIndexing       bool    `json:"isIndexing"`
	TotalImages      int     `json:"totalImages"`
	IndexedImages    int     `json:"indexedImages"`
}

func ConnectEvents(ctx context.Context, c *Client, h EventHandler) error {
	backoff := time.Second

	for {
		if err := connectEventsOnce(ctx, c, h); err == nil && ctx.Err() != nil {
			return ctx.Err()
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}

		if backoff < 5*time.Second {
			backoff += time.Second
		}
	}
}

func connectEventsOnce(ctx context.Context, c *Client, h EventHandler) error {
	endpoint, err := resolveEventsEndpoint(c.Host, c.ClientID)
	if err != nil {
		return err
	}

	conn, _, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{
		HTTPHeader: http.Header{
			"c-id": []string{c.ClientID},
		},
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	clockSync, err := Encrypt(c.SessionKey, []byte(fmt.Sprintf("%d", time.Now().UnixMilli()+c.serverTimeOffset)))
	if err != nil {
		return err
	}

	if err := conn.Write(ctx, websocket.MessageBinary, clockSync); err != nil {
		return err
	}

	for {
		messageType, payload, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}

		if messageType != websocket.MessageBinary {
			return fmt.Errorf("unexpected websocket message type %d", messageType)
		}
		if len(payload) < 1 {
			return errors.New("event frame missing type byte")
		}

		eventType := payload[0]
		plaintext, err := Decrypt(c.SessionKey, payload[1:])
		if err != nil {
			return err
		}

		if err := dispatchEvent(eventType, plaintext, h); err != nil {
			return err
		}
	}
}

func resolveEventsEndpoint(host, clientID string) (string, error) {
	baseURL, err := parseBaseURL(host)
	if err != nil {
		return "", err
	}

	switch baseURL.Scheme {
	case "https":
		baseURL.Scheme = "wss"
	case "http":
		baseURL.Scheme = "ws"
	default:
		return "", fmt.Errorf("unsupported host scheme %q", baseURL.Scheme)
	}

	baseURL.Path = "/"
	baseURL.RawQuery = "cid=" + clientID
	baseURL.Fragment = ""

	return baseURL.String(), nil
}

func dispatchEvent(eventType byte, plaintext []byte, h EventHandler) error {
	switch eventType {
	case eventTypeMessageCreated:
		return decodeAndHandle(plaintext, h.MessageCreated)
	case eventTypeMessageDeleted:
		return decodeAndHandle(plaintext, h.MessageDeleted)
	case eventTypeMessageUpdated:
		return decodeAndHandle(plaintext, h.MessageUpdated)
	case eventTypeFeedsFetched:
		return decodeAndHandle(plaintext, h.FeedsFetched)
	case eventTypeScreenMirroring:
		return decodeAndHandle(plaintext, h.ScreenMirroring)
	case eventTypeWebRTCSignaling:
		return decodeAndHandle(plaintext, h.WebRTCSignaling)
	case eventTypeNotificationCreated:
		return decodeAndHandle(plaintext, h.NotificationCreated)
	case eventTypeNotificationUpdated:
		return decodeAndHandle(plaintext, h.NotificationUpdated)
	case eventTypeNotificationDeleted:
		return decodeAndHandle(plaintext, h.NotificationDeleted)
	case eventTypeNotificationRefreshed:
		return decodeAndHandle(plaintext, h.NotificationRefreshed)
	case eventTypePomodoroAction:
		return decodeAndHandle(plaintext, h.PomodoroAction)
	case eventTypePomodoroSettingsUpdate:
		return decodeAndHandle(plaintext, h.PomodoroSettingsUpdate)
	case eventTypeMessageCleared:
		return decodeAndHandle(plaintext, h.MessageCleared)
	case eventTypeScreenMirrorAudioGranted:
		return decodeAndHandle(plaintext, h.ScreenMirrorAudioGranted)
	case eventTypeBookmarkUpdated:
		return decodeAndHandle(plaintext, h.BookmarkUpdated)
	case eventTypeDownloadProgress:
		return decodeAndHandle(plaintext, h.DownloadProgress)
	case eventTypeReserved17:
		if h.Reserved17 != nil {
			h.Reserved17(Reserved17Event{Payload: append(json.RawMessage(nil), plaintext...)})
		}
		return nil
	case eventTypeChannelsUpdated:
		if h.ChannelsUpdated == nil {
			return nil
		}

		var channels []ChatChannelPayload
		if err := unmarshalEvent(plaintext, &channels); err != nil {
			return err
		}

		h.ChannelsUpdated(ChannelsUpdatedEvent{Channels: channels})
		return nil
	case eventTypeImageSearchUpdated:
		return decodeAndHandle(plaintext, h.ImageSearchUpdated)
	default:
		return fmt.Errorf("unknown event type %d", eventType)
	}
}

func decodeAndHandle[T any](payload []byte, handler func(T)) error {
	if handler == nil {
		return nil
	}

	var event T
	if err := unmarshalEvent(payload, &event); err != nil {
		return err
	}

	handler(event)
	return nil
}

func unmarshalEvent(payload []byte, dst any) error {
	if len(payload) == 0 {
		payload = []byte("null")
	}

	return json.Unmarshal(payload, dst)
}
