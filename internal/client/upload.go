package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

const (
	uploadChunkSize     int64 = 1 << 20
	uploadConcurrency         = 3
	queryUploadedChunks       = `query uploadedChunks($fileId: String!) { uploadedChunks(fileId: $fileId) }`
	mutationMergeChunks       = `mutation mergeChunks($fileId: String!, $totalChunks: Int!, $path: String!, $replace: Boolean!, $isAppFile: Boolean!) {
  mergeChunks(fileId: $fileId, totalChunks: $totalChunks, path: $path, replace: $replace, isAppFile: $isAppFile)
}`
)

type uploadInfo struct {
	FileID string `json:"fileId"`
	Index  int    `json:"index"`
	Size   int64  `json:"size"`
}

type uploadedChunksResponse struct {
	Data struct {
		UploadedChunks []int `json:"uploadedChunks"`
	} `json:"data"`
}

type mergeChunksResponse struct {
	Data struct {
		MergeChunks bool `json:"mergeChunks"`
	} `json:"data"`
}

type uploadJob struct {
	Index  int
	Offset int64
	Size   int64
}

func Upload(ctx context.Context, c *Client, localPath, remotePath string, progressFn func(done, total int64)) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	info, err := file.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("upload requires a regular file: %s", localPath)
	}

	fileID, err := buildUploadFileID(file, info)
	if err != nil {
		return err
	}

	totalSize := info.Size()
	totalChunks := chunkCount(totalSize)

	uploaded, err := fetchUploadedChunks(ctx, c, fileID)
	if err != nil {
		return err
	}

	var doneBytes atomic.Int64
	for index := range uploaded {
		doneBytes.Add(chunkSizeForIndex(totalSize, index))
	}
	if progressFn != nil {
		progressFn(doneBytes.Load(), totalSize)
	}

	jobs := make(chan uploadJob)
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg       sync.WaitGroup
		errOnce  sync.Once
		firstErr error
	)

	workerCount := uploadConcurrency
	if totalChunks < workerCount {
		workerCount = totalChunks
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-workerCtx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}

					if err := uploadChunk(workerCtx, c, file, fileID, job); err != nil {
						errOnce.Do(func() {
							firstErr = err
							cancel()
						})
						return
					}

					current := doneBytes.Add(job.Size)
					if progressFn != nil {
						progressFn(current, totalSize)
					}
				}
			}
		}()
	}

sendLoop:
	for index := 0; index < totalChunks; index++ {
		if uploaded[index] {
			continue
		}

		job := uploadJob{
			Index:  index,
			Offset: int64(index) * uploadChunkSize,
			Size:   chunkSizeForIndex(totalSize, index),
		}

		select {
		case <-workerCtx.Done():
			break sendLoop
		case jobs <- job:
		}
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return firstErr
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := workerCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	var mergeResp mergeChunksResponse
	if err := c.GraphQL(ctx, mutationMergeChunks, map[string]any{
		"fileId":      fileID,
		"isAppFile":   false,
		"path":        remotePath,
		"replace":     false,
		"totalChunks": totalChunks,
	}, &mergeResp); err != nil {
		return err
	}
	if !mergeResp.Data.MergeChunks {
		return errors.New("mergeChunks returned false")
	}

	return nil
}

func buildUploadFileID(file *os.File, info os.FileInfo) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	descriptor := fmt.Sprintf("%s|%d|%d|%s",
		filepath.Base(file.Name()),
		info.Size(),
		info.ModTime().UnixMilli(),
		hex.EncodeToString(hasher.Sum(nil)),
	)

	sum := sha256.Sum256([]byte(descriptor))
	return hex.EncodeToString(sum[:]), nil
}

func fetchUploadedChunks(ctx context.Context, c *Client, fileID string) (map[int]bool, error) {
	var resp uploadedChunksResponse
	if err := c.GraphQL(ctx, queryUploadedChunks, map[string]any{
		"fileId": fileID,
	}, &resp); err != nil {
		return nil, err
	}

	uploaded := make(map[int]bool, len(resp.Data.UploadedChunks))
	for _, index := range resp.Data.UploadedChunks {
		if index >= 0 {
			uploaded[index] = true
		}
	}

	return uploaded, nil
}

func uploadChunk(ctx context.Context, c *Client, file *os.File, fileID string, job uploadJob) error {
	infoBody, err := json.Marshal(uploadInfo{
		FileID: fileID,
		Index:  job.Index,
		Size:   job.Size,
	})
	if err != nil {
		return err
	}

	encryptedInfo, err := Encrypt(c.SessionKey, infoBody)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	infoPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{`form-data; name="info"`},
		"Content-Type":        []string{"application/octet-stream"},
	})
	if err != nil {
		return err
	}
	if _, err := infoPart.Write(encryptedInfo); err != nil {
		return err
	}

	filePart, err := writer.CreateFormFile("file", fmt.Sprintf("chunk-%d", job.Index))
	if err != nil {
		return err
	}
	if _, err := io.Copy(filePart, io.NewSectionReader(file, job.Offset, job.Size)); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	endpoint, err := resolveHTTPEndpoint(c.Host, "/upload_chunk")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return err
	}
	req.Header.Set("c-id", c.ClientID)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
		return nil
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	default:
		responseBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("upload chunk %d failed with status %d", job.Index, resp.StatusCode)
		}
		return fmt.Errorf("upload chunk %d failed with status %d: %s", job.Index, resp.StatusCode, bytes.TrimSpace(responseBody))
	}
}

func chunkCount(totalSize int64) int {
	if totalSize == 0 {
		return 1
	}

	return int((totalSize + uploadChunkSize - 1) / uploadChunkSize)
}

func chunkSizeForIndex(totalSize int64, index int) int64 {
	if totalSize == 0 {
		return 0
	}

	offset := int64(index) * uploadChunkSize
	remaining := totalSize - offset
	if remaining <= uploadChunkSize {
		return remaining
	}

	return uploadChunkSize
}
