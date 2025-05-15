package main

import (
	"log"
	"net/http"

	"github.com/dylan0804/image-processing-tool/internal/api"
	"github.com/dylan0804/image-processing-tool/internal/api/handlers"
	"github.com/dylan0804/image-processing-tool/internal/api/imaging"
	"github.com/dylan0804/image-processing-tool/internal/api/logger"
	"github.com/dylan0804/image-processing-tool/internal/api/response"
	"github.com/dylan0804/image-processing-tool/internal/api/storage"
)

func main() {
	mux := http.NewServeMux()
	logger, err := logger.InitLogger()
	if err != nil {
		panic("Failed to init logger: " + err.Error())
	}
	defer logger.Sync()
	// set up redis
	sessionStore, err := storage.NewRedisSessionStore()
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	// set up response
	response := response.NewResponse()

	// set up imaging
	imaging := imaging.NewImaging()

	// set up handlers
	imageHandler := handlers.NewImageHandler(response, sessionStore, imaging)

	routes := api.NewRoutes(mux, imageHandler, logger)

	routes.InitRoutes()
}