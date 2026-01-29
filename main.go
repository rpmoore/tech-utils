package main

import (
	"embed"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"

	"tech-utils/config"
	"tech-utils/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	e := echo.New()
	e.HideBanner = true

	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
	e.Renderer = &Template{templates: tmpl}

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

	e.GET("/static/*", echo.WrapHandler(http.FileServer(http.FS(staticFS))))

	homeHandler := handlers.NewHomeHandler()
	e.GET("/", homeHandler.Index)

	uuidHandler := handlers.NewUUIDHandler()
	e.GET("/uuid", uuidHandler.Index)
	e.POST("/uuid/generate", uuidHandler.Generate)

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	logger.Info("starting server", slog.String("port", cfg.Port))
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
