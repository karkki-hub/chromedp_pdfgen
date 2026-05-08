package chromedp

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/labstack/echo/v4"
)

const (
	paperWidthIn  = 8.27  // A4
	paperHeightIn = 11.69 // A4
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

func Generate(html string) (string, error) {
	tmpFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		return "", err
	}
	tmpFile.Close()

	outputPath := tmpFile.Name()
	if err = renderToFile(html, outputPath); err != nil {
		os.Remove(outputPath)
		return "", err
	}

	return outputPath, nil
}

func GenerateNamed(html, filename string) (string, error) {
	os.MkdirAll("pdfs", 0o755)
	outputPath := filepath.Join("pdfs", filename)
	if err := renderToFile(html, outputPath); err != nil {
		return "", err
	}
	return outputPath, nil
}

func renderToFile(html, outputPath string) error {
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
				WithPaperWidth(paperWidthIn).
				WithPaperHeight(paperHeightIn).
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

func GenerateHandler(c echo.Context) error {
	// Read HTML body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	html := string(body)

	// Validate HTML
	if err := ValidateBody(html); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Get filename from URL param
	rawFilename := c.Param("filename")

	filename, err := SanitizeFilename(rawFilename)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	slog.Info("pdf request",
		"filename", filename,
		"html_length", len(html),
	)

	// Generate PDF
	path, err := GenerateNamed(html, filename)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.Attachment(path, filename)
}
