package entity

type PingRequest struct{}

type PingRequestQueue struct {
	PingRequest

	CallbackCh string `json:"callback_ch"`
}

type PingResponse struct {
	Message string `json:"message"`
}
