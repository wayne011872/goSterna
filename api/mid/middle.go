package mid

import (
	"net/http"
	"github.com/gin-gonic/gin"

	"github.com/wayne011872/goSterna/auth"
)
type Middle interface {
	GetName() string
	GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc
}
type Middleware func(http.HandlerFunc) http.HandlerFunc

// buildChain builds the middlware chain recursively, functions are first class
func BuildChain(f http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	// if our chain is done, use the original handlerfunc
	if len(m) == 0 {
		return f
	}
	// otherwise nest the handlerfuncs
	return m[0](BuildChain(f, m[1:]...))
}

type AuthMidInter interface {
	Middle
	AddAuthPath(path string, method string, isAuth bool, group []auth.UserPerm)
}

type GinMiddle interface {
	GetName()string
	Handler()gin.HandlerFunc
}

type AuthGinMidInter interface {
	GinMiddle
	AddAuthPath(path string, method string, isAuth bool, group []auth.UserPerm)
}