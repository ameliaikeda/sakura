// Package main is the implementation of "sakura", an image optimizer and s3 uploader.
//
// The main package just handles ensuring all environment variables are set, as well
// as setting up the clients used for dependency injection, and then starting the webserver.
//
// sakura should be ideally called using environment variables, but also supports flags.
//
//	sakura \
//		-aws-endpoint http://localhost:8081 -aws-secret-key abcdef -aws-access-key-id ABCDEF123 \
//		-image-bucket lolibrary-prod-images -thumbnail-bucket lolibrary-prod-thumbnails \
//		-host 0.0.0.0 -port 3000 -max-height 300 -max-width 400
//
// Environment Variables:
//
// - `AWS_ACCESS_KEY_ID`: AWS Access Key ID for access to s3.
// - `AWS_SECRET_KEY`: AWS Secret Key for access to s3.
// - `AWS_ENDPOINT`: The custom AWS endpoint to upload to (for DigitalOcean spaces, integration tests, etc).
// - `MAX_HEIGHT`: The maximum bounded height for images when thumbnailling.
// - `MAX_WIDTH`: The maximum bounded width for images when thumbnailling.
// - `IMAGE_BUCKET`: The s3 bucket used for uploading main images.
// - `THUMBNAIL_BUCKET`: The s3 bucket used for uploading thumbnails.
// - `HOST`: The host to bind the http server to.
// - `PORT`: The port to bind the http server to.
package main

import (
	"github.com/ameliaikeda/sakura"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/facebookgo/httpdown"
	"github.com/namsral/flag"

	"log"
	"net/http"
	"time"
)

// aws credential configuration.
var awsEndpoint, awsAccessKeyID, awsSecretKey string

// sakura-specific configuration
var (
	maxHeight, maxWidth uint
	imageBucket, thumbnailBucket string
)

// http webserver parameters
var (
	host string
	port uint
)

func init() {
	flag.StringVar(&host, "host", "0.0.0.0", "the host to bind the http server to.")
	flag.UintVar(&port, "port", 3000, "the port to bind the http server to.")
	flag.UintVar(&maxHeight, "max-height", 300, "the maximum bounded height for images when generating thumbnails")
	flag.UintVar(&maxWidth, "max-width", 300, "the maximum bounded width for images when generating thumbnails")
	flag.StringVar(&imageBucket, "image-bucket", "images", "the image bucket used for main images")
	flag.StringVar(&thumbnailBucket, "thumbnail-bucket", "thumbnails", "the image bucket used for thumbnail images")
	flag.StringVar(&awsEndpoint, "aws-endpoint", "", "custom AWS endpoint to upload to (for digitalocean spaces, integration tests, etc)")
	flag.StringVar(&awsAccessKeyID, "aws-access-key-id", "", "AWS Access Key ID for access to s3")
	flag.StringVar(&awsSecretKey, "aws-secret-key", "", "AWS Secret Key for access to s3")

	flag.Parse()
}

func main() {
	// new up an s3 uploader
	manager := s3()

	// new up a sakura thumbnailer
	uploader := sakura.NewUploader(&sakura.UploaderConfig{
		Manager: manager,
		ImageBucket: imageBucket,
		ThumbnailBucket: thumbnailBucket,
	})

	// create a sakura client which doubles as a handler.
	s := sakura.New(uploader, &sakura.Config{
		MaxHeight: maxHeight,
		MaxWidth: maxWidth,
	})

	// create the HTTP server and pass in the handler
	srv := &http.Server{
		ReadTimeout: time.Minute * 1,
		WriteTimeout: time.Minute * 1,
		Handler: s,
	}

	if err := httpdown.ListenAndServe(srv, &httpdown.HTTP{}); err != nil {
		log.Fatal(err)
	}
}

func s3() *s3manager.Uploader {
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint: &awsEndpoint,

	}))

	// Create an uploader with the session and default options
	return s3manager.NewUploader(sess)
}
