package main

import (
	"net/http"

	"github.com/dylan0804/image-processing-tool/internal/api"
	"github.com/dylan0804/image-processing-tool/internal/api/handlers"
	"github.com/dylan0804/image-processing-tool/internal/api/logger"
	"github.com/dylan0804/image-processing-tool/internal/api/response"
)

func main() {
	mux := http.NewServeMux()
	logger, err := logger.InitLogger()
	if err != nil {
		panic("Failed to init logger: " + err.Error())
	}
	defer logger.Sync()

	// set up response
	response := response.NewResponse()

	// set up handlers
	imageHandler := handlers.NewImageHandler(response, logger)

	routes := api.NewRoutes(mux, imageHandler, logger)

	routes.InitRoutes()
}