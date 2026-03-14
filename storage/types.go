package storage

import (
	"errors"
	"time"
)

type UpdateAttributes struct {
	ContentType string
	CustomTime  time.Time
}

type ObjectAttributes struct {
	ETag           string     `json:"etag"`
	ExpirationTime *time.Time `json:"expirationTime"`
	ContentType    string     `json:"contentType"`
	UpdatedTime    time.Time  `json:"updatedTime"`
	Size           int64      `json:"size"`
}

var ErrNoSuchFile = errors.New("no such file")
