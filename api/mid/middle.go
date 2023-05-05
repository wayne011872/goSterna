package mid

import (
	"github.com/gin-gonic/gin"
)

type GinMiddle interface {
	GetName()string
	Handler()gin.HandlerFunc
}