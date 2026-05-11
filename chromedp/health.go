package chromedp

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"start_time": time.Now().Format("2006-01-02 15:04:05 Monday"),
		"status":     "OK",
	})
}
