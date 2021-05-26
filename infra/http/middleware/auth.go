package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/infra/http/presenter"
	"github.com/minghsu0107/saga-product/repo"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrTokenExpired is token expired error
	ErrTokenExpired = errors.New("token expired")
)

func extractToken(r *http.Request) string {
	bearToken := r.Header.Get(conf.JWTAuthHeader)
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

// JWTAuth authorize a request by checking jwt token in the Authentication header
func (m *JWTAuthChecker) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := extractToken(c.Request)
		if accessToken == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		authResult, err := m.repo.Auth(c.Request.Context(), accessToken)
		if err != nil {
			m.logger.Error(err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if authResult.Expired {
			c.AbortWithStatusJSON(http.StatusUnauthorized, presenter.ErrResponse{
				Message: ErrTokenExpired.Error(),
			})
			return
		}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), conf.CustomerKey, authResult.CustomerID))
		c.Next()
	}
}

// JWTAuthChecker is the jwt authorization middleware type
type JWTAuthChecker struct {
	repo   repo.AuthRepository
	logger *log.Entry
}

// NewJWTAuthChecker is the factory of JWTAuthChecker
func NewJWTAuthChecker(config *conf.Config, repo repo.AuthRepository) *JWTAuthChecker {
	return &JWTAuthChecker{
		repo: repo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "middleware:JWTAuthChecker",
		}),
	}
}
