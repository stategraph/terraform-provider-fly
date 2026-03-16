package apimodels

type OIDCTokenRequest struct {
	Aud string `json:"aud"`
}

type OIDCTokenResponse struct {
	Token string `json:"token"`
}
