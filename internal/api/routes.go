package api

import (
	"net/http"

	"github.com/dylan0804/image-processing-tool/internal/api/handlers"
	"github.com/dylan0804/image-processing-tool/internal/api/middleware"
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
	r.mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.mux.HandleFunc("POST /api/v1/image/upload", r.imageHandler.UploadImage)
	r.mux.HandleFunc("POST /api/v1/image/resize", r.imageHandler.BlurImage)
	r.mux.HandleFunc("POST /api/v1/image/sharpen", r.imageHandler.SharpenImage)

	handler := middleware.LoggingMiddleware(r.logger, r.mux)

	r.logger.Info("app running on port :8080")

	http.ListenAndServe(":8080", handler)
}