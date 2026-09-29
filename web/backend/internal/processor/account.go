package processor

import (
	"backend/constant"
	"backend/internal/context"
	"backend/model"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/free-ran-ue/util"
	"golang.org/x/crypto/bcrypt"
)

// MigrateAccountCase renames accounts stored before accounts became
// case-insensitive to their upper-case form, re-owning their work records and
// todos. It fails without changing anything when two accounts would collide.
func (p *Processor) MigrateAccountCase() error {
	accounts, err := p.ListAccounts()
	if err != nil {
		return fmt.Errorf("failed to list accounts: %v", err)
	}

	owners := map[string]string{}
	for _, acc := range accounts {
		upper := normalizeAccount(acc.Account)
		if other, taken := owners[upper]; taken {
			return fmt.Errorf("accounts %q and %q both become %q; rename one of them first", other, acc.Account, upper)
		}
		owners[upper] = acc.Account
	}

	for _, acc := range accounts {
		upper := normalizeAccount(acc.Account)
		if acc.Account == upper {
			continue
		}
		if err := p.RenameAccount(acc.Account, upper); err != nil {
			return fmt.Errorf("failed to rename account %s to %s: %v", acc.Account, upper, err)
		}
		p.ProcLog.Infof("Account %s renamed to %s", acc.Account, upper)
	}
	return nil
}

// InitSystemAdmin upper-cases stored accounts (MigrateAccountCase), then makes
// sure the config admin exists in the db as the only system account, with the
// password from config. Its i18n is kept.
func (p *Processor) InitSystemAdmin() error {
	if err := p.MigrateAccountCase(); err != nil {
		return err
	}

	accounts, err := p.ListAccounts()
	if err != nil {
		return fmt.Errorf("failed to list accounts: %v", err)
	}

	for _, acc := range accounts {
		if acc.IsSystem && acc.Account != p.username {
			acc.IsSystem = false
			if err := p.UpdateAccount(acc); err != nil {
				return fmt.Errorf("failed to clear system flag of %s: %v", acc.Account, err)
			}
			p.ProcLog.Infof("Account %s is no longer the system admin", acc.Account)
		}
	}

	acc, err := p.GetAccount(p.username)
	switch {
	case errors.Is(err, context.ErrAccountNotFound):
		hash, err := hashPassword(p.password)
		if err != nil {
			return err
		}
		if err := p.CreateAccount(&model.Account{
			Account:  p.username,
			Name:     p.username,
			Password: hash,
			Role:     constant.ROLE_ADMIN,
			I18n:     constant.DEFAULT_I18N,
			IsSystem: true,
		}); err != nil {
			return fmt.Errorf("failed to create system admin: %v", err)
		}
		p.ProcLog.Infof("System admin %s created", p.username)
		return nil
	case err != nil:
		return fmt.Errorf("failed to get system admin: %v", err)
	}

	if ok, legacy := checkPassword(acc.Password, p.password); !ok || legacy {
		hash, err := hashPassword(p.password)
		if err != nil {
			return err
		}
		acc.Password = hash
	}
	if acc.Name == "" {
		acc.Name = acc.Account
	}
	acc.Role = constant.ROLE_ADMIN
	acc.IsSystem = true

	if err := p.UpdateAccount(acc); err != nil {
		return fmt.Errorf("failed to update system admin: %v", err)
	}
	return nil
}

func (p *Processor) Login(req *model.RequestLogin) (*model.ResponseLogin, *model.ErrorDetail) {
	account := normalizeAccount(req.Account)
	p.ProcLog.Debugf("Processing login for account: %s", account)

	acc, err := p.GetAccount(account)
	if err != nil && !errors.Is(err, context.ErrAccountNotFound) {
		p.ProcLog.Errorf("Failed to get account %s: %v", account, err)
		return nil, errInternal("Failed to get account")
	}
	var ok, legacy bool
	if err == nil {
		ok, legacy = checkPassword(acc.Password, req.Password)
	}
	if !ok {
		return nil, &model.ErrorDetail{
			HttpStatus: http.StatusUnauthorized,
			Detail:     "Invalid account or incorrect password",
		}
	}
	if legacy {
		p.upgradePasswordHash(acc, req.Password)
	}

	token, errDetail := p.createToken(acc)
	if errDetail != nil {
		return nil, errDetail
	}

	return &model.ResponseLogin{
		Message: "Login successful",
		Token:   token,
	}, nil
}

// Authenticate resolves the JWT subject to the current account, so role and
// deletion changes apply to already-issued tokens.
func (p *Processor) Authenticate(account string) (*model.Account, *model.ErrorDetail) {
	account = normalizeAccount(account)
	acc, err := p.GetAccount(account)
	if errors.Is(err, context.ErrAccountNotFound) {
		return nil, &model.ErrorDetail{
			HttpStatus: http.StatusUnauthorized,
			Detail:     "Account no longer exists",
		}
	}
	if err != nil {
		p.ProcLog.Errorf("Failed to get account %s: %v", account, err)
		return nil, errInternal("Failed to get account")
	}
	return acc, nil
}

func (p *Processor) GetMe(acc *model.Account) *model.ResponseGetMe {
	user := model.NewUser(acc)
	return &model.ResponseGetMe{
		Message: "Get me successful",
		User:    &user,
	}
}

func (p *Processor) UpdateMe(acc *model.Account, req *model.RequestUpdateMe) (*model.ResponseUpdateMe, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing update me for account: %s", acc.Account)

	updated := *acc
	updated.I18n = req.I18n
	if errDetail := p.updateAccount(&updated); errDetail != nil {
		return nil, errDetail
	}

	token, errDetail := p.createToken(&updated)
	if errDetail != nil {
		return nil, errDetail
	}

	return &model.ResponseUpdateMe{
		Message: "Update me successful",
		Token:   token,
	}, nil
}

func (p *Processor) ChangeMyPassword(acc *model.Account, req *model.RequestChangeMyPassword) (*model.ResponseChangeMyPassword, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing change password for account: %s", acc.Account)

	if acc.IsSystem {
		return nil, errForbidden("The system admin password is managed by the config file")
	}
	if ok, _ := checkPassword(acc.Password, req.OldPassword); !ok {
		return nil, errForbidden("Incorrect old password")
	}

	hash, err := hashPassword(req.NewPassword)
	if err != nil {
		p.ProcLog.Errorf("Failed to hash password for %s: %v", acc.Account, err)
		return nil, errInternal("Failed to hash password")
	}

	updated := *acc
	updated.Password = hash
	if errDetail := p.updateAccount(&updated); errDetail != nil {
		return nil, errDetail
	}

	return &model.ResponseChangeMyPassword{
		Message: "Change password successful",
	}, nil
}

func (p *Processor) ListUsers() (*model.ResponseListUsers, *model.ErrorDetail) {
	accounts, err := p.ListAccounts()
	if err != nil {
		p.ProcLog.Errorf("Failed to list accounts: %v", err)
		return nil, errInternal("Failed to list users")
	}

	sort.Slice(accounts, func(i, j int) bool { return accounts[i].Account < accounts[j].Account })
	users := make([]model.User, 0, len(accounts))
	for _, acc := range accounts {
		users = append(users, model.NewUser(acc))
	}

	return &model.ResponseListUsers{
		Message: "List users successful",
		Users:   users,
	}, nil
}

func (p *Processor) CreateUser(req *model.RequestCreateUser) (*model.ResponseCreateUser, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing create user: %s", req.Account)

	account := normalizeAccount(req.Account)
	if account == "" {
		return nil, errBadRequest("Account must not be blank")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errBadRequest("Name must not be blank")
	}

	// the initial password is the account; the user can change it after login
	hash, err := hashPassword(account)
	if err != nil {
		p.ProcLog.Errorf("Failed to hash password for %s: %v", account, err)
		return nil, errInternal("Failed to hash password")
	}

	acc := &model.Account{
		Account:  account,
		Name:     name,
		Password: hash,
		Role:     req.Role,
		I18n:     req.I18n,
		IsSystem: false,
	}
	if err := p.CreateAccount(acc); err != nil {
		if errors.Is(err, context.ErrAccountExists) {
			return nil, &model.ErrorDetail{
				HttpStatus: http.StatusConflict,
				Detail:     "Account already exists",
			}
		}
		p.ProcLog.Errorf("Failed to create account %s: %v", account, err)
		return nil, errInternal("Failed to create user")
	}

	user := model.NewUser(acc)
	return &model.ResponseCreateUser{
		Message: "Create user successful",
		User:    &user,
	}, nil
}

func (p *Processor) UpdateUser(account string, req *model.RequestUpdateUser) (*model.ResponseUpdateUser, *model.ErrorDetail) {
	account = normalizeAccount(account)
	p.ProcLog.Debugf("Processing update user: %s", account)

	acc, errDetail := p.getUser(account)
	if errDetail != nil {
		return nil, errDetail
	}
	if acc.IsSystem {
		return nil, errForbidden("The system admin cannot be modified")
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errBadRequest("Name must not be blank")
		}
		acc.Name = name
	}
	if req.Password != nil {
		hash, err := hashPassword(*req.Password)
		if err != nil {
			p.ProcLog.Errorf("Failed to hash password for %s: %v", account, err)
			return nil, errInternal("Failed to hash password")
		}
		acc.Password = hash
	}
	if req.Role != nil {
		acc.Role = *req.Role
	}
	if req.I18n != nil {
		acc.I18n = *req.I18n
	}

	if errDetail := p.updateAccount(acc); errDetail != nil {
		return nil, errDetail
	}

	user := model.NewUser(acc)
	return &model.ResponseUpdateUser{
		Message: "Update user successful",
		User:    &user,
	}, nil
}

func (p *Processor) DeleteUser(operator *model.Account, account string) (*model.ResponseDeleteUser, *model.ErrorDetail) {
	account = normalizeAccount(account)
	p.ProcLog.Debugf("Processing delete user %s by %s", account, operator.Account)

	acc, errDetail := p.getUser(account)
	if errDetail != nil {
		return nil, errDetail
	}
	if acc.IsSystem {
		return nil, errForbidden("The system admin cannot be deleted")
	}
	if acc.Account == operator.Account {
		return nil, errForbidden("You cannot delete yourself")
	}

	if err := p.DeleteAccount(account); err != nil {
		if errors.Is(err, context.ErrAccountNotFound) {
			return nil, errUserNotFound()
		}
		p.ProcLog.Errorf("Failed to delete account %s: %v", account, err)
		return nil, errInternal("Failed to delete user")
	}

	return &model.ResponseDeleteUser{
		Message: "Delete user successful",
	}, nil
}

func (p *Processor) getUser(account string) (*model.Account, *model.ErrorDetail) {
	acc, err := p.GetAccount(account)
	if errors.Is(err, context.ErrAccountNotFound) {
		return nil, errUserNotFound()
	}
	if err != nil {
		p.ProcLog.Errorf("Failed to get account %s: %v", account, err)
		return nil, errInternal("Failed to get user")
	}
	return acc, nil
}

func (p *Processor) updateAccount(acc *model.Account) *model.ErrorDetail {
	if err := p.UpdateAccount(acc); err != nil {
		if errors.Is(err, context.ErrAccountNotFound) {
			return errUserNotFound()
		}
		p.ProcLog.Errorf("Failed to update account %s: %v", acc.Account, err)
		return errInternal("Failed to update user")
	}
	return nil
}

func (p *Processor) createToken(acc *model.Account) (string, *model.ErrorDetail) {
	token, err := util.CreateJWT(p.jwtSecret, acc.Account, p.jwtExpiresIn, map[string]any{
		constant.JWT_CLAIM_NAME: acc.Name,
		constant.JWT_CLAIM_ROLE: acc.Role,
		constant.JWT_CLAIM_I18N: acc.I18n,
	})
	if err != nil {
		p.ProcLog.Errorf("Failed to create JWT for account %s: %v", acc.Account, err)
		return "", errInternal("Failed to create JWT")
	}
	return token, nil
}

// upgradePasswordHash re-hashes a password that matched a legacy
// case-sensitive hash; failures only cost the upgrade, not the login.
func (p *Processor) upgradePasswordHash(acc *model.Account, password string) {
	hash, err := hashPassword(password)
	if err != nil {
		p.ProcLog.Errorf("Failed to upgrade password hash for %s: %v", acc.Account, err)
		return
	}
	upgraded := *acc
	upgraded.Password = hash
	if err := p.UpdateAccount(&upgraded); err != nil {
		p.ProcLog.Errorf("Failed to upgrade password hash for %s: %v", acc.Account, err)
		return
	}
	p.ProcLog.Infof("Password hash of %s upgraded to case-insensitive", acc.Account)
}

// accounts are case-insensitive: stored and compared in upper case
func normalizeAccount(account string) string {
	return strings.ToUpper(strings.TrimSpace(account))
}

// passwords are case-insensitive: hashed and compared in upper case
func normalizePassword(password string) string {
	return strings.ToUpper(password)
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(normalizePassword(password)), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	return string(hash), nil
}

// checkPassword reports whether password matches hash ignoring case. legacy is
// true when it only matched a hash made before passwords were
// case-insensitive (exact case), so the caller can re-hash it.
func checkPassword(hash, password string) (ok bool, legacy bool) {
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(normalizePassword(password))) == nil {
		return true, false
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
		return true, true
	}
	return false, false
}

func errBadRequest(detail string) *model.ErrorDetail {
	return &model.ErrorDetail{
		HttpStatus: http.StatusBadRequest,
		Detail:     detail,
	}
}

func errUserNotFound() *model.ErrorDetail {
	return errNotFound("User not found")
}

func errNotFound(detail string) *model.ErrorDetail {
	return &model.ErrorDetail{
		HttpStatus: http.StatusNotFound,
		Detail:     detail,
	}
}

func errForbidden(detail string) *model.ErrorDetail {
	return &model.ErrorDetail{
		HttpStatus: http.StatusForbidden,
		Detail:     detail,
	}
}

func errInternal(detail string) *model.ErrorDetail {
	return &model.ErrorDetail{
		HttpStatus: http.StatusInternalServerError,
		Detail:     detail,
	}
}
