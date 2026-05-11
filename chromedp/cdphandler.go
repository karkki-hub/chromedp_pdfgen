package chromedp

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RequestBody struct {
	HTML     string `json:"html"`
	Filename string `json:"filename"`
	Size     string `json:"size"` // A4, Letter, Legal
}

func GenerateHandler(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var req RequestBody
	if err := json.Unmarshal(body, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}

	if req.HTML == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "html field is required"})
	}

	if req.Filename == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "filename field is required"})
	}

	if len(req.Filename) > 50 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "filename must be 50 characters or less"})
	}

	if err := ValidateBody(req.HTML); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	filename, err := SanitizeFilename(req.Filename)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	slog.Info("pdf request", "filename", filename, "html_length", len(req.HTML))

	path, err := GenerateNamed(req.HTML, filename, req.Size)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.Attachment(path, filename)
}
