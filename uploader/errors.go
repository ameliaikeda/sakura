package uploader

import "github.com/pkg/errors"

var (
	// ErrNoManager is returned if you give a nil manager to New
	ErrNoManager = errors.New("sakura/uploader: no *s3manager.Uploader instance was given")
)
