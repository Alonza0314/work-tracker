package internal

import (
	"backend/model"
	"net/http"

	"github.com/free-ran-ue/util"
	"github.com/gin-gonic/gin"
)

// work table routes for any logged-in account, served behind the auth
// middleware; the "everyone" routes check AllowViewAll in the processor
func (b *backend) getWorkRoutes() util.Routes {
	return util.Routes{
		{
			Name:        "GetWorkOptions",
			Method:      http.MethodGet,
			Pattern:     "/work/options",
			HandlerFunc: withLogging("GetWorkOptions", b.WrkLog, b.handleGetWorkOptions),
		},
		{
			Name:        "ListWorkMembers",
			Method:      http.MethodGet,
			Pattern:     "/work/members",
			HandlerFunc: withLogging("ListWorkMembers", b.WrkLog, b.handleListWorkMembers),
		},
		{
			Name:        "ListMissingEntries",
			Method:      http.MethodGet,
			Pattern:     "/work/missing",
			HandlerFunc: withLogging("ListMissingEntries", b.WrkLog, b.handleListMissingEntries),
		},
		{
			Name:        "ListAllWorkRecords",
			Method:      http.MethodGet,
			Pattern:     "/work-records",
			HandlerFunc: withLogging("ListAllWorkRecords", b.WrkLog, b.handleListAllWorkRecords),
		},
		{
			Name:        "ListMyWorkRecords",
			Method:      http.MethodGet,
			Pattern:     "/me/work-records",
			HandlerFunc: withLogging("ListMyWorkRecords", b.WrkLog, b.handleListMyWorkRecords),
		},
		{
			Name:        "CreateMyWorkRecord",
			Method:      http.MethodPost,
			Pattern:     "/me/work-records",
			HandlerFunc: withLogging("CreateMyWorkRecord", b.WrkLog, b.handleCreateMyWorkRecord),
		},
		{
			Name:        "UpdateMyWorkRecord",
			Method:      http.MethodPut,
			Pattern:     "/me/work-records/:id",
			HandlerFunc: withLogging("UpdateMyWorkRecord", b.WrkLog, b.handleUpdateMyWorkRecord),
		},
		{
			Name:        "DeleteMyWorkRecord",
			Method:      http.MethodDelete,
			Pattern:     "/me/work-records/:id",
			HandlerFunc: withLogging("DeleteMyWorkRecord", b.WrkLog, b.handleDeleteMyWorkRecord),
		},
		{
			Name:        "GetWeekSummary",
			Method:      http.MethodGet,
			Pattern:     "/me/week-summary",
			HandlerFunc: withLogging("GetWeekSummary", b.WrkLog, b.handleGetWeekSummary),
		},
		{
			Name:        "ListMyTodos",
			Method:      http.MethodGet,
			Pattern:     "/me/todos",
			HandlerFunc: withLogging("ListMyTodos", b.WrkLog, b.handleListMyTodos),
		},
		{
			Name:        "CreateMyTodo",
			Method:      http.MethodPost,
			Pattern:     "/me/todos",
			HandlerFunc: withLogging("CreateMyTodo", b.WrkLog, b.handleCreateMyTodo),
		},
		{
			Name:        "UpdateMyTodo",
			Method:      http.MethodPut,
			Pattern:     "/me/todos/:id",
			HandlerFunc: withLogging("UpdateMyTodo", b.WrkLog, b.handleUpdateMyTodo),
		},
		{
			Name:        "DeleteMyTodo",
			Method:      http.MethodDelete,
			Pattern:     "/me/todos/:id",
			HandlerFunc: withLogging("DeleteMyTodo", b.WrkLog, b.handleDeleteMyTodo),
		},
		{
			Name:        "CompleteMyTodo",
			Method:      http.MethodPost,
			Pattern:     "/me/todos/:id/complete",
			HandlerFunc: withLogging("CompleteMyTodo", b.WrkLog, b.handleCompleteMyTodo),
		},
	}
}

func (b *backend) handleGetWorkOptions(c *gin.Context) {
	response, errDetail := b.Processor.GetWorkOptions()
	if errDetail != nil {
		b.WrkLog.Warnf("Get work options failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWorkOptions{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleListWorkMembers(c *gin.Context) {
	response, errDetail := b.Processor.ListWorkMembers(currentAccount(c))
	if errDetail != nil {
		b.WrkLog.Warnf("List work members failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWorkMembers{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleListMissingEntries(c *gin.Context) {
	var req model.RequestMissingEntries
	if err := c.ShouldBindQuery(&req); err != nil {
		b.WrkLog.Warnf("Invalid missing entries request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseMissingEntries{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.ListMissingEntries(currentAccount(c), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("List missing entries failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseMissingEntries{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleListAllWorkRecords(c *gin.Context) {
	var req model.RequestListWorkRecords
	if err := c.ShouldBindQuery(&req); err != nil {
		b.WrkLog.Warnf("Invalid list work records request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseWorkRecordList{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.ListAllWorkRecords(currentAccount(c), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("List work records failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWorkRecordList{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleListMyWorkRecords(c *gin.Context) {
	var req model.RequestListMyWorkRecords
	if err := c.ShouldBindQuery(&req); err != nil {
		b.WrkLog.Warnf("Invalid list my work records request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseWorkRecordList{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.ListMyWorkRecords(currentAccount(c), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("List my work records failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWorkRecordList{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleCreateMyWorkRecord(c *gin.Context) {
	var req model.RequestSaveWorkEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		b.WrkLog.Warnf("Invalid create work record request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseWorkRecord{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.CreateMyWorkRecord(currentAccount(c), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("Create work record failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWorkRecord{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleUpdateMyWorkRecord(c *gin.Context) {
	var req model.RequestSaveWorkEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		b.WrkLog.Warnf("Invalid update work record request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseWorkRecord{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.UpdateMyWorkRecord(currentAccount(c), c.Param("id"), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("Update work record failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWorkRecord{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleDeleteMyWorkRecord(c *gin.Context) {
	response, errDetail := b.Processor.DeleteMyWorkRecord(currentAccount(c), c.Param("id"))
	if errDetail != nil {
		b.WrkLog.Warnf("Delete work record failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseDeleteWorkRecord{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleGetWeekSummary(c *gin.Context) {
	var req model.RequestWeekSummary
	if err := c.ShouldBindQuery(&req); err != nil {
		b.WrkLog.Warnf("Invalid week summary request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseWeekSummary{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.GetWeekSummary(currentAccount(c), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("Get week summary failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWeekSummary{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleListMyTodos(c *gin.Context) {
	response, errDetail := b.Processor.ListMyTodos(currentAccount(c))
	if errDetail != nil {
		b.WrkLog.Warnf("List todos failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseTodos{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleCreateMyTodo(c *gin.Context) {
	var req model.RequestSaveWorkEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		b.WrkLog.Warnf("Invalid create todo request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseTodo{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.CreateMyTodo(currentAccount(c), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("Create todo failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseTodo{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleUpdateMyTodo(c *gin.Context) {
	var req model.RequestSaveWorkEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		b.WrkLog.Warnf("Invalid update todo request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseTodo{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.UpdateMyTodo(currentAccount(c), c.Param("id"), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("Update todo failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseTodo{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleDeleteMyTodo(c *gin.Context) {
	response, errDetail := b.Processor.DeleteMyTodo(currentAccount(c), c.Param("id"))
	if errDetail != nil {
		b.WrkLog.Warnf("Delete todo failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseDeleteTodo{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleCompleteMyTodo(c *gin.Context) {
	var req model.RequestCompleteTodo
	if err := c.ShouldBindJSON(&req); err != nil {
		b.WrkLog.Warnf("Invalid complete todo request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseWorkRecord{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.CompleteMyTodo(currentAccount(c), c.Param("id"), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("Complete todo failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseWorkRecord{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
