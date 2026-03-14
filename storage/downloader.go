package storage

import (
	"context"
	"fmt"
	"net/http"
)

// DownloadByURLThenUpload performs a file download with default client, then uploads result
// to specified location in bucket.
func DownloadByURLThenUpload(ctx context.Context, s Storage, url string, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build request to download: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	if resp.StatusCode < 199 || resp.StatusCode > 299 {
		return fmt.Errorf("non-ok status: %d", resp.StatusCode)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	err = s.UploadFile(ctx, name, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}
