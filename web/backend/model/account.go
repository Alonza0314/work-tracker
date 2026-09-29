package model

// Account is the persisted user record; Password holds a bcrypt hash.
type Account struct {
	Account  string `json:"account"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
	I18n     string `json:"i18n"`
	IsSystem bool   `json:"isSystem"`
}

// User is the public view of an Account (no password).
type User struct {
	Account  string `json:"account"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	I18n     string `json:"i18n"`
	IsSystem bool   `json:"isSystem"`
}

func NewUser(acc *Account) User {
	return User{
		Account:  acc.Account,
		Name:     acc.Name,
		Role:     acc.Role,
		I18n:     acc.I18n,
		IsSystem: acc.IsSystem,
	}
}

type RequestLogin struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ResponseLogin struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

type ResponseGetMe struct {
	Message string `json:"message"`
	User    *User  `json:"user,omitempty"`
}

type RequestUpdateMe struct {
	I18n string `json:"i18n" binding:"required,oneof=zh-TW en"`
}

type ResponseUpdateMe struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

type RequestChangeMyPassword struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type ResponseChangeMyPassword struct {
	Message string `json:"message"`
}

type ResponseListUsers struct {
	Message string `json:"message"`
	Users   []User `json:"users"`
}

// RequestCreateUser has no password: a new user's password is its account.
type RequestCreateUser struct {
	Account string `json:"account" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Role    string `json:"role" binding:"required,oneof=admin default"`
	I18n    string `json:"i18n" binding:"required,oneof=zh-TW en"`
}

type ResponseCreateUser struct {
	Message string `json:"message"`
	User    *User  `json:"user,omitempty"`
}

// RequestUpdateUser fields are optional; only non-nil fields are applied.
type RequestUpdateUser struct {
	Name     *string `json:"name" binding:"omitempty,min=1"`
	Password *string `json:"password" binding:"omitempty,min=1"`
	Role     *string `json:"role" binding:"omitempty,oneof=admin default"`
	I18n     *string `json:"i18n" binding:"omitempty,oneof=zh-TW en"`
}

type ResponseUpdateUser struct {
	Message string `json:"message"`
	User    *User  `json:"user,omitempty"`
}

type ResponseDeleteUser struct {
	Message string `json:"message"`
}
