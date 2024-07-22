package main

import (
	"fmt"
	"gotsapi"
	"os"
	"time"

	"github.com/gookit/goutil/dump"
	"github.com/labstack/echo/v4"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type ExampleParams struct {
	Name  string `json:"name"`
	Users []User `json:"users"`
}

type ExampleResponse struct {
	Greeting string `json:"greeting"`
}

func ExampleHandler1(c echo.Context, params ExampleParams) (*ExampleResponse, error) {
	dump.P(params)

	return nil, fmt.Errorf("shit happens")

	// return &ExampleResponse{Greeting: "Hello, " + params.Name}, nil
}

func ExampleHandler2(c echo.Context, params ExampleParams) (*ExampleResponse, error) {
	dump.P(params)

	return nil, fmt.Errorf("shit happens")

	// return &ExampleResponse{Greeting: "Hello, " + params.Name}, nil
}

func main() {
	e := echo.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		c.JSON(400, map[string]string{
			"error": err.Error(),
		})
	}
	th := gotsapi.NewTypedHandlers(e)

	gotsapi.AddHandler(th, ExampleHandler1)
	gotsapi.AddHandler(th, ExampleHandler2)

	now := time.Now()
	os.WriteFile("scripts/apiclient.ts", []byte(th.GenerateTypescriptClient()), 0644)
	fmt.Println(time.Since(now))

	e.Logger.Fatal(e.Start(":8080"))
}
