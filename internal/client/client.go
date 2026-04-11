package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"time"
)

const graphqlContentType = "multipart/form-data"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrTimeout      = errors.New("connection_timeout")

	serverTimePattern = regexp.MustCompile(`window\.__SERVER_TIME__\s*=\s*(\d+)`)
)

type Client struct {
	Host        string
	ClientID    string
	SessionKey  []byte
	URLTokenKey []byte

	serverTimeOffset int64
	httpClient       *http.Client
}

func NewClient(host, clientID, token string) (*Client, error) {
	sessionKey, err := DeriveSessionKey(token)
	if err != nil {
		return nil, err
	}

	return &Client{
		Host:       host,
		ClientID:   clientID,
		SessionKey: sessionKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Client) FetchServerTimeOffset(ctx context.Context) error {
	endpoint, err := resolveHTTPEndpoint(c.Host, "/")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("c-id", c.ClientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return ErrTimeout
		}
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch server time failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	matches := serverTimePattern.FindSubmatch(body)
	if len(matches) != 2 {
		return errors.New("server time not found in response")
	}

	var serverTimeMS int64
	if _, err := fmt.Sscanf(string(matches[1]), "%d", &serverTimeMS); err != nil {
		return fmt.Errorf("parse server time: %w", err)
	}

	c.serverTimeOffset = serverTimeMS - time.Now().UnixMilli()

	return nil
}

func (c *Client) FetchURLToken(ctx context.Context) error {
	var resp struct {
		Data struct {
			App struct {
				URLToken string `json:"urlToken"`
			} `json:"app"`
		} `json:"data"`
	}
	if err := c.GraphQL(ctx, `query { app { urlToken } }`, nil, &resp); err != nil {
		return err
	}
	key, err := base64.StdEncoding.DecodeString(resp.Data.App.URLToken)
	if err != nil {
		return fmt.Errorf("decode url token: %w", err)
	}
	c.URLTokenKey = key
	return nil
}

func (c *Client) GraphQL(ctx context.Context, query string, variables map[string]any, result any) error {
	payload := struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}{
		Query:     query,
		Variables: variables,
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	serverTimeMS := time.Now().UnixMilli() + c.serverTimeOffset
	wrappedPayload := WrapGraphQL(rawPayload, serverTimeMS)

	encryptedPayload, err := Encrypt(c.SessionKey, wrappedPayload)
	if err != nil {
		return err
	}

	endpoint, err := resolveHTTPEndpoint(c.Host, "/graphql")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encryptedPayload))
	if err != nil {
		return err
	}
	req.Header.Set("c-id", c.ClientID)
	req.Header.Set("Content-Type", graphqlContentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return ErrTimeout
		}
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	default:
		return fmt.Errorf("graphql request failed with status %d", resp.StatusCode)
	}

	encryptedResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	plaintext, err := Decrypt(c.SessionKey, encryptedResponse)
	if err != nil {
		return err
	}

	if result == nil {
		return nil
	}

	return json.Unmarshal(plaintext, result)
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
