package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Item struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Enriched bool   `json:"enriched"`
}

func validateItem(item *Item) error {
	if item.Name == "" {
		return fmt.Errorf("name is required")
	}
	if item.Value == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func enrichItem(item *Item) {
	item.Enriched = true
}

func processItem(item *Item) error {
	if err := validateItem(item); err != nil {
		return err
	}
	enrichItem(item)
	return nil
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

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
		if err := processItem(item); err != nil {
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
