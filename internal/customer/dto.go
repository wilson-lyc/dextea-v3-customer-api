package customer

type LoginRequest struct {
	Code     string   `json:"code" binding:"required"`
	Platform Platform `json:"platform" binding:"required"`
}

type LoginResponse struct {
	Code string `json:"code"`
}
