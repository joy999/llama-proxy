package middleware

import (
	"net/http"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   interface{} `json:"param"`
	Code    string      `json:"code"`
}

func ResponseHandler(r *ghttp.Request) {
	r.Middleware.Next()

	contentType := r.Response.Header().Get("Content-Type")
	if contentType == "" {
		r.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	if err := r.GetError(); err != nil {
		handleError(r, err)
	} else {
		handleSuccess(r)
	}
}

func handleError(r *ghttp.Request, err error) {
	errorDetail := ErrorDetail{
		Message: err.Error(),
		Type:    "server_error",
		Param:   nil,
		Code:    "internal_error",
	}

	httpStatus := http.StatusInternalServerError

	if gerror.HasStack(err) {
		code := gerror.Code(err)
		if code != nil {
			switch code.Code() {
			case http.StatusBadRequest:
				httpStatus = http.StatusBadRequest
				errorDetail.Type = "invalid_request_error"
				errorDetail.Code = "bad_request"
			case http.StatusNotFound:
				httpStatus = http.StatusNotFound
				errorDetail.Type = "invalid_request_error"
				errorDetail.Code = "not_found"
			case http.StatusUnauthorized:
				httpStatus = http.StatusUnauthorized
				errorDetail.Type = "authentication_error"
				errorDetail.Code = "unauthorized"
			case http.StatusForbidden:
				httpStatus = http.StatusForbidden
				errorDetail.Type = "permission_error"
				errorDetail.Code = "forbidden"
			}
		}
	}

	switch err.Error() {
	case "model parameter is required":
		httpStatus = http.StatusBadRequest
		errorDetail.Type = "invalid_request_error"
		errorDetail.Code = "missing_model"
	case "model not found":
		httpStatus = http.StatusNotFound
		errorDetail.Type = "invalid_request_error"
		errorDetail.Code = "model_not_found"
	case "model is not loaded":
		httpStatus = http.StatusBadRequest
		errorDetail.Type = "invalid_request_error"
		errorDetail.Code = "model_not_loaded"
	}

	r.Response.Status = httpStatus
	r.Response.ClearBuffer()
	r.Response.WriteJson(ErrorResponse{Error: errorDetail})
}

func handleSuccess(r *ghttp.Request) {
	if r.Response.Status == 0 {
		r.Response.Status = http.StatusOK
	}

	if r.Response.Header().Get("Content-Type") == "" {
		r.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	if r.Response.Buffer() == nil || len(r.Response.Buffer()) == 0 {
		if handlerRes := r.GetHandlerResponse(); handlerRes != nil {
			r.Response.ClearBuffer()
			r.Response.WriteJson(handlerRes)
		}
	}
}

func WriteError(r *ghttp.Request, status int, message string, errorType string, code string) {
	r.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.Response.WriteStatus(status)
	r.Response.WriteJson(g.Map{
		"error": g.Map{
			"message": message,
			"type":    errorType,
			"param":   nil,
			"code":    code,
		},
	})
}
