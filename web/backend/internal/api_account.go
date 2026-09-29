package internal

import (
	"backend/model"
	"net/http"

	"github.com/free-ran-ue/util"
	"github.com/gin-gonic/gin"
)

func (b *backend) getAccountRoutes() util.Routes {
	return util.Routes{
		{
			Name:        "Login",
			Method:      http.MethodPost,
			Pattern:     "/login",
			HandlerFunc: withLogging("Login", b.AccLog, b.handleLogin),
		},
		{
			Name:        "Logout",
			Method:      http.MethodPost,
			Pattern:     "/logout",
			HandlerFunc: withLogging("Logout", b.AccLog, b.handleLogout),
		},
	}
}

// routes for the logged-in account itself, served behind the auth middleware
func (b *backend) getMeRoutes() util.Routes {
	return util.Routes{
		{
			Name:        "GetMe",
			Method:      http.MethodGet,
			Pattern:     "/me",
			HandlerFunc: withLogging("GetMe", b.AccLog, b.handleGetMe),
		},
		{
			Name:        "UpdateMe",
			Method:      http.MethodPut,
			Pattern:     "/me",
			HandlerFunc: withLogging("UpdateMe", b.AccLog, b.handleUpdateMe),
		},
		{
			Name:        "ChangeMyPassword",
			Method:      http.MethodPut,
			Pattern:     "/me/password",
			HandlerFunc: withLogging("ChangeMyPassword", b.AccLog, b.handleChangeMyPassword),
		},
	}
}

func (b *backend) handleLogin(c *gin.Context) {
	var req model.RequestLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		b.AccLog.Warnf("Invalid login request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseLogin{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.Login(&req)
	if errDetail != nil {
		b.AccLog.Warnf("Login failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseLogin{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleLogout(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (b *backend) handleGetMe(c *gin.Context) {
	c.JSON(http.StatusOK, b.Processor.GetMe(currentAccount(c)))
}

func (b *backend) handleUpdateMe(c *gin.Context) {
	var req model.RequestUpdateMe
	if err := c.ShouldBindJSON(&req); err != nil {
		b.AccLog.Warnf("Invalid update me request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseUpdateMe{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.UpdateMe(currentAccount(c), &req)
	if errDetail != nil {
		b.AccLog.Warnf("Update me failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseUpdateMe{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleChangeMyPassword(c *gin.Context) {
	var req model.RequestChangeMyPassword
	if err := c.ShouldBindJSON(&req); err != nil {
		b.AccLog.Warnf("Invalid change password request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseChangeMyPassword{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.ChangeMyPassword(currentAccount(c), &req)
	if errDetail != nil {
		b.AccLog.Warnf("Change password failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseChangeMyPassword{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
