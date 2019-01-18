package uploader

import (
	"github.com/ameliaikeda/sakura"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type uploader struct {
	manager         *s3manager.Uploader
	imageBucket     string
	thumbnailBucket string
}

func (u *uploader) UploadThumbnail(name string, buf []byte) error {
	return nil
}

func (u *uploader) UploadImage(name string, buf []byte) error {
	return nil
}

func New(manager *s3manager.Uploader, thumbnailBucket, imageBucket string) (sakura.Uploader, error) {
	if manager == nil {
		return nil, ErrNoManager
	}

	return &uploader{}, nil
}
