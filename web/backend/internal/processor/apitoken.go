package processor

import (
	"backend/constant"
	"backend/internal/context"
	"backend/model"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"
)

const (
	// random bytes in a token (256 bits)
	apiTokenRandomBytes = 32
	// characters of the token kept to recognize it in a list
	apiTokenPrefixLength = 10
	// lastUsedAt is written at most this often per token
	apiTokenTouchInterval = time.Minute
)

// CreateMyApiToken makes a personal access token for the caller. The token is
// returned here only; the database keeps its SHA-256.
func (p *Processor) CreateMyApiToken(acc *model.Account, req *model.RequestCreateApiToken) (*model.ResponseCreateApiToken, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing create API token for %s", acc.Account)

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errBadRequest("Name must not be blank")
	}
	days := req.ExpiresInDays
	if days == 0 {
		days = constant.API_TOKEN_DEFAULT_DAYS
	}
	if !slices.Contains(constant.API_TOKEN_EXPIRY_DAYS, days) {
		return nil, errBadRequest("Expiry must be 30, 60, 180 or 365 days")
	}

	existing, err := p.ListApiTokens(acc.Account)
	if err != nil {
		p.ProcLog.Errorf("Failed to list API tokens of %s: %v", acc.Account, err)
		return nil, errInternal("Failed to list API tokens")
	}
	if len(existing) >= constant.API_TOKEN_MAX_PER_ACCOUNT {
		return nil, &model.ErrorDetail{
			HttpStatus: http.StatusConflict,
			Detail:     "Too many API tokens; revoke one first",
		}
	}

	random := make([]byte, apiTokenRandomBytes)
	if _, err := rand.Read(random); err != nil {
		p.ProcLog.Errorf("Failed to generate an API token: %v", err)
		return nil, errInternal("Failed to generate an API token")
	}
	plain := constant.API_TOKEN_PREFIX + base64.RawURLEncoding.EncodeToString(random)

	now := p.now()
	token := &model.ApiToken{
		Account: acc.Account,
		Name:    name,
		Prefix:  plain[:apiTokenPrefixLength],
		Hash:    hashApiToken(plain),

		CreatedAt: now,
		ExpiresAt: now.AddDate(0, 0, days),
	}
	if err := p.CreateApiToken(token); err != nil {
		p.ProcLog.Errorf("Failed to store an API token of %s: %v", acc.Account, err)
		return nil, errInternal("Failed to create API token")
	}
	p.ProcLog.Infof("API token %s created for %s", token.ID, acc.Account)

	info := model.NewApiTokenInfo(token)
	return &model.ResponseCreateApiToken{
		Message:  "Create API token successful",
		Token:    plain,
		ApiToken: &info,
	}, nil
}

func (p *Processor) ListMyApiTokens(acc *model.Account) (*model.ResponseApiTokens, *model.ErrorDetail) {
	tokens, err := p.ListApiTokens(acc.Account)
	if err != nil {
		p.ProcLog.Errorf("Failed to list API tokens of %s: %v", acc.Account, err)
		return nil, errInternal("Failed to list API tokens")
	}

	infos := make([]model.ApiTokenInfo, 0, len(tokens))
	for _, token := range tokens {
		infos = append(infos, model.NewApiTokenInfo(token))
	}
	return &model.ResponseApiTokens{
		Message: "List API tokens successful",
		Tokens:  infos,
	}, nil
}

// DeleteMyApiToken revokes one of the caller's tokens; other people's IDs
// answer 404.
func (p *Processor) DeleteMyApiToken(acc *model.Account, id string) (*model.ResponseDeleteApiToken, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing delete API token %s for %s", id, acc.Account)

	tokens, err := p.ListApiTokens(acc.Account)
	if err != nil {
		p.ProcLog.Errorf("Failed to list API tokens of %s: %v", acc.Account, err)
		return nil, errInternal("Failed to list API tokens")
	}
	if !slices.ContainsFunc(tokens, func(t *model.ApiToken) bool { return t.ID == id }) {
		return nil, errNotFound("API token not found")
	}

	if err := p.DeleteApiToken(id); err != nil {
		if errors.Is(err, context.ErrApiTokenNotFound) {
			return nil, errNotFound("API token not found")
		}
		p.ProcLog.Errorf("Failed to delete API token %s: %v", id, err)
		return nil, errInternal("Failed to delete API token")
	}
	p.ProcLog.Infof("API token %s of %s revoked", id, acc.Account)

	return &model.ResponseDeleteApiToken{
		Message: "Delete API token successful",
	}, nil
}

// AuthenticateApiToken resolves a token (from an Authorization header) to its
// account, the same way a JWT resolves through Authenticate.
func (p *Processor) AuthenticateApiToken(plain string) (*model.Account, *model.ErrorDetail) {
	hash := hashApiToken(plain)
	token, err := p.GetApiToken(hash)
	if errors.Is(err, context.ErrApiTokenNotFound) {
		return nil, &model.ErrorDetail{
			HttpStatus: http.StatusUnauthorized,
			Detail:     "Invalid API token",
		}
	}
	if err != nil {
		p.ProcLog.Errorf("Failed to get an API token: %v", err)
		return nil, errInternal("Failed to check the API token")
	}

	now := p.now()
	if !now.Before(token.ExpiresAt) {
		return nil, &model.ErrorDetail{
			HttpStatus: http.StatusUnauthorized,
			Detail:     "API token expired",
		}
	}

	acc, errDetail := p.Authenticate(token.Account)
	if errDetail != nil {
		return nil, errDetail
	}

	if now.Sub(token.LastUsedAt) >= apiTokenTouchInterval {
		if err := p.TouchApiToken(hash, now); err != nil {
			// only the usage record is lost; the request goes on
			p.ProcLog.Errorf("Failed to record the use of API token %s: %v", token.ID, err)
		}
	}
	return acc, nil
}

func hashApiToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
