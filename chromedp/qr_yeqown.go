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
func CreateQRWithLogo(content string) {
	qr, err := qrcode.New(content)
	if err != nil {
		fmt.Printf("create qrcode failed: %v\n", err)
		return
	}

	getLogoURL := "https://localhost8080/fetchlogo"
	errl := UrlGet(getLogoURL)
	if errl != nil {
		fmt.Printf("failed to fetch logo: %v\n", errl)
		return
	}

	options := []standard.ImageOption{
		// standard.WithLogoImageFileJPEG("logo.jpg"),
		standard.WithQRWidth(70),
	}
	writer, err := standard.New("qrcode_with_logo.png", options...)
	if err != nil {
		fmt.Printf("create writer failed: %v\n", err)
		return
	}
	outfile, err := os.Create("qrcode_with_logo.png")
	if err != nil {
		fmt.Printf("create output file failed: %v\n", err)
		return
	}
	defer outfile.Close()

	defer writer.Close()
	if err = qr.Save(writer); err != nil {
		fmt.Printf("save qrcode failed: %v\n", err)
	}
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

	outfile, err := os.Create("logo.jpg")
	if err != nil {
		return err
	}
	defer outfile.Close()

	_, err = io.Copy(outfile, resp.Body)
	return err
}
