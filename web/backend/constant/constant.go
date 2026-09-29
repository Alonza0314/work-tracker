package constant

// api prefix
const (
	API_PREFIX = "/api"
)

// account role
const (
	ROLE_ADMIN   = "admin"
	ROLE_DEFAULT = "default"
)

// account i18n
const (
	I18N_ZH_TW = "zh-TW"
	I18N_EN    = "en"

	DEFAULT_I18N = I18N_ZH_TW
)

// jwt claims
const (
	JWT_CLAIM_NAME = "name"
	JWT_CLAIM_ROLE = "role"
	JWT_CLAIM_I18N = "i18n"
)

// gin context keys
const (
	CTX_KEY_ACCOUNT = "account"
)
