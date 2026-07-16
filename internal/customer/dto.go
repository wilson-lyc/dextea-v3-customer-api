package customer

type LoginRequest struct {
	Code     string   `json:"code" binding:"required"`
	Platform Platform `json:"platform" binding:"required"`
}

type LoginResponse struct {
	Token    string    `json:"token"`
	Customer *Customer `json:"customer"`
}
