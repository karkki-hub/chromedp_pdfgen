package chromedp

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"

	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const (
	renderTimeout = 20 * time.Second
	renderSettle  = 2 * time.Second
)

var (
	ErrMissingFilename = errors.New("X-PDF-Name header is required")
	ErrEmptyBody       = errors.New("HTML body must not be empty")
)

func SanitizeFilename(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", ErrMissingFilename
	}

	name = filepath.Base(name)
	if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		name += ".pdf"
	}

	return name, nil
}

func ValidateBody(html string) error {
	if strings.TrimSpace(html) == "" {
		return ErrEmptyBody
	}
	return nil
}

func GenerateNamed(html, filename string, width, height float64) (string, error) {
	if err := os.MkdirAll("/temp", 0o755); err != nil {
		return "", err
	}
	outputPath := filepath.Join("/temp", filename)
	if err := renderToFile(html, outputPath, width, height); err != nil {
		return "", err
	}
	return outputPath, nil
}

func renderToFile(html, outputPath string, width, height float64) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, renderTimeout)
	defer cancel()

	wrapped := `<html><head><meta charset="UTF-8">` +
		`<style>body{margin:0;}</style></head><body>` + html + `</body></html>`

	htmlURL := "data:text/html," + url.PathEscape(wrapped)

	var pdfBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(htmlURL),
		chromedp.Sleep(renderSettle),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(width).
				WithPaperHeight(height).
				Do(ctx)
			pdfBuf = buf
			return err
		}),
	)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, pdfBuf, 0o644)
}
