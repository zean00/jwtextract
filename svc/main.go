package main

import (
	"fmt"
	"net/http"
	"net/textproto"

	"github.com/labstack/echo/v4"
)

const subjectHeader = "subject"
const phoneHeader = "phone_number"

func main() {
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fmt.Println("DEBUG")
			fmt.Println(c.Request().Header.Get(subjectHeader))
			fmt.Println(c.Request().Header.Get(phoneHeader))
			//fmt.Println(c.Request().Header.Get(textproto.CanonicalMIMEHeaderKey(subjectHeader)))
			if c.Request().Header.Get(subjectHeader) == "" {
				c.Request().Header.Set(subjectHeader, c.Request().Header.Get(textproto.CanonicalMIMEHeaderKey(subjectHeader)))
			}

			if c.Request().Header.Get(phoneHeader) == "" {
				c.Request().Header.Set(phoneHeader, c.Request().Header.Get(textproto.CanonicalMIMEHeaderKey(phoneHeader)))
			}

			return next(c)
		}
	})
	e.GET("/head", func(c echo.Context) error {
		for k := range c.Request().Header {
			fmt.Println(k, ":", c.Request().Header.Get(k))
		}
		return c.String(http.StatusOK, "OK")
	})
	e.Logger.Fatal(e.Start(":8080"))
}
