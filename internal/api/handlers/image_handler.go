package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dylan0804/image-processing-tool/internal/api/imaging"
	"github.com/dylan0804/image-processing-tool/internal/api/interfaces"
	"github.com/dylan0804/image-processing-tool/internal/api/logger"
	"github.com/dylan0804/image-processing-tool/internal/api/response"
	"github.com/dylan0804/image-processing-tool/internal/api/storage"
	"github.com/dylan0804/image-processing-tool/internal/models/request"
	"github.com/google/uuid"
	"go.uber.org/zap"
)
type ImageHandler struct {
	response *response.Response
	sessionStore storage.RedisSessionStore
	imaging imaging.Imaging
}

func NewImageHandler(response *response.Response, sessionStore storage.RedisSessionStore, imaging imaging.Imaging) *ImageHandler {
	return &ImageHandler{
		response: response,
		sessionStore: sessionStore,
		imaging: imaging,
	}
}

func (i *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	logger := logger.LoggerFromContext(r.Context())

	err := r.ParseMultipartForm(10 >> 20)
	if err != nil {
		logger.Error("Failed to parse form", zap.Error(err))
		i.response.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		logger.Error("Failed to get file", zap.Error(err))
		i.response.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	sessionID := uuid.New().String()

	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, sessionID+filepath.Ext(header.Filename))

	err = i.sessionStore.Set(r.Context(), sessionID, interfaces.SessionData{
		OriginalFilename: header.Filename,
		TempPath: tempPath,
		UploadTime: time.Now(),
	})
	if err != nil {
		logger.Error("Failed to store metadata to redis", zap.Error(err))
		i.response.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tempFile, err := os.Create(tempPath)
	if err != nil {
        logger.Error("Failed to create temp file", zap.Error(err))
        i.response.WriteError(w, "Server error", http.StatusInternalServerError)
        return
    }
    defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
    if err != nil {
        logger.Error("Failed to save file", zap.Error(err))
        i.response.WriteError(w, "Server error", http.StatusInternalServerError)
        return
    }

	logger.Info("Image stored at", zap.String("temp path", tempPath))

	i.response.WriteSuccess(w, &response.BaseResponse{
		Success: true,
		Data: map[string]interface{}{
			"sessionId": sessionID,
		},
		Err: nil,
	})
}

func (i *ImageHandler) BlurImage(w http.ResponseWriter, r *http.Request) {
	logger := logger.LoggerFromContext(r.Context())

	var blurImageRequest request.BlurImageRequest

	if err := json.NewDecoder(r.Body).Decode(&blurImageRequest); err != nil {
		logger.Error("Failed to decode request body", zap.Error(err))
		i.response.WriteError(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	session, ok, err := i.sessionStore.Get(r.Context(), blurImageRequest.SessionID)
	if err != nil || !ok{
		logger.Error("Failed to get current image", zap.Error(err))
		i.response.WriteError(w, "Failed to get current image", http.StatusInternalServerError)
		return
	}

	img, err := i.imaging.Open(session.TempPath)
	if err != nil {
		logger.Error("Failed to open image", zap.Error(err))
		i.response.WriteError(w, "Failed to open image", http.StatusInternalServerError)
		return
	}

	sigmaInt, err := strconv.Atoi(blurImageRequest.Sigma)
	if err != nil {
		logger.Error("Failed to convert sigma to int", zap.Error(err))
		i.response.WriteError(w, "Failed to convert sigma to int", http.StatusInternalServerError)
		return
	}	

	blurredImg := i.imaging.Blur(img, float64(sigmaInt))

	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, uuid.New().String()+filepath.Ext(session.OriginalFilename))

	err = i.imaging.Save(blurredImg, tempPath)
	if err != nil {
		logger.Error("Failed to save blurred image", zap.Error(err))
		i.response.WriteError(w, "Failed to save blurred image", http.StatusInternalServerError)
		return
	}

	oldTempPath := session.TempPath
	session.TempPath = tempPath

	err = i.sessionStore.Set(r.Context(), blurImageRequest.SessionID, session)
	if err != nil {
		logger.Error("Failed to update session", zap.Error(err))
		i.response.WriteError(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	// clean up
	if oldTempPath != "" && tempPath != oldTempPath {
		os.Remove(oldTempPath)
	}

	i.response.WriteSuccess(w, &response.BaseResponse{
		Success: true,
		Data: map[string]interface{}{
			"sessionId": blurImageRequest.SessionID,
			"path": tempPath,
			"operation": "blur",
			"sigma": blurImageRequest.Sigma,
		},
		Err: nil,
	})
}

func (i *ImageHandler) SharpenImage(w http.ResponseWriter, r *http.Request) {
	logger := logger.LoggerFromContext(r.Context())

	var req request.SharpenImageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", zap.Error(err))
		i.response.WriteError(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	// check from redis
	session, exists, err := i.sessionStore.Get(r.Context(), req.SessionID)
	if err != nil || !exists {
		logger.Error("Failed to get session ID", zap.Error(err))
		i.response.WriteError(w, "Failed to get session ID", http.StatusBadRequest)
		return
	}

	// if ok
	// get image
	img, err := i.imaging.Open(session.TempPath)
	if err != nil {
		logger.Error("Failed to open image", zap.Error(err))
		i.response.WriteError(w, "Failed to open image", http.StatusBadRequest)
		return
	}

	// convert sigma to int
	sigmaInt, err := strconv.Atoi(req.Sigma)
	if err != nil {
		logger.Error("Failed to convert sigma to int", zap.Error(err))
		i.response.WriteError(w, "Failed to convert sigma to int", http.StatusInternalServerError)
		return
	}	

	// sharpen image
	sharpenedImg := i.imaging.Sharpen(img, float64(sigmaInt))

	// create a new tmp file
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, uuid.NewString()+filepath.Ext(session.OriginalFilename))

	// save file
	err = i.imaging.Save(sharpenedImg, tempPath)
	if err != nil {
		logger.Error("Failed to save sharpened image", zap.Error(err))
		i.response.WriteError(w, "Failed to save sharpened image", http.StatusInternalServerError)
		return
	}

	// delete the old one
	oldTempPath := session.TempPath
	session.TempPath = tempPath

	err = i.sessionStore.Set(r.Context(), req.SessionID, session)
	if err != nil {
		logger.Error("Failed to update session", zap.Error(err))
		i.response.WriteError(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	if oldTempPath != "" && oldTempPath != tempPath {
		os.Remove(oldTempPath)
	}

	// return
	i.response.WriteSuccess(w, &response.BaseResponse{
		Success: true,
		Data: map[string]interface{}{
			"sessionId": req.SessionID,
			"path": tempPath,
			"operation": "sharpen",
			"sigma": req.Sigma,
		},
	})
}