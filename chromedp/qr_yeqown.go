package chromedp

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// createQRWithLogo generates a QR code using the WithLogo option.
func CreateQRWithLogo(content string, logoURL string, dimension int) error {
	var options []standard.ImageOption

	fmt.Printf("Creating QR code with content: %s, logoURL: %s, dimension: %d\n",
		content, logoURL, dimension)

	qr, err := qrcode.New(content)
	if err != nil {
		fmt.Printf("create qrcode failed: %v\n", err)
		return err
	}

	if logoURL != "" { // If a logo URL is provided, fetch the logo and include it in the QR code options
		err = UrlGet(logoURL)
		if err != nil {
			fmt.Printf("failed to fetch logo: %v\n", err)
			return err
		}

		options = []standard.ImageOption{ // Set the logo image and QR code width based on the dimension
			standard.WithLogoImageFileJPEG("logo1.jpg"),
			standard.WithQRWidth(reversePattern(dimension)),
			standard.WithBorderWidth(0),
		}
	} else { // If no logo URL is provided, just set the QR code width based on the dimension
		options = []standard.ImageOption{
			standard.WithQRWidth(reversePattern(dimension)),
			standard.WithBorderWidth(0),
		}
	}

	writer, err := standard.New("qrcode_with_logo.png", options...) // Create a new writer with the specified options
	if err != nil {
		fmt.Printf("create writer failed: %v\n", err)
		return err
	}
	defer writer.Close()

	if err = qr.Save(writer); err != nil { // Save the QR code using the writer, which will generate the image file
		fmt.Printf("save qrcode failed: %v\n", err)
		return err
	}

	return nil
}

func UrlGet(url string) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Pretend to be a browser
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	outfile, err := os.Create("logo1.jpg")
	if err != nil {
		return err
	}
	defer outfile.Close()

	_, err = io.Copy(outfile, resp.Body)
	return err
}

func reversePattern(value int) uint8 { // The pattern size is determined by the dimension divided by 21, which is the number of modules in a version 1 QR code.

	a := uint8(value / 21)
	fmt.Println("Calculated pattern size:", a, value)
	return a
}
