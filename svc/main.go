package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/head", func(c echo.Context) error {
		for k := range c.Request().Header {
			fmt.Println(k, ":", c.Request().Header.Get(k))
		}
		return c.String(http.StatusOK, "OK")
	})
	e.Logger.Fatal(e.Start(":8080"))
}
