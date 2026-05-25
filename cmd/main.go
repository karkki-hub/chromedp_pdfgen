package main

import (
	// "fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/karkki-hub/chromedp_pdfgen/chromedp"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// err := chromedp.Qr()

	// content := "https://github.com/yeqown/go-qrcode"
	// chromedp.CreateQRWithLogo(content)

	// fmt.Println("All QR codes generated successfully.")

	// if err != nil {
	// 	slog.Error("failed to generate QR code", "error", err)
	// 	os.Exit(1)
	// }
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogMethod:    true,
		LogLatency:   true,
		LogRequestID: true,
		LogError:     true,
		HandleError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			level := slog.LevelInfo
			if v.Error != nil || v.Status >= 500 {
				level = slog.LevelError
			} else if v.Status >= 400 {
				level = slog.LevelWarn
			}
			slog.Log(c.Request().Context(), level, "request",
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency,
				"request_id", v.RequestID,
				"error", v.Error,
			)
			return nil
		},
	}))

	e.Static("/", "UI")
	e.POST("/v1/generatepdf", chromedp.GenerateHandler)
	e.GET("/health", chromedp.HealthHandler)
	e.POST("/qr1", chromedp.Qr1Handler)
	e.POST("/qr2", chromedp.Qr2Handler)
	e.GET("/fetchlogo", chromedp.FetchLogoHandler)

	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		slog.Error("shutting down server", "error", err)
		os.Exit(1)
	}
}
