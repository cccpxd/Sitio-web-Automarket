package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/jpeg"
	_ "image/png"

	xdraw "golang.org/x/image/draw"

	"automarket/internal/domain"
)

const (
	normalizedMaxSide = 1600
	jpegQuality       = 85
)

// ImageProcessor validates the real file content and normalizes accepted
// images to a bounded JPEG before they reach PostgreSQL.
type ImageProcessor struct{}

func NewImageProcessor() *ImageProcessor { return &ImageProcessor{} }

func (ImageProcessor) Process(raw RawPhoto) (domain.CarPhoto, error) {
	if len(raw.Data) == 0 {
		return domain.CarPhoto{}, domain.ErrPhotoDecodeFailed
	}
	if len(raw.Data) > domain.MaxPhotoBytes {
		return domain.CarPhoto{}, domain.ErrPhotoTooLarge
	}

	contentType := detectContentType(raw.Data)
	if !domain.IsAllowedPhotoType(contentType) {
		return domain.CarPhoto{}, domain.ErrPhotoTypeInvalid
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(raw.Data))
	if err != nil {
		return domain.CarPhoto{}, domain.ErrPhotoDecodeFailed
	}
	if err := domain.ValidatePhotoUpload(contentType, int64(len(raw.Data)), config.Width, config.Height); err != nil {
		return domain.CarPhoto{}, err
	}

	source, format, err := image.Decode(bytes.NewReader(raw.Data))
	if err != nil {
		return domain.CarPhoto{}, domain.ErrPhotoDecodeFailed
	}
	if format != "jpeg" && format != "png" {
		return domain.CarPhoto{}, domain.ErrPhotoTypeInvalid
	}

	normalized := resizeToMax(source, normalizedMaxSide)
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, normalized, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return domain.CarPhoto{}, domain.ErrPhotoWriteFailed
	}

	encoded := buffer.Bytes()
	sum := sha256.Sum256(encoded)
	return domain.CarPhoto{
		ContentType:  "image/jpeg",
		ByteSize:     int64(len(encoded)),
		SHA256:       hex.EncodeToString(sum[:]),
		Width:        normalized.Bounds().Dx(),
		Height:       normalized.Bounds().Dy(),
		OriginalName: domain.SanitizeOriginalName(raw.Name),
		Data:         encoded,
	}, nil
}

func detectContentType(data []byte) string {
	head := data
	if len(head) > 512 {
		head = head[:512]
	}
	return detectContent(head)
}

func detectContent(head []byte) string {
	switch {
	case bytes.HasPrefix(head, []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg"
	case bytes.HasPrefix(head, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return "image/png"
	default:
		return "application/octet-stream"
	}
}

func resizeToMax(source image.Image, maxSide int) image.Image {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	if width <= maxSide && height <= maxSide {
		return source
	}

	targetWidth, targetHeight := width, height
	if width >= height {
		targetWidth = maxSide
		targetHeight = height * maxSide / width
	} else {
		targetHeight = maxSide
		targetWidth = width * maxSide / height
	}
	if targetWidth < 1 {
		targetWidth = 1
	}
	if targetHeight < 1 {
		targetHeight = 1
	}

	target := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	xdraw.CatmullRom.Scale(target, target.Bounds(), source, bounds, xdraw.Src, nil)
	return target
}
