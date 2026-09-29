package internal

import (
	"backend/model"
	"fmt"
	"io"
	"net/http"

	"github.com/free-ran-ue/util"
	"github.com/gin-gonic/gin"
)

// largest backup zip accepted by a restore
const maxRestoreBytes = 50 << 20

// backup, restore and reset routes, served behind the auth + admin middleware
func (b *backend) getSystemRoutes() util.Routes {
	return util.Routes{
		{
			Name:        "DownloadBackup",
			Method:      http.MethodGet,
			Pattern:     "/system/backup",
			HandlerFunc: withLogging("DownloadBackup", b.SysLog, b.handleDownloadBackup),
		},
		{
			Name:        "RestoreBackup",
			Method:      http.MethodPost,
			Pattern:     "/system/restore",
			HandlerFunc: withLogging("RestoreBackup", b.SysLog, b.handleRestoreBackup),
		},
		{
			Name:        "ResetSystem",
			Method:      http.MethodPost,
			Pattern:     "/system/reset",
			HandlerFunc: withLogging("ResetSystem", b.SysLog, b.handleResetSystem),
		},
	}
}

func (b *backend) handleDownloadBackup(c *gin.Context) {
	data, fileName, errDetail := b.Processor.CreateBackup()
	if errDetail != nil {
		b.SysLog.Warnf("Download backup failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, gin.H{
			"message": errDetail.Detail,
		})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	c.Data(http.StatusOK, "application/zip", data)
}

func (b *backend) handleRestoreBackup(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRestoreBytes)
	header, err := c.FormFile("file")
	if err != nil {
		b.SysLog.Warnf("Invalid restore request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseRestore{
			Message: "A backup zip of at most 50 MB is required in the file field",
		})
		return
	}
	file, err := header.Open()
	if err != nil {
		b.SysLog.Warnf("Invalid restore request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseRestore{
			Message: "Invalid request",
		})
		return
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(file)
	if err != nil {
		b.SysLog.Warnf("Invalid restore request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseRestore{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.RestoreBackup(data)
	if errDetail != nil {
		b.SysLog.Warnf("Restore backup failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseRestore{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (b *backend) handleResetSystem(c *gin.Context) {
	var req model.RequestReset
	if err := c.ShouldBindJSON(&req); err != nil {
		b.SysLog.Warnf("Invalid reset request from %s: %v\n", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, model.ResponseReset{
			Message: "Invalid request",
		})
		return
	}

	response, errDetail := b.Processor.ResetAll(&req)
	if errDetail != nil {
		b.SysLog.Warnf("Reset failed for %s: %s", c.ClientIP(), errDetail.Detail)
		c.JSON(errDetail.HttpStatus, model.ResponseReset{
			Message: errDetail.Detail,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
