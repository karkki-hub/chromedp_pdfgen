package qrgen

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"

	qrgode "github.com/ahmedtahas/qr-gode"
)

// content - data to encode in the QR code
// dimension - pixel dimensions of the QR code image (e.g. 300 for 300x300)
// border - optional border width around the QR code (default is 4)
// logoURL - optional URL of the logo image to embed in the center of the QR code
func CreateQRWithLogo2(content string, logoURL string, dimension int, border int) error {

	builder := qrgode.New(content).
		Size(dimension).
		QuietZone(border).
		LogoBackground("transparent").
		LogoMode(qrgode.LogoOverlay).
		ErrorCorrection(qrgode.LevelH)

	if logoURL != "" {
		if err := UrlGet(logoURL); err != nil {
			fmt.Printf("failed to fetch logo: %v\n", err)
			return err
		}

		builder = builder.Logo("logo1.jpg")

		for _, w := range builder.ScannabilityWarnings() {
			fmt.Printf("scannability warning: %s\n", w)
		}
	}

	// Get PNG bytes from the library
	pngBytes, err := builder.PNG()
	if err != nil {
		fmt.Printf("generate qrcode failed: %v\n", err)
		return err
	}

	// Decode the generated PNG
	qrImg, err := png.Decode(bytesReader(pngBytes))
	if err != nil {
		fmt.Printf("decode qrcode failed: %v\n", err)
		return err
	}

	bounds := qrImg.Bounds()
	white := image.NewRGBA(bounds)
	draw.Draw(white, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(white, bounds, qrImg, image.Point{}, draw.Over)

	out, err := os.Create("qrcode_with_logo.png")
	if err != nil {
		return err
	}
	defer out.Close()

	if err = png.Encode(out, white); err != nil {
		fmt.Printf("save qrcode failed: %v\n", err)
		return err
	}

	fmt.Println("QR code saved to qrcode_with_logo.png")
	return nil
}

// bytesReader wraps a byte slice as an io.Reader.
func bytesReader(b []byte) io.Reader {
	return &bytesReaderImpl{b: b, pos: 0}
}

type bytesReaderImpl struct {
	b   []byte
	pos int
}

func (r *bytesReaderImpl) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.b) {
		return 0, io.EOF
	}
	n = copy(p, r.b[r.pos:])
	r.pos += n
	return n, nil
}
