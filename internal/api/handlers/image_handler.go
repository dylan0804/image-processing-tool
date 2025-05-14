package handlers

import (
	"io"
	"net/http"
	"os"

	"github.com/dylan0804/image-processing-tool/internal/api/response"
	"go.uber.org/zap"
)

type BaseResponse struct {
	Err error `json:"error"`
	Message any `json:"message"`
	Success bool `json:"success"`
}

type ImageHandler struct {
	response *response.Response
	logger *zap.Logger
}

func NewImageHandler(response *response.Response, logger *zap.Logger) *ImageHandler {
	return &ImageHandler{
		response: response,
		logger: logger, 
	}
}

func (i *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 >> 20)
	if err != nil {
		i.logger.Error("Failed to parse form", zap.Error(err))
		i.response.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		i.logger.Error("Failed to get file", zap.Error(err))
		i.response.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dst, err := os.Create(header.Filename)
	if err != nil {
		i.response.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = io.Copy(dst, file)
	if err != nil {
		i.response.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	i.response.WriteSuccess(w, &response.BaseResponse{
		Success: true,
		Message: "Upload success",
		Err: nil,
	})
}