//go:generate easyjson -all models.go
package webserver

import (
	"errors"
	"math"
)

var ErrInternal = errors.New("internal error")

type Response struct {
	Data       interface{}    `json:"data,omitempty"`
	Error      *ErrorResponse `json:"error,omitempty"`
	Pagination *Pagination    `json:"pagination,omitempty"`
	Ok         bool           `json:"ok"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type Pagination struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	PerPage int `json:"per_page"`
}

func GetSuccessResponse(data interface{}) *Response {
	return &Response{
		Ok:   true,
		Data: data,
	}
}

func GetResponseWithError(err error, code int) *Response {
	if err == nil {
		err = ErrInternal
	}

	return &Response{
		Ok: false,
		Error: &ErrorResponse{
			Message: err.Error(),
			Code:    code,
		},
	}
}

func GetSuccessResponseList(data interface{}, total, page, perPage int) *Response {
	var pages int
	if perPage == 0 || total == 0 {
		pages = 0
	} else {
		pages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	return &Response{
		Ok:   true,
		Data: data,
		Pagination: &Pagination{
			Total:   total,
			Page:    page,
			Pages:   pages,
			PerPage: perPage,
		},
	}
}
