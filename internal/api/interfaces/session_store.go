package interfaces

import (
	"time"
)

type SessionData struct {
	OriginalFilename string    `json:"originalFilename"`
    TempPath         string    `json:"tempPath"`
    UploadTime       time.Time `json:"uploadTime"`
}

type SessionStore interface {
	Set(sessionID string, data SessionData) error
    Get(sessionID string) (SessionData, bool, error)
    Delete(sessionID string) error
}