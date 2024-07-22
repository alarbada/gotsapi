package pkg

import "github.com/labstack/echo/v4"

type SomeHandlerReq struct{}
type SomeHandlerRes struct{}

func SomeHandler(e echo.Context, params SomeHandlerReq) (*SomeHandlerRes, error) {
	return nil, nil
}
