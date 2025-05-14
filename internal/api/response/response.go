package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {}

func NewResponse() *Response {
	return &Response{}
}

type BaseResponse struct {
	Err error `json:"error"`
	Message any `json:"message"`
	Success bool `json:"success"`
}

func (r *Response) WriteSuccess(w http.ResponseWriter, data *BaseResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(data)
}

func (r *Response) WriteError(w http.ResponseWriter, errMsg string, code int) {
	http.Error(w, errMsg, code)
}