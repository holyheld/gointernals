package gcloudstorage

import (
	"context"
	"errors"
	"io"
)

type gcsReadSeeker struct {
	ctx    context.Context
	name   string
	size   int64
	offset int64
	s      *Storage
}

func (rs *gcsReadSeeker) Read(p []byte) (n int, err error) {
	if rs.offset >= rs.size {
		return 0, io.EOF
	}

	rc, err := rs.s.downloadRangeReader(rs.ctx, rs.name, rs.offset, int64(len(p)))
	if err != nil {
		return 0, err
	}
	defer rc.Close()

	n, err = rc.Read(p)
	rs.offset += int64(n)

	return n, err
}

func (rs *gcsReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = rs.offset + offset
	case io.SeekEnd:
		newOffset = rs.size + offset
	}
	if newOffset < 0 || newOffset > rs.size {
		return 0, errors.New("invalid seek offset")
	}
	rs.offset = newOffset

	return rs.offset, nil
}
