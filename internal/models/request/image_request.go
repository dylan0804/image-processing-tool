package request

type BlurImageRequest struct {
	SessionID string `json:"sessionID"`
	Sigma string `json:"sigma"`
}

type SharpenImageRequest struct {
	SessionID string `json:"sessionID"`
	Sigma string `json:"sigma"`
}