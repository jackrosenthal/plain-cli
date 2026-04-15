package client

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

const initContentType = "multipart/form-data"

type loginRequest struct {
	Password       string `json:"password"`
	BrowserName    string `json:"browserName"`
	BrowserVersion string `json:"browserVersion"`
	OSName         string `json:"osName"`
	OSVersion      string `json:"osVersion"`
	IsMobile       bool   `json:"isMobile"`
}

type loginResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func InitLogin(ctx context.Context, host, clientID string, sessionKey []byte) (valid bool, prefilledPassword string, err error) {
	endpoint, err := resolveHTTPEndpoint(host, "/init")
	if err != nil {
		return false, "", err
	}

	var body io.Reader = http.NoBody
	if sessionKey != nil {
		u, err := uuid.NewRandom()
		if err != nil {
			return false, "", err
		}
		tokenProbeID := u.String()

		payload, err := Encrypt(sessionKey, []byte(tokenProbeID))
		if err != nil {
			return false, "", err
		}
		body = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return false, "", err
	}
	req.Header.Set("c-id", clientID)
	req.Header.Set("Content-Type", initContentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusForbidden:
		return false, "", errors.New("web access disabled on this device")
	default:
		return false, "", fmt.Errorf("init request failed with status %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}

	if len(respBody) == 0 {
		return true, "", nil
	}

	return false, string(respBody), nil
}

func CheckToken(ctx context.Context, host, clientID string, sessionKey []byte) (valid bool, err error) {
	valid, _, err = InitLogin(ctx, host, clientID, sessionKey)
	return valid, err
}

func Login(ctx context.Context, host, clientID, password string, onPending ...func()) (token string, err error) {
	endpoint, err := resolveWebSocketEndpoint(host, clientID)
	if err != nil {
		return "", err
	}

	loginKey := DeriveLoginKey(password)
	requestBody, err := buildLoginPayload(password)
	if err != nil {
		return "", err
	}

	encryptedRequest, err := Encrypt(loginKey, requestBody)
	if err != nil {
		return "", err
	}

	conn, _, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{
		HTTPHeader: http.Header{
			"c-id": []string{clientID},
		},
	})
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	if err := conn.Write(ctx, websocket.MessageBinary, encryptedRequest); err != nil {
		return "", err
	}

	pendingCalled := false

	for {
		messageType, payload, err := conn.Read(ctx)
		if err != nil {
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) && closeErr.Reason != "" {
				return "", errors.New(closeErr.Reason)
			}
			return "", err
		}

		if messageType != websocket.MessageBinary {
			return "", fmt.Errorf("unexpected websocket message type %d", messageType)
		}

		plaintext, err := Decrypt(loginKey, payload)
		if err != nil {
			return "", err
		}

		var response loginResponse
		if err := json.Unmarshal(plaintext, &response); err != nil {
			return "", err
		}

		if response.Status == "PENDING" {
			if !pendingCalled && len(onPending) > 0 && onPending[0] != nil {
				onPending[0]()
				pendingCalled = true
			}
			continue
		}

		if response.Token != "" {
			return response.Token, nil
		}

		return "", errors.New("authentication did not return a token")
	}
}

func buildLoginPayload(password string) ([]byte, error) {
	payload := loginRequest{
		Password:       sha512Hex(password),
		BrowserName:    "Chrome",
		BrowserVersion: "120",
		OSName:         "Windows",
		OSVersion:      "10",
		IsMobile:       false,
	}

	return json.Marshal(payload)
}

func resolveHTTPEndpoint(host, requestPath string) (string, error) {
	baseURL, err := parseBaseURL(host)
	if err != nil {
		return "", err
	}

	baseURL.Path = requestPath
	baseURL.RawQuery = ""
	baseURL.Fragment = ""

	return baseURL.String(), nil
}

func resolveWebSocketEndpoint(host, clientID string) (string, error) {
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
	baseURL.RawQuery = url.Values{
		"auth": []string{"1"},
		"cid":  []string{clientID},
	}.Encode()
	baseURL.Fragment = ""

	return baseURL.String(), nil
}

func parseBaseURL(host string) (*url.URL, error) {
	if !strings.Contains(host, "://") {
		host = "https://" + host
	}

	parsed, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	if parsed.Host == "" {
		return nil, fmt.Errorf("invalid host %q", host)
	}

	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}

	return parsed, nil
}

func sha512Hex(value string) string {
	sum := sha512.Sum512([]byte(value))
	return hex.EncodeToString(sum[:])
}
