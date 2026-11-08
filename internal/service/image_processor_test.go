package service

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"automarket/internal/domain"
)

func encodedImage(t *testing.T, format string, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 80, A: 255})
		}
	}
	var buffer bytes.Buffer
	var err error
	if format == "png" {
		err = png.Encode(&buffer, img)
	} else {
		err = jpeg.Encode(&buffer, img, &jpeg.Options{Quality: 95})
	}
	if err != nil {
		t.Fatalf("encode image: %v", err)
	}
	return buffer.Bytes()
}

func TestImageProcessorNormalizesJPEGAndRejectsFakeFiles(t *testing.T) {
	processor := NewImageProcessor()
	photo, err := processor.Process(RawPhoto{Name: `..\unsafe/name.png`, Data: encodedImage(t, "png", 100, 100)})
	if err != nil {
		t.Fatalf("process png: %v", err)
	}
	if photo.ContentType != "image/jpeg" || photo.Width != 100 || photo.Height != 100 || len(photo.Data) == 0 {
		t.Fatalf("normalized photo = %+v", photo)
	}
	if photo.OriginalName != "name.png" {
		t.Fatalf("sanitized name = %q", photo.OriginalName)
	}
	if _, err := processor.Process(RawPhoto{Name: "fake.jpg", Data: []byte("not an image")}); err == nil {
		t.Fatal("fake image was accepted")
	}
	if _, err := processor.Process(RawPhoto{Name: "tiny.jpg", Data: encodedImage(t, "jpeg", 10, 10)}); err != domain.ErrPhotoDimensions {
		t.Fatalf("tiny image error = %v", err)
	}
}
