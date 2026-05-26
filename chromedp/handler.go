package chromedp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	// "image"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
)

var (
	errMissingFilename = errors.New("filename field is required")
	errEmptyHTML       = errors.New("html field is required")
)

type QrRequest struct {
	Content   string `json:"content"`   //contains the data to encode in the QR code
	Dimension string `json:"dimension"` //specifies the pixel dimensions of the level 1 QR code
	Border    int    `json:"border"`    //optional border width around the QR code (default is 4)
	LogoURL   string `json:"logo_url"`  //optional URL of the logo image to embed in the center of the QR code
}

type RequestBody struct {
	HTML         string `json:"html"`
	Filename     string `json:"filename"`
	Size         string `json:"size"`
	CustomWidth  string `json:"custom_width"`
	CustomHeight string `json:"custom_height"`
}

type PageSize struct {
	Width  float64
	Height float64
}

var validSizes = map[string]PageSize{
	"A4":     {Width: 8.27, Height: 11.69},
	"Letter": {Width: 8.5, Height: 11},
	"Legal":  {Width: 8.5, Height: 14},
}

func sanitizeFilename(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", errMissingFilename
	}
	name = filepath.Base(name)
	if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		name += ".pdf"
	}
	return name, nil
}

func validateSize(size string, customWidth, customHeight float64) (PageSize, error) {
	if customWidth != 0 || customHeight != 0 {
		if size != "" {
			return PageSize{}, fmt.Errorf("size and custom_width/custom_height are mutually exclusive")
		}
		if customWidth <= 0 || customHeight <= 0 {
			return PageSize{}, fmt.Errorf("custom_width and custom_height must be positive")
		}
		return PageSize{Width: customWidth, Height: customHeight}, nil
	}

	if size == "" {
		size = "A4"
	}
	ps, ok := validSizes[size]
	if !ok {
		var lines []string
		for name, dims := range validSizes {
			lines = append(lines, fmt.Sprintf("  %s  width - %.2f  height - %.2f", name, dims.Width, dims.Height))
		}
		sort.Strings(lines)
		return PageSize{}, fmt.Errorf(
			"invalid size '%s'\nvalid sizes:\n%s\nor provide custom_width and custom_height instead",
			size, strings.Join(lines, "\n"),
		)
	}
	return ps, nil
}

func parseCustomDimensions(w, h string) (float64, float64, error) {
	if w == "" && h == "" {
		return 0, 0, nil
	}
	if w == "" || h == "" {
		return 0, 0, fmt.Errorf("custom_width and custom_height must both be provided together")
	}
	fw, err := strconv.ParseFloat(w, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid custom_width '%s': must be a number (e.g. 8.5)", w)
	}
	fh, err := strconv.ParseFloat(h, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid custom_height '%s': must be a number (e.g. 11)", h)
	}
	return fw, fh, nil
}

func GenerateHandler(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var req RequestBody
	if err := json.Unmarshal(body, &req); err != nil {
		slog.Error("failed to parse request JSON", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}

	if strings.TrimSpace(req.HTML) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": errEmptyHTML.Error()})
	}
	if len(req.Filename) > 50 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "filename must be 50 characters or less"})
	}

	filename, err := sanitizeFilename(req.Filename)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	w, h, err := parseCustomDimensions(req.CustomWidth, req.CustomHeight)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	pageSize, err := validateSize(req.Size, w, h)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	slog.Info("pdf request", "filename", filename, "html_length", len(req.HTML), "page_size", pageSize)

	Htmlbyts, err := base64.StdEncoding.DecodeString(req.HTML)
	if err != nil {
		slog.Error("failed to decode HTML content", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid base64 HTML content: " + err.Error()})
	}

	Html := string(Htmlbyts)

	Htmlerr := ValidateHTML(Html)
	if Htmlerr != nil {
		slog.Error("invalid HTML content", "error", Htmlerr)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid HTML content: " + Htmlerr.Error()})
	}

	path, err := GenerateNamed(Html, filename, pageSize.Width, pageSize.Height)
	if err != nil {
		slog.Error("pdf generation failed", "filename", filename, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.Attachment(path, filename)
}

func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"time":   time.Now().Format("2006-01-02 15:04:05 Monday"),
		"status": "OK",
	})
}

func ValidateHTML(html string) error {
	if strings.TrimSpace(html) == "" {
		return errors.New("HTML content cannot be empty")
	}

	lower := strings.ToLower(html)
	for _, tag := range []string{"<html", "<head", "<body"} {
		if !strings.Contains(lower, tag) {
			return fmt.Errorf("HTML content must contain <%s> tag", tag)
		}
	}

	if _, err := template.New("validate").Parse(html); err != nil {
		return fmt.Errorf("HTML content is not valid: %v", err)
	}
	return nil
}

func Qr1Handler(c echo.Context) error { // expects JSON body with "content", "dimension", and optional "logo_url" fields
	var req QrRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}
	if strings.TrimSpace(req.Content) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "content field is required"}) // 400 if content is empty
	}
	if strings.TrimSpace(req.Dimension) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "dimension field is required"}) // 400 if dimension is empty
	}
	dim, err := strconv.ParseInt(req.Dimension, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid dimension: " + err.Error()})
	}
	if dim <= 50 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "dimension must be greater than 50"}) // 400 if dimension is not greater than 50
	}
	if req.Border < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "border must be non-negative"}) // 400 if border is negative
	}
	if err := CreateQRWithLogo(req.Content, req.LogoURL, int(dim), req.Border); err != nil {
		slog.Error("failed to create QR code with logo", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create QR code: " + err.Error()})
	}
	return c.Attachment("qrcode_with_logo.png", "qrcode_with_logo.png")
}

func Qr2Handler(c echo.Context) error {
	var req QrRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}
	if strings.TrimSpace(req.Content) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "content field is required"}) // 400 if content is empty
	}
	if strings.TrimSpace(req.Dimension) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "dimension field is required"}) // 400 if dimension is empty
	}
	dim, err := strconv.ParseInt(req.Dimension, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid dimension: " + err.Error()})
	}
	if dim <= 50 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "dimension must be greater than 50"}) // 400 if dimension is not greater than 50
	}
	if req.Border < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "border must be non-negative"}) // 400 if border is negative
	}
	if err := CreateQRWithLogo2(req.Content, req.LogoURL, int(dim), req.Border); err != nil {
		slog.Error("failed to create QR code with logo", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create QR code: " + err.Error()})
	}
	return c.Attachment("qrcode_with_logo.png", "qrcode_with_logo.png")
}

func FetchLogoHandler(c echo.Context) error {
	return c.Attachment("logo.jpg", "download.jpg")
}
