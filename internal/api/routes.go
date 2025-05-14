package api

import (
	"net/http"

	"github.com/dylan0804/image-processing-tool/internal/api/handlers"
	"go.uber.org/zap"
)

type Route struct {
	mux  *http.ServeMux
	imageHandler *handlers.ImageHandler
	logger *zap.Logger
}

func NewRoutes(mux *http.ServeMux, i *handlers.ImageHandler, logger *zap.Logger) *Route {
	return &Route{
		mux: mux,
		imageHandler: i,
		logger: logger,
	}
}

func (r *Route) InitRoutes() {
	r.mux.HandleFunc("POST /api/v1/image/upload", r.imageHandler.UploadImage)

	r.logger.Info("app running on port :8080")

	http.ListenAndServe(":8080", r.mux)
}