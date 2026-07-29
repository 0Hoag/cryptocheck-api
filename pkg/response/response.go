package response

import (
	"fmt"
	"net/http"
	"runtime"

	pkgErrors "github.com/0Hoag/cryptocheck-api/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Resp is the response format.
type Resp struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Errors    any    `json:"errors,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// NewOKResp returns a new OK response with the given data.
func NewOKResp(data any) Resp {
	return Resp{
		ErrorCode: 0,
		Message:   "Success",
		Data:      data,
	}
}

func withRequestID(c *gin.Context, resp Resp) Resp {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		resp.RequestID = requestID
	}
	return resp
}

// Ok returns a new OK response with the given data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, withRequestID(c, NewOKResp(data)))
}

// Unauthorized returns a new Unauthorized response with the given data.
func Unauthorized(c *gin.Context) {
	status, resp := parseError(pkgErrors.NewUnauthorizedHTTPError(), c)
	c.JSON(status, withRequestID(c, resp))
}

func Forbidden(c *gin.Context) {
	status, resp := parseError(pkgErrors.NewForbiddenHTTPError(), c)
	c.JSON(status, withRequestID(c, resp))
}

func parseError(err error, c *gin.Context) (int, Resp) {
	//print error . type
	fmt.Printf("Error: %T, %v\n", err, err)
	switch parsedErr := err.(type) {
	case *pkgErrors.ValidationError:
		return http.StatusBadRequest, Resp{
			ErrorCode: parsedErr.Code,
			Message:   parsedErr.Error(),
		}
	case *pkgErrors.PermissionError:
		return http.StatusBadRequest, Resp{
			ErrorCode: parsedErr.Code,
			Message:   parsedErr.Error(),
		}
	case *pkgErrors.ValidationErrorCollector:
		return http.StatusBadRequest, Resp{
			ErrorCode: ValidationErrorCode,
			Message:   ValidationErrorMsg,
			Errors:    parsedErr.Errors(),
		}
	case *pkgErrors.PermissionErrorCollector:
		return http.StatusBadRequest, Resp{
			ErrorCode: PermissionErrorCode,
			Message:   PermissionErrorMsg,
			Errors:    parsedErr.Errors(),
		}
	case *pkgErrors.HTTPError:
		statusCode := parsedErr.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusBadRequest
		}

		return statusCode, Resp{
			ErrorCode: parsedErr.Code,
			Message:   parsedErr.Message,
		}
	default:
		return http.StatusInternalServerError, Resp{
			ErrorCode: 500,
			Message:   DefaultErrorMessage,
		}
	}
}

// Error returns a new Error response with the given error.
func Error(c *gin.Context, err error) {
	status, resp := parseError(err, c)
	c.JSON(status, withRequestID(c, resp))
}

// HttpError returns a new Error response with the given error.
func HttpError(c *gin.Context, err *pkgErrors.HTTPError) {
	status, resp := parseError(err, c)
	c.JSON(status, withRequestID(c, resp))
}

// ErrorMapping is a map of error to HTTPError.
type ErrorMapping map[error]*pkgErrors.HTTPError

// ErrorWithMap returns a new Error response with the given error.
func ErrorWithMap(c *gin.Context, err error, eMap ErrorMapping) {
	if httpErr, ok := eMap[err]; ok {
		Error(c, httpErr)
		return
	}

	Error(c, err)
}

func PanicError(c *gin.Context, err any) {
	if err == nil {
		c.JSON(parseError(nil, c))
	} else {
		c.JSON(parseError(err.(error), c))
	}
}

func captureStackTrace() []string {
	var pcs [defaultStackTraceDepth]uintptr
	n := runtime.Callers(2, pcs[:])
	if n == 0 {
		return nil
	}

	var stackTrace []string
	for _, pc := range pcs[:n] {
		f := runtime.FuncForPC(pc)
		if f != nil {
			file, line := f.FileLine(pc)
			stackTrace = append(stackTrace, fmt.Sprintf("%s:%d %s", file, line, f.Name()))
		}
	}

	return stackTrace
}

func parseErrorWithData(err error, c *gin.Context, data any) (int, Resp) {
	switch parsedErr := err.(type) {
	case *pkgErrors.ValidationError:
		return http.StatusBadRequest, Resp{
			ErrorCode: parsedErr.Code,
			Message:   parsedErr.Error(),
			Data:      data,
		}
	case *pkgErrors.PermissionError:
		return http.StatusBadRequest, Resp{
			ErrorCode: parsedErr.Code,
			Message:   parsedErr.Error(),
			Data:      data,
		}
	case *pkgErrors.HTTPError:
		statusCode := parsedErr.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusBadRequest
		}
		return statusCode, Resp{
			ErrorCode: parsedErr.Code,
			Message:   parsedErr.Message,
			Data:      data,
		}
	default:
		return http.StatusBadRequest, Resp{
			ErrorCode: http.StatusBadRequest,
			Message:   err.Error(),
			Data:      data,
		}
	}
}

func ErrorWithData(c *gin.Context, err error, data any) {
	status, resp := parseErrorWithData(err, c, data)
	c.JSON(status, withRequestID(c, resp))
}

// StatusError returns a controlled HTTP error while preserving the shared API
// envelope and request correlation ID.
func StatusError(c *gin.Context, status, errorCode int, message string) {
	c.JSON(status, withRequestID(c, Resp{ErrorCode: errorCode, Message: message}))
}

func StatusErrorWithData(c *gin.Context, status, errorCode int, message string, data any) {
	c.JSON(status, withRequestID(c, Resp{ErrorCode: errorCode, Message: message, Data: data}))
}
