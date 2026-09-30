package model

import "time"

// ApiToken is a personal access token. Only the SHA-256 of the token is
// stored (Hash); the token itself is shown once, when it is created.
type ApiToken struct {
	ID      string `json:"id"`
	Account string `json:"account"`
	Name    string `json:"name"`
	// the first characters of the token, to recognize it in a list
	Prefix string `json:"prefix"`
	Hash   string `json:"hash"`

	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
}

// ApiTokenInfo is the public view of an ApiToken (no hash).
type ApiTokenInfo struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"createdAt"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
}

func NewApiTokenInfo(token *ApiToken) ApiTokenInfo {
	info := ApiTokenInfo{
		ID:        token.ID,
		Name:      token.Name,
		Prefix:    token.Prefix,
		CreatedAt: token.CreatedAt,
		ExpiresAt: token.ExpiresAt,
	}
	if !token.LastUsedAt.IsZero() {
		lastUsed := token.LastUsedAt
		info.LastUsedAt = &lastUsed
	}
	return info
}

// RequestCreateApiToken names a new token. ExpiresInDays is one of
// API_TOKEN_EXPIRY_DAYS, and API_TOKEN_DEFAULT_DAYS when zero.
type RequestCreateApiToken struct {
	Name          string `json:"name" binding:"required"`
	ExpiresInDays int    `json:"expiresInDays"`
}

// ResponseCreateApiToken carries the token itself, the only time it is shown.
type ResponseCreateApiToken struct {
	Message  string        `json:"message"`
	Token    string        `json:"token,omitempty"`
	ApiToken *ApiTokenInfo `json:"apiToken,omitempty"`
}

type ResponseApiTokens struct {
	Message string         `json:"message"`
	Tokens  []ApiTokenInfo `json:"tokens"`
}

type ResponseDeleteApiToken struct {
	Message string `json:"message"`
}
