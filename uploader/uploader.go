package uploader

import (
	"bytes"
	"github.com/ameliaikeda/sakura"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// uploader is an implementation of sakura.Uploader that uploads images to s3-compatible storage.
type uploader struct {
	manager *s3manager.Uploader
	cfg     sakura.UploaderConfig
}

// UploadThumbnail uploads an image to the thumbnail bucket on s3-compatible storage.
func (u *uploader) UploadThumbnail(name string, buf []byte) error {
	return u.upload(name, u.cfg.ThumbnailBucket, buf)
}

// UploadImage uploads an image to the thumbnail bucket on s3-compatible storage.
func (u *uploader) UploadImage(name string, buf []byte) error {
	return u.upload(name, u.cfg.ImageBucket, buf)
}

// New creates a new uploader instance configured to upload to s3-compatible storage.
func New(manager *s3manager.Uploader, cfg sakura.UploaderConfig) sakura.Uploader {
	return &uploader{
		manager: manager,
		cfg:     cfg,
	}
}

// upload is common between UploadThumbnail and UploadImage.
func (u *uploader) upload(name string, bucket string, buf []byte) error {
	// make a Reader so the buffer can't be manipulated.
	buffer := bytes.NewReader(buf)

	_, err := u.manager.Upload(&s3manager.UploadInput{
		ACL:    aws.String("public-read"),
		Bucket: &bucket,
		Body:   buffer,
		Key:    &name,
	})

	//u.log.WithFields(logrus.Fields{
	//	"bucket": bucket,
	//	"name": name,
	//	"len": len(buf),
	//	"uri": output.Location,
	//}).Info("uploaded image to s3")

	return err
}
