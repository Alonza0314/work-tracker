package internal

import (
	"backend/model"
	"net/http"

	"github.com/free-ran-ue/util"
	"github.com/gin-gonic/gin"
)

// holiday calendar routes for any logged-in account, served behind the auth
// middleware
func (b *backend) getHolidayRoutes() util.Routes {
	return util.Routes{
		{
			Name:        "ListHolidays",
			Method:      http.MethodGet,
			Pattern:     "/holidays",
			HandlerFunc: withLogging("ListHolidays", b.WrkLog, b.handleListHolidays),
		},
	}
}

// holiday calendar management routes, served behind the auth + admin
// middleware
func (b *backend) getHolidaySettingRoutes() util.Routes {
	return util.Routes{
		{
			Name:        "SyncHolidays",
			Method:      http.MethodPost,
			Pattern:     "/holidays/sync",
			HandlerFunc: withLogging("SyncHolidays", b.WrkLog, b.handleSyncHolidays),
		},
		{
			Name:        "SaveHoliday",
			Method:      http.MethodPut,
			Pattern:     "/holidays/:date",
			HandlerFunc: withLogging("SaveHoliday", b.WrkLog, b.handleSaveHoliday),
		},
		{
			Name:        "DeleteHoliday",
			Method:      http.MethodDelete,
			Pattern:     "/holidays/:date",
			HandlerFunc: withLogging("DeleteHoliday", b.WrkLog, b.handleDeleteHoliday),
		},
	}
}

func (b *backend) handleListHolidays(c *gin.Context) {
	var req model.RequestListHolidays
	if err := c.ShouldBindQuery(&req); err != nil {
		b.WrkLog.Warnf("Invalid list holidays request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseHolidays{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.ListYearHolidays(&req)
	if errDetail != nil {
		b.WrkLog.Warnf("List holidays failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseHolidays{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleSyncHolidays(c *gin.Context) {
	response, errDetail := b.Processor.SyncHolidaysNow()
	if errDetail != nil {
		b.WrkLog.Warnf("Sync holidays failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseSyncHolidays{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleSaveHoliday(c *gin.Context) {
	var req model.RequestSaveHoliday
	if err := c.ShouldBindJSON(&req); err != nil {
		b.WrkLog.Warnf("Invalid save holiday request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseHoliday{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.SaveManualHoliday(c.Param("date"), &req)
	if errDetail != nil {
		b.WrkLog.Warnf("Save holiday failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseHoliday{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleDeleteHoliday(c *gin.Context) {
	response, errDetail := b.Processor.DeleteManualHoliday(c.Param("date"))
	if errDetail != nil {
		b.WrkLog.Warnf("Delete holiday failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseDeleteHoliday{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
