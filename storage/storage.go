package storage

import (
	"context"
	"io"
)

type Storage interface {
	io.Closer
	// DownloadFileReader creates full file download reader.
	DownloadFileReader(ctx context.Context, name string) (io.ReadCloser, error)

	// DownloadFileBytes downloads the file using file name.
	DownloadFileBytes(ctx context.Context, name string) ([]byte, error)

	// DownloadRangeReader creates range file download reader.
	DownloadRangeReader(
		ctx context.Context,
		name string,
		offset int64,
		length int64,
	) (io.ReadCloser, error)

	// Downloader creates [io.ReadSeeker] as file download reader.
	//
	// Useful for serving files where full content download ahead of time is expensive
	// (e.g. with [http.ServeContent]).
	Downloader(ctx context.Context, name string, size int64) io.ReadSeeker

	// UploadFile uploads the file to the storage using file name and provided reader.
	UploadFile(ctx context.Context, name string, r io.Reader) error

	// UpdateFile updates object attributes.
	UpdateFile(ctx context.Context, name string, attrs UpdateAttributes) error

	// Attributes retrieves object attributes.
	Attributes(ctx context.Context, name string) (*ObjectAttributes, error)

	// DeleteFile removes file from the storage.
	DeleteFile(ctx context.Context, name string) error
}
