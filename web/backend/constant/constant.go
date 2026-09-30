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

// work table
const (
	WORK_DATE_LAYOUT = "2006-01-02"
	WORK_MAX_HOURS   = 24
	WORK_HOURS_STEP  = 0.5
)

// WORK_CATEGORY_COLORS is the category color palette; the frontend maps each
// name to its colors. New categories get the least used one.
var WORK_CATEGORY_COLORS = []string{
	"blue",
	"sky",
	"teal",
	"green",
	"lime",
	"amber",
	"orange",
	"red",
	"pink",
	"purple",
}

// holiday calendar
const (
	HOLIDAY_TYPE_HOLIDAY = "holiday" // a day off on a weekday
	HOLIDAY_TYPE_WORKDAY = "workday" // a working weekend day (makeup day)

	HOLIDAY_SOURCE_GOV    = "gov"    // synced from the government office calendar
	HOLIDAY_SOURCE_MANUAL = "manual" // set by an admin, wins over gov

	WORK_DAILY_HOURS = 8

	// days checked for missing entries, ending at the client's yesterday
	WORK_MISSING_DAYS = 30
)

// backup
const (
	BACKUP_APP     = "work-tracker"
	BACKUP_VERSION = 1

	// the text a reset request must carry
	RESET_CONFIRM = "RESET"
)

// api tokens
const (
	// every API token starts with this, which tells it apart from a JWT
	API_TOKEN_PREFIX = "wt_"

	API_TOKEN_MAX_PER_ACCOUNT = 10
	API_TOKEN_DEFAULT_DAYS    = 365
)

// API_TOKEN_EXPIRY_DAYS are the lifetimes a token can be created with.
var API_TOKEN_EXPIRY_DAYS = []int{30, 60, 180, 365}
