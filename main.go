package main

import (
	"log/slog"
	"os"

	"tech-utils/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.Info("request",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.Error("request error",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("error", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	logger.Info("starting server", slog.String("port", cfg.Port))
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
