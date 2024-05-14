package mid

import (
	"github.com/gin-gonic/gin"

	"github.com/wayne011872/goSterna/auth"
)

type GinMiddle interface {
	GetName()string
	Handler()gin.HandlerFunc
}

type AuthGinMidInter interface {
	GinMiddle
	AddAuthPath(path string, method string, isAuth bool, group []auth.UserPerm)
}