// Package processor handles taking
package processor

import (
	"github.com/ameliaikeda/sakura"
	"github.com/pkg/errors"
	"gopkg.in/h2non/bimg.v1"
	"io"
	"io/ioutil"
)

const (
	// DefaultQuality is the default JPEG quality given to main images.
	DefaultQuality = 97

	// DefaultThumbnailQuality is the default JPEG quality given to thumbnails.
	DefaultThumbnailQuality = 95
)

// supported is the list of values we support for image generation.
var supported = map[string]bimg.ImageType{
	"jpeg": bimg.JPEG,
}

type processor struct {
	cfg              sakura.GeneratorConfig
	format           bimg.ImageType
	quality          int
	thumbnailQuality int
}

// ProcessImage
func (p *processor) ProcessImage(body io.Reader) ([]byte, error) {
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, errors.Wrap(err, errBodyReadFailed)
	}

	img := bimg.NewImage(buf)

	// when processing an image on its own, we should simply change it to JPEG with the given quality.
	// TODO: constrain the image to max 8192x8192?

	opts := bimg.Options{
		Quality:       p.quality,
		StripMetadata: true,
		Type:          p.format,
	}

	b, err := img.Process(opts)
	if err != nil {
		return nil, errors.Wrap(err, errImageProcessFailed)
	}

	return b, nil
}

func (p *processor) ProcessThumbnail(body io.Reader) ([]byte, error) {
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, errors.Wrap(err, errBodyReadFailed)
	}

	img := bimg.NewImage(buf)

	size, err := img.Size()
	if err != nil {
		return nil, errors.Wrap(err, errImageSizeUnknown)
	}

	width, height := p.ratio(size)

	opts := bimg.Options{
		Width:         int(width),
		Height:        int(height),
		Quality:       p.thumbnailQuality,
		StripMetadata: true,
		Type:          p.format,
	}

	b, err := img.Process(opts)
	if err != nil {
		return nil, errors.Wrap(err, errImageProcessFailed)
	}

	return b, nil
}

// ratio takes details about an image's size and returns an aspect-ratio-constrained
// tuple of (width, height), configured to a max width and height set on the processor.
func (p *processor) ratio(meta bimg.ImageSize) (uint, uint) {
	// local variables to make this easier
	width, height := uint(meta.Width), uint(meta.Height)
	maxWidth, maxHeight := p.cfg.MaxWidth, p.cfg.MaxHeight

	if maxWidth/maxHeight > width/height {
		return width * maxHeight / height, maxHeight
	}

	return maxWidth, height * maxWidth / width
}

// New creates a new Generator from a config instance
func New(cfg sakura.GeneratorConfig) sakura.Generator {
	format, ok := supported[cfg.ImageType]
	if !ok {
		format = bimg.JPEG
	}

	quality, thumbnailQuality := cfg.Quality, cfg.ThumbnailQuality

	if quality == 0 {
		quality = DefaultQuality
	}

	if thumbnailQuality == 0 {
		thumbnailQuality = DefaultThumbnailQuality
	}

	return &processor{
		cfg:              cfg,
		format:           format,
		quality:          int(quality),
		thumbnailQuality: int(thumbnailQuality),
	}
}
