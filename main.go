package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Item struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Enriched bool   `json:"enriched"`
}

func validateItem(ctx context.Context, item *Item) error {
	_, span := otel.Tracer("go-echo").Start(ctx, "validateItem")
	defer span.End()

	if item.Name == "" {
		return fmt.Errorf("name is required")
	}
	if item.Value == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func enrichItem(ctx context.Context, item *Item) {
	_, span := otel.Tracer("go-echo").Start(ctx, "enrichItem")
	defer span.End()

	item.Enriched = true
}

func processItem(ctx context.Context, item *Item) error {
	ctx, span := otel.Tracer("go-echo").Start(ctx, "processItem")
	defer span.End()

	if err := validateItem(ctx, item); err != nil {
		return err
	}
	enrichItem(ctx, item)
	return nil
}

func httpMetricsMiddleware() echo.MiddlewareFunc {
	meter := otel.Meter("go-echo")
	requestDuration, _ := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of HTTP server requests"),
	)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start).Seconds()
			requestDuration.Record(c.Request().Context(), duration,
				metric.WithAttributes(
					attribute.String("http.request.method", c.Request().Method),
					attribute.Int("http.response.status_code", c.Response().Status),
					attribute.String("http.route", c.Path()),
					attribute.String("url.scheme", c.Scheme()),
				),
			)
			return err
		}
	}
}

func main() {
	ctx := context.Background()
	tp, err := initTracer(ctx)
	if err != nil {
		panic(err)
	}
	defer tp.Shutdown(ctx)

	mp, err := initMeter(ctx)
	if err != nil {
		panic(err)
	}
	defer mp.Shutdown(ctx)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware("go-echo"))
	e.Use(httpMetricsMiddleware())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "service": "go-echo"})
	})

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/items", func(c echo.Context) error {
		item := new(Item)
		if err := c.Bind(item); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if err := processItem(c.Request().Context(), item); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusCreated, item)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
