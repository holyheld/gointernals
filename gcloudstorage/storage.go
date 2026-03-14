package gcloudstorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	gostorage "cloud.google.com/go/storage"
	"github.com/holyheld/gointernals/storage"
)

// Storage is a storage wrapper struct exposing methods that are useful for
// the caller.
type Storage struct {
	client    *gostorage.Client
	handle    *gostorage.BucketHandle
	timeout   time.Duration
	chunkSize *int
}

var _ storage.Storage = (*Storage)(nil)

// WithTimeout sets the timeout limit on inner contexts to prevent
// long requests from bloating goroutines scheduler.
func WithTimeout(timeout time.Duration) func(*Storage) {
	return func(s *Storage) {
		s.timeout = timeout
	}
}

// WithChunkSize limits the maximum chunk size used by storage.Writer
//
// Note that retries are not supported for chunk size 0.
func WithChunkSize(size int) func(*Storage) {
	return func(s *Storage) {
		s.chunkSize = &size
	}
}

func NewBucket(ctx context.Context, name string, opts ...func(*Storage)) (*Storage, error) {
	client, err := gostorage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	handle := client.Bucket(name)

	s := &Storage{
		client:  client,
		handle:  handle,
		timeout: time.Second * 30,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Close closes the inner client
//
// Note that inner client becomes useless after that.
func (s *Storage) Close() error {
	err := s.client.Close()
	if err != nil {
		return fmt.Errorf("failed to close inner client: %w", err)
	}

	return nil
}

// DownloadFileReader creates full file download reader.
func (s *Storage) DownloadFileReader(ctx context.Context, name string) (io.ReadCloser, error) {
	return s.downloadRangeReader(ctx, name, 0, -1)
}

// DownloadFileBytes downloads the file using file name.
func (s *Storage) DownloadFileBytes(ctx context.Context, name string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	r, err := s.downloadRangeReader(ctx, name, 0, -1)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(r)
}

// DownloadRangeReader creates range file download reader.
func (s *Storage) DownloadRangeReader(
	ctx context.Context,
	name string,
	offset int64,
	length int64,
) (io.ReadCloser, error) {
	return s.downloadRangeReader(ctx, name, offset, length)
}

// Downloader creates [io.ReadSeeker].
//
// Useful for serving files where full content download ahead of time is expensive
// (e.g. with [http.ServeContent]).
func (s *Storage) Downloader(ctx context.Context, name string, size int64) io.ReadSeeker {
	return &gcsReadSeeker{
		ctx:  ctx,
		name: name,
		size: size,
		s:    s,
	}
}

// UploadFile uploads the file using file name and provided reader.
func (s *Storage) UploadFile(ctx context.Context, name string, r io.Reader) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	objectHandle := s.handle.Object(name)
	w := objectHandle.NewWriter(ctx)

	if s.chunkSize != nil {
		w.ChunkSize = *s.chunkSize
	}

	defer func() {
		_ = w.Close()
	}()

	_, err := io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("failed to write object to bucket: %w", err)
	}

	return nil
}

// UpdateFile updates object attributes.
func (s *Storage) UpdateFile(
	ctx context.Context,
	name string,
	attrs storage.UpdateAttributes,
) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	meta := make(map[string]string, 0)
	if !attrs.CustomTime.IsZero() {
		meta["expires"] = attrs.CustomTime.UTC().Format(time.UnixDate)
	}

	_, err := s.handle.Object(name).Update(
		ctx, gostorage.ObjectAttrsToUpdate{
			Metadata:    meta,
			ContentType: attrs.ContentType,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update object attributes: %w", err)
	}

	return nil
}

func (s *Storage) Attributes(ctx context.Context, name string) (*storage.ObjectAttributes, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	attrs, err := s.handle.Object(name).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes: %w", err)
	}

	var exp *time.Time

	emeta, ok := attrs.Metadata["expires"]
	if ok {
		e, err := time.Parse(time.UnixDate, emeta)
		if err == nil {
			e = e.UTC()
			exp = &e
		}
	}

	return &storage.ObjectAttributes{
		ETag:           attrs.Etag,
		ExpirationTime: exp,
		UpdatedTime:    attrs.Updated,
		ContentType:    attrs.ContentType,
		Size:           attrs.Size,
	}, nil
}

// DeleteFile removes file from bucket storage.
func (s *Storage) DeleteFile(ctx context.Context, name string) error {
	err := s.handle.Object(name).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

func (s *Storage) downloadRangeReader(
	ctx context.Context,
	name string,
	offset int64,
	length int64,
) (io.ReadCloser, error) {
	objectHandle := s.handle.Object(name)

	r, err := objectHandle.NewRangeReader(ctx, offset, length)
	if err != nil {
		if errors.Is(err, gostorage.ErrObjectNotExist) {
			return nil, storage.ErrNoSuchFile
		}

		return nil, fmt.Errorf("failed to create reader on object: %w", err)
	}

	return r, nil
}
