package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/dylan0804/image-processing-tool/internal/api/interfaces"
	"github.com/dylan0804/image-processing-tool/internal/api/response"
	"github.com/dylan0804/image-processing-tool/internal/models/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ====
type mockSessionStore struct {
	store map[string]interfaces.SessionData
}

func newMockSessionStore() *mockSessionStore {
	return &mockSessionStore{
		store: make(map[string]interfaces.SessionData),
	}
}

func (m *mockSessionStore) Set(ctx context.Context, sessionID string, data interfaces.SessionData) error {
	m.store[sessionID] = data
	return nil
}

func (m *mockSessionStore) Get(ctx context.Context, sessionID string) (interfaces.SessionData, bool, error) {
	data, exists := m.store[sessionID]
	return data, exists, nil
}

func (m *mockSessionStore) Delete(ctx context.Context, sessionID string) error {
	delete(m.store, sessionID)
	return nil
}
// =====

type mockImaging struct {
	openError error
	src image.Image
}

func newMockImaging() *mockImaging {
	return &mockImaging{
		src: image.NewRGBA(image.Rect(1, 2, 3, 4)),
	}
}

func (m *mockImaging) Open(path string) (image.Image, error) {
	if m.openError != nil {
		return nil, m.openError
	}
	return m.src, nil
}
func (m *mockImaging) Blur(img image.Image, sigma float64) *image.NRGBA {
	return imaging.Blur(m.src, sigma)
}
func (m *mockImaging) Save(img image.Image, path string) error {
	return nil
}
func (m *mockImaging) Sharpen(image image.Image, sigma float64) image.Image {
	return m.src
}
// =========

func TestImageHandler_UploadImage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "image-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testImagePath := filepath.Join(tempDir, "test-image.jpg")
	createTestImage(t, testImagePath)
	defer os.Remove(testImagePath)

	mockStore := newMockSessionStore()
	respHelper := response.NewResponse()

	handler := NewImageHandler(respHelper, mockStore, nil)

	tests := []struct{
		name string
		setupRequest func() (*http.Request, error)
		checkResponse func(*httptest.ResponseRecorder)
	}{
		{
			name: "Successful upload",
			setupRequest: func() (*http.Request, error) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				part, err := writer.CreateFormFile("image", "test-image.jpg")
				if err != nil {
					return nil, err
				}

				file, err := os.Open(testImagePath)
				if err != nil {
					return nil, err
				}
				defer file.Close()

				_, err = io.Copy(part, file)
				if err != nil {
					return nil, err
				}

				writer.Close()

				req := httptest.NewRequest("POST", "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())

				return req, nil
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				var resp response.BaseResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)

				assert.True(t, resp.Success)

				msg, ok := resp.Data.(map[string]interface{})
				assert.True(t, ok)

				sessionID, ok := msg["sessionId"].(string)
				assert.True(t, ok)
				assert.NotEmpty(t, sessionID)

				_, exists, err := mockStore.Get(context.Background(), sessionID)
				assert.NoError(t, err)
				assert.True(t, exists)
			},			
		},
		{
			name: "Missing image file",
			setupRequest: func() (*http.Request, error) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()

				req := httptest.NewRequest("POST", "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				var resp response.BaseResponse

				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)

				assert.False(t, resp.Success)				
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := tc.setupRequest()
			require.NoError(t, err)

			rec := httptest.NewRecorder()

			ctx := req.Context()

			req = req.WithContext(ctx)

			handler.UploadImage(rec, req)

			tc.checkResponse(rec)
		})
	}
}

func TestImageHandler_BlurImage(t *testing.T) {
	mockStore := newMockSessionStore()
	mockStore.Set(context.Background(), "image-sessionId", interfaces.SessionData{
		TempPath: "/path/to/temp",
	})
	
	respHelper := response.NewResponse()
	mockImaging := newMockImaging()

	handler := NewImageHandler(respHelper, mockStore, mockImaging)

	testcases := []struct{
		name string
		setupRequest func() (*http.Request, error)
		checkResponse func(*httptest.ResponseRecorder)
	}{
		{
			name: "Successfully blurred image",
			setupRequest: func() (*http.Request, error) {
				blurImageRequest := request.BlurImageRequest{
					SessionID: "image-sessionId",
					Sigma: "10",
				}

				jsonBytes, err := json.Marshal(&blurImageRequest)
				require.NoError(t, err)

				req := httptest.NewRequest("POST", "/blur", bytes.NewBuffer(jsonBytes))
				req.Header.Set("Content-Type", "application/json")

				return req, nil
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				var resp response.BaseResponse

				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)

				assert.True(t, resp.Success)

				data, ok := resp.Data.(map[string]interface{})
				assert.True(t, ok)

				sessionId, ok := data["sessionId"].(string)
				assert.True(t, ok)
				assert.NotEmpty(t, sessionId)

				_, exists, err := mockStore.Get(context.Background(), "image-sessionId")
				assert.NoError(t, err)
				assert.True(t, exists)
			},		
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := tc.setupRequest()
			require.NoError(t, err)

			rec := httptest.NewRecorder()

			ctx := req.Context()

			req = req.WithContext(ctx)

			handler.BlurImage(rec, req)

			tc.checkResponse(rec)
		})
	}
}

func TestImageHandler_SharpenImage(t *testing.T) {
	mockStore := newMockSessionStore()
	respHelper := response.NewResponse()
	mockImaging := newMockImaging()

	handler := NewImageHandler(respHelper, mockStore, mockImaging)

	testcases := []struct{
		name string
		setupRequest func() (*http.Request, error)
		checkResponse func(*httptest.ResponseRecorder)
	}{
		{
			name: "Successfully sharpen image",
			setupRequest: func() (*http.Request, error) {
				// setup json body
				sharpenImgReq := &request.SharpenImageRequest{
					SessionID: "session-imageId",
					Sigma: "42",
				}

				jsonBytes, err := json.Marshal(sharpenImgReq)
				assert.NoError(t, err)

				req, err := http.NewRequest("POST", "/sharpen", bytes.NewBuffer(jsonBytes))
				assert.NoError(t, err)	
			
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				var response response.BaseResponse

				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)

				data, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)

				sessionId, ok := data["sessionId"].(string)
				assert.True(t, ok)
				assert.NotEmpty(t, sessionId)

				path, ok := data["path"].(string)
				assert.True(t, ok)
				assert.NotEmpty(t, path)

				operation, ok := data["operation"].(string)
				assert.True(t, ok)
				assert.Equal(t, "sharpen", operation)

				_, exists, err := mockStore.Get(context.Background(), sessionId)
				assert.NoError(t, err)
				assert.True(t, exists)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := tc.setupRequest()
			require.NoError(t, err)

			ctx := req.Context()
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()

			mockStore.Set(req.Context(), "session-imageId", interfaces.SessionData{
				TempPath: "path/to/temp",
			})

			handler.SharpenImage(rec, req)

			tc.checkResponse(rec)
		})
	}
}

func createTestImage(t *testing.T, path string) {
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	img := image.NewRGBA(image.Rect(1, 2, 3, 4))

	err = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
	require.NoError(t, err)
}