package internal

import (
	"backend/model"
	"net/http"

	"github.com/free-ran-ue/util"
	"github.com/gin-gonic/gin"
)

// work table settings routes, served behind the auth + admin middleware
func (b *backend) getWorkSettingRoutes() util.Routes {
	return util.Routes{
		{
			Name:        "CreateCategory",
			Method:      http.MethodPost,
			Pattern:     "/categories",
			HandlerFunc: withLogging("CreateCategory", b.WrkLog, b.handleCreateCategory),
		},
		{
			Name:        "UpdateCategory",
			Method:      http.MethodPut,
			Pattern:     "/categories/:id",
			HandlerFunc: withLogging("UpdateCategory", b.WrkLog, b.handleUpdateCategory),
		},
		{
			Name:        "DeleteCategory",
			Method:      http.MethodDelete,
			Pattern:     "/categories/:id",
			HandlerFunc: withLogging("DeleteCategory", b.WrkLog, b.handleDeleteCategory),
		},
		{
			Name:        "CreateProject",
			Method:      http.MethodPost,
			Pattern:     "/projects",
			HandlerFunc: withLogging("CreateProject", b.WrkLog, b.handleCreateProject),
		},
		{
			Name:        "UpdateProject",
			Method:      http.MethodPut,
			Pattern:     "/projects/:id",
			HandlerFunc: withLogging("UpdateProject", b.WrkLog, b.handleUpdateProject),
		},
		{
			Name:        "DeleteProject",
			Method:      http.MethodDelete,
			Pattern:     "/projects/:id",
			HandlerFunc: withLogging("DeleteProject", b.WrkLog, b.handleDeleteProject),
		},
		{
			Name:        "UpdateWorkSetting",
			Method:      http.MethodPut,
			Pattern:     "/settings/work",
			HandlerFunc: withLogging("UpdateWorkSetting", b.WrkLog, b.handleUpdateWorkSetting),
		},
	}
}

func (b *backend) handleCreateCategory(c *gin.Context) {
	b.handleCreateWorkOption(c, "category", b.Processor.CreateWorkCategory)
}

func (b *backend) handleUpdateCategory(c *gin.Context) {
	b.handleUpdateWorkOption(c, "category", b.Processor.UpdateWorkCategory)
}

func (b *backend) handleDeleteCategory(c *gin.Context) {
	b.handleDeleteWorkOption(c, "category", b.Processor.DeleteWorkCategory)
}

func (b *backend) handleCreateProject(c *gin.Context) {
	b.handleCreateWorkOption(c, "project", b.Processor.CreateWorkProject)
}

func (b *backend) handleUpdateProject(c *gin.Context) {
	b.handleUpdateWorkOption(c, "project", b.Processor.UpdateWorkProject)
}

func (b *backend) handleDeleteProject(c *gin.Context) {
	b.handleDeleteWorkOption(c, "project", b.Processor.DeleteWorkProject)
}

func (b *backend) handleCreateWorkOption(c *gin.Context, kind string, create func(*model.RequestCreateWorkOption) (*model.ResponseWorkOption, *model.ErrorDetail)) {
	var req model.RequestCreateWorkOption
	if err := c.ShouldBindJSON(&req); err != nil {
		b.WrkLog.Warnf("Invalid create %s request from %s: %v\n", kind, c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseWorkOption{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := create(&req)
	if errDetail != nil {
		b.WrkLog.Warnf("Create %s failed for %s: %s", kind, c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWorkOption{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleUpdateWorkOption(c *gin.Context, kind string, update func(string, *model.RequestUpdateWorkOption) (*model.ResponseWorkOption, *model.ErrorDetail)) {
	var req model.RequestUpdateWorkOption
	if err := c.ShouldBindJSON(&req); err != nil {
		b.WrkLog.Warnf("Invalid update %s request from %s: %v\n", kind, c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseWorkOption{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := update(c.Param("id"), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("Update %s failed for %s: %s", kind, c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWorkOption{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleDeleteWorkOption(c *gin.Context, kind string, remove func(string) (*model.ResponseDeleteWorkOption, *model.ErrorDetail)) {
	response, errDetail := remove(c.Param("id"))
	if errDetail != nil {
		b.WrkLog.Warnf("Delete %s failed for %s: %s", kind, c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseDeleteWorkOption{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleUpdateWorkSetting(c *gin.Context) {
	var req model.RequestUpdateWorkSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		b.WrkLog.Warnf("Invalid update work setting request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseWorkSetting{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.SaveWorkSetting(&req)
	if errDetail != nil {
		b.WrkLog.Warnf("Update work setting failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWorkSetting{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
