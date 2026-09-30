package internal

import (
	"backend/constant"
	"backend/model"
	"net/http"
	"strings"

	"github.com/free-ran-ue/util"
	"github.com/gin-gonic/gin"
)

func addMiddleware(g *gin.Engine) {
	g.Use(middlewareExample)
}

func middlewareExample(c *gin.Context) {
	// do something before request

	c.Next()
}

// addAuthMiddleware validates the bearer credential and loads the current
// account from the db into the gin context (read it with currentAccount). The
// credential is an API token when it starts with API_TOKEN_PREFIX, else a JWT.
func addAuthMiddleware(b *backend) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Authorization format must be Bearer <token>",
			})
			c.Abort()
			return
		}

		var acc *model.Account
		var errDetail *model.ErrorDetail
		if strings.HasPrefix(parts[1], constant.API_TOKEN_PREFIX) {
			acc, errDetail = b.Processor.AuthenticateApiToken(parts[1])
		} else {
			claims, err := util.ValidateJWT(parts[1], b.jwt.secret)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message": "Invalid token: " + err.Error(),
				})
				c.Abort()
				return
			}

			account, _ := claims["sub"].(string)
			acc, errDetail = b.Processor.Authenticate(account)
		}
		if errDetail != nil {
			c.JSON(errDetail.HttpStatus, gin.H{
				"message": errDetail.Detail,
			})
			c.Abort()
			return
		}

		c.Set(constant.CTX_KEY_ACCOUNT, acc)
		c.Next()
	}
}

// addAdminMiddleware must run after addAuthMiddleware.
func addAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if currentAccount(c).Role != constant.ROLE_ADMIN {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Admin permission required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func currentAccount(c *gin.Context) *model.Account {
	return c.MustGet(constant.CTX_KEY_ACCOUNT).(*model.Account)
}
