package gotsclient

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
)

type Handler[P any, R any] func(c echo.Context, params P) (R, error)

type TypedHandlers struct {
	e        *echo.Echo
	handlers map[string]reflect.Type // "package.handler" -> Handler type
}

func NewTypedHandlers(e *echo.Echo) *TypedHandlers {
	return &TypedHandlers{
		e:        e,
		handlers: make(map[string]reflect.Type),
	}
}

func AddHandler[P any, R any](th *TypedHandlers, handler Handler[P, R]) {
	handlerFunc := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())
	fullName := handlerFunc.Name()

	parts := strings.Split(fullName, ".")

	// Ensure we have at least two parts (package and function name)
	if len(parts) < 2 {
		panic("Invalid function name format")
	}

	packageName := parts[len(parts)-2]
	{
		parts := strings.Split(packageName, "/")
		if len(parts) > 0 {
			packageName = parts[len(parts)-1]
		}
	}

	handlerName := parts[len(parts)-1]

	path := fmt.Sprintf("/%s.%s", packageName, handlerName)
	fullPath := fmt.Sprintf("%s.%s", packageName, handlerName)

	th.handlers[fullPath] = reflect.TypeOf(handler)
	fmt.Println(path)

	th.e.POST(path, func(c echo.Context) error {
		var params P
		if err := c.Bind(&params); err != nil {
			return echo.NewHTTPError(400, err.Error())
		}

		result, err := handler(c, params)
		if err != nil {
			return c.JSON(400, map[string]string{
				"message": err.Error(),
			})
		}

		return c.JSON(200, result)
	})
}

func (th *TypedHandlers) GenerateTypescriptClient() string {
	var sb strings.Builder

	// Generate ApiError type and ApiResponse type
	sb.WriteString(`
export interface ApiError {
  message: string
  statusCode?: number
}
export type ApiResponse<T> =
  | { data: T; error: null }
  | { data: null; error: ApiError }
`)

	// Generate ApiClient interface
	sb.WriteString("export interface ApiClient {\n")
	packages := make(map[string]map[string]reflect.Type)
	for fullPath, handlerType := range th.handlers {
		parts := strings.Split(fullPath, ".")
		if len(parts) != 2 {
			continue
		}
		packageName, handlerName := parts[0], parts[1]
		packageParts := strings.Split(packageName, "/")
		simplifiedPackageName := packageParts[len(packageParts)-1]
		if packages[simplifiedPackageName] == nil {
			packages[simplifiedPackageName] = make(map[string]reflect.Type)
		}
		packages[simplifiedPackageName][handlerName] = handlerType
	}

	for packageName, handlers := range packages {
		sb.WriteString(fmt.Sprintf("  %s: {\n", packageName))
		for handlerName, handlerType := range handlers {
			if handlerType.NumIn() < 2 || handlerType.NumOut() < 1 {
				fmt.Printf("Warning: unexpected handler signature for %s.%s\n", packageName, handlerName)
				continue
			}
			paramsType := handlerType.In(1)
			returnType := handlerType.Out(0)
			sb.WriteString(fmt.Sprintf("    %s: (params: %s) => Promise<ApiResponse<%s>>\n",
				handlerName,
				generateTypescriptType(paramsType),
				generateTypescriptType(returnType)))
		}
		sb.WriteString("  }\n")
	}
	// Add beforeRequest method to ApiClient interface
	sb.WriteString("  beforeRequest(hook: (config: RequestInit) => void): void\n")
	sb.WriteString("}\n")

	// Generate createApiClient function
	sb.WriteString(`
export const createApiClient = (baseUrl: string): ApiClient => {
  let beforeRequestHook: ((config: RequestInit) => void) | null = null

  async function doFetch(path: string, params: unknown) {
    try {
      const config: RequestInit = {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(params),
      }

      if (beforeRequestHook) {
        beforeRequestHook(config)
      }

      const response = await fetch(` + "`${baseUrl}/${path}`" + `, config)
      if (!response.ok) {
        return {
          data: null,
          error: {
            message: "API request failed",
            statusCode: response.status,
          },
        }
      }
      const data = await response.json()
      return { data, error: null }
    } catch (error) {
      return {
        data: null,
        error: {
          message:
            error instanceof Error ? error.message : "Unknown error occurred",
        },
      }
    }
  }
  const client: ApiClient = {
`)

	// Generate client methods
	for packageName, handlers := range packages {
		sb.WriteString(fmt.Sprintf("    %s: {\n", packageName))
		for handlerName := range handlers {
			sb.WriteString(fmt.Sprintf("      %s: (params) => doFetch(\"%s.%s\", params),\n", handlerName, packageName, handlerName))
		}
		sb.WriteString("    },\n")
	}

	// Add beforeRequest method implementation
	sb.WriteString(`    beforeRequest: (hook) => {
      beforeRequestHook = hook
    },
`)

	sb.WriteString(`  }
  return client
}
`)

	return sb.String()
}

func generateTypescriptType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Struct:
		var sb strings.Builder
		sb.WriteString("{\n")
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldName := field.Tag.Get("json")
			if fieldName == "" {
				fieldName = field.Name
			}
			sb.WriteString(fmt.Sprintf("    %s: %s\n", fieldName, generateTypescriptType(field.Type)))
		}
		sb.WriteString("  }")
		return sb.String()
	case reflect.Slice, reflect.Array:
		return generateTypescriptType(t.Elem()) + "[]"
	case reflect.Ptr:
		return generateTypescriptType(t.Elem())
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Interface:
		return "any"
	default:
		return "unknown"
	}
}

func WriteToFile(th *TypedHandlers, filename string) error {
	generated := []byte(th.GenerateTypescriptClient())
	return os.WriteFile(filename, generated, 0644)
}
