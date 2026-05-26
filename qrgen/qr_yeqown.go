package qrgen

import (
	"fmt"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// content - data to encode in the QR code
// dimension - pixel dimensions of the QR code image (e.g. 300 for 300x300)
// border - optional border width around the QR code (default is 4)
// logoURL - optional URL of the logo image to embed in the center of the QR code
func CreateQRWithLogo(content string, logoURL string, dimension int, border int) error {
	fmt.Printf("Creating QR code with content: %s, logoURL: %s, dimension: %d, border: %d\n",
		content, logoURL, dimension, border)

	qr, err := qrcode.New(content)
	if err != nil {
		fmt.Printf("create qrcode failed: %v\n", err)
		return err
	}

	version := qrVersionFromContent(content)
	modules := (version-1)*4 + 21
	qrWidth := uint8(dimension / modules)
	if qrWidth < 1 {
		qrWidth = 1
	}

	fmt.Printf("QR version: %d, modules: %d, qrWidth per module: %d\n", version, modules, qrWidth)

	var options []standard.ImageOption

	if logoURL != "" {
		if err = UrlGet(logoURL); err != nil {
			fmt.Printf("failed to fetch logo: %v\n", err)
			return err
		}

		nativeQRSize := int(qrWidth) * modules
		targetLogoSize := nativeQRSize / 5

		fmt.Printf("Native QR size: %dpx, target logo size: %dpx\n", nativeQRSize, targetLogoSize)

		if err = resizeLogoToTarget("logo1.jpg", targetLogoSize); err != nil {
			fmt.Printf("failed to resize logo: %v\n", err)
			return err
		}

		options = []standard.ImageOption{
			standard.WithLogoImageFileJPEG("logo1.jpg"),
			standard.WithQRWidth(qrWidth),
			standard.WithBorderWidth(border),
		}
	} else {
		options = []standard.ImageOption{
			standard.WithQRWidth(qrWidth),
			standard.WithBorderWidth(border),
		}
	}

	writer, err := standard.New("qrcode_with_logo.png", options...)
	if err != nil {
		fmt.Printf("create writer failed: %v\n", err)
		return err
	}
	defer writer.Close()

	if err = qr.Save(writer); err != nil {
		fmt.Printf("save qrcode failed: %v\n", err)
		return err
	}

	if err = resizeImage("qrcode_with_logo.png", dimension); err != nil {
		fmt.Printf("resize failed: %v\n", err)
		return err
	}

	return nil
}
