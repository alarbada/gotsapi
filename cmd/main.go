package main

import (
	"gotsapi"
	"gotsapi/cmd/nested/pkg"

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

	return &ExampleResponse{Greeting: "Hello, " + params.Name}, nil
}

func ExampleHandler2(c echo.Context, params ExampleParams) (*ExampleResponse, error) {
	dump.P(params)

	return &ExampleResponse{Greeting: "Hello, " + params.Name}, nil
}

func HelloWorld(c echo.Context, params struct{}) (string, error) {
	return "hello world", nil
}


func main() {
	e := echo.New()
	th := gotsapi.NewTypedHandlers(e)

	gotsapi.AddHandler(th, ExampleHandler1)
	gotsapi.AddHandler(th, ExampleHandler2)
	gotsapi.AddHandler(th, HelloWorld)
	gotsapi.AddHandler(th, pkg.SomeHandler)

	gotsapi.WriteToFile(th, "scripts/apiclient.ts")

	e.Logger.Fatal(e.Start(":8080"))
}
