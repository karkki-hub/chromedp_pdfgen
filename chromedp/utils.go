package chromedp

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"

	"github.com/nfnt/resize"
)

func qrVersionFromContent(content string) int {
	contentLen := len(content)

	// ISO 18004 byte-mode capacity at error correction level M
	capacities := []int{
		14, 26, 42, 62, 84, 106, 122, 154, 180, 213, // v1-10
		251, 287, 331, 370, 411, 461, 512, 549, 597, 648, // v11-20
		702, 742, 823, 875, 916, 1000, 1062, 1128, 1193, 1267, // v21-30
		1373, 1455, 1541, 1631, 1725, 1812, 1914, 1992, 2102, 2216, // v31-40
	}

	for version, capacity := range capacities {
		if contentLen <= capacity {
			return version + 1 // versions are 1-indexed
		}
	}
	return 40 // max version
}

// resizeLogoToTarget resizes the logo at path to targetSize x targetSize in place.
func resizeLogoToTarget(path string, targetSize int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("decode logo: %w", err)
	}

	resized := resize.Resize(uint(targetSize), uint(targetSize), img, resize.Lanczos3)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Printf("Resized logo to %dx%d\n", targetSize, targetSize)
	return jpeg.Encode(out, resized, &jpeg.Options{Quality: 95})
}

// resize image and overwrites the file as PNG.
func resizeImage(path string, size int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	img, format, err := image.Decode(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("decode image (format=%s): %w", format, err)
	}

	resized := resize.Resize(uint(size), uint(size), img, resize.Lanczos3)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	return png.Encode(out, resized)
}

// UrlGet downloads the resource at url and saves it as logo1.jpg.
func UrlGet(url string) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
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
