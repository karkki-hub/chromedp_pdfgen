package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/karkki-hub/chromedp_pdfgen/chromedp"
)

func main() {
	e := echo.New()

	e.Use(middleware.Recover())

	e.Static("/", "UI")

	e.POST("/v1/generatepdf", chromedp.GenerateHandler)
	e.GET("/health", chromedp.HealthHandler)

	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		slog.Error("shutting down server", "error", err)
		os.Exit(1)
	}
}
