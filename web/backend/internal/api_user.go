package internal

import (
	"backend/model"
	"net/http"

	"github.com/free-ran-ue/util"
	"github.com/gin-gonic/gin"
)

// user management routes, served behind the auth + admin middleware
func (b *backend) getUserRoutes() util.Routes {
	return util.Routes{
		{
			Name:        "ListUsers",
			Method:      http.MethodGet,
			Pattern:     "/users",
			HandlerFunc: withLogging("ListUsers", b.UsrLog, b.handleListUsers),
		},
		{
			Name:        "CreateUser",
			Method:      http.MethodPost,
			Pattern:     "/users",
			HandlerFunc: withLogging("CreateUser", b.UsrLog, b.handleCreateUser),
		},
		{
			Name:        "UpdateUser",
			Method:      http.MethodPut,
			Pattern:     "/users/:account",
			HandlerFunc: withLogging("UpdateUser", b.UsrLog, b.handleUpdateUser),
		},
		{
			Name:        "DeleteUser",
			Method:      http.MethodDelete,
			Pattern:     "/users/:account",
			HandlerFunc: withLogging("DeleteUser", b.UsrLog, b.handleDeleteUser),
		},
	}
}

func (b *backend) handleListUsers(c *gin.Context) {
	response, errDetail := b.Processor.ListUsers()
	if errDetail != nil {
		b.UsrLog.Warnf("List users failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseListUsers{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleCreateUser(c *gin.Context) {
	var req model.RequestCreateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		b.UsrLog.Warnf("Invalid create user request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseCreateUser{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.CreateUser(&req)
	if errDetail != nil {
		b.UsrLog.Warnf("Create user failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseCreateUser{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleUpdateUser(c *gin.Context) {
	var req model.RequestUpdateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		b.UsrLog.Warnf("Invalid update user request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseUpdateUser{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.UpdateUser(c.Param("account"), &req)
	if errDetail != nil {
		b.UsrLog.Warnf("Update user failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseUpdateUser{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleDeleteUser(c *gin.Context) {
	response, errDetail := b.Processor.DeleteUser(currentAccount(c), c.Param("account"))
	if errDetail != nil {
		b.UsrLog.Warnf("Delete user failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseDeleteUser{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
