package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type fileIDPayload struct {
	Path    string `json:"path"`
	MediaID string `json:"mediaId"`
}

func EncodeFileID(sessionKey []byte, path, mediaID string) (string, error) {
	var plaintext []byte
	if mediaID == "" {
		plaintext = []byte(path)
	} else {
		payload, err := json.Marshal(fileIDPayload{
			Path:    path,
			MediaID: mediaID,
		})
		if err != nil {
			return "", err
		}

		plaintext = payload
	}

	ciphertext, err := Encrypt(sessionKey, plaintext)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DownloadFile(ctx context.Context, c *Client, path, mediaID string) (io.ReadCloser, error) {
	key := c.URLTokenKey
	if len(key) == 0 {
		key = c.SessionKey
	}
	fileID, err := EncodeFileID(key, path, mediaID)
	if err != nil {
		return nil, err
	}

	endpoint, err := resolveHTTPEndpoint(c.Host, "/fs")
	if err != nil {
		return nil, err
	}

	requestURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	requestURL.RawQuery = url.Values{
		"dl": []string{"1"},
		"id": []string{fileID},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("c-id", c.ClientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return nil, ErrTimeout
		}
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return resp.Body, nil
	case http.StatusUnauthorized:
		_ = resp.Body.Close()
		return nil, ErrUnauthorized
	case http.StatusForbidden:
		_ = resp.Body.Close()
		return nil, ErrForbidden
	default:
		_ = resp.Body.Close()
		return nil, fmt.Errorf("download file failed with status %d", resp.StatusCode)
	}
}
