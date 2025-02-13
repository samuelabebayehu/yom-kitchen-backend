package middlewares

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const DBContextKey = "database"

func DatabaseMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), DBContextKey, db)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetDBFromContext(c *gin.Context) *gorm.DB {
	db, ok := c.Request.Context().Value(DBContextKey).(*gorm.DB)
	if !ok || db == nil {
		return nil
	}
	return db
}
