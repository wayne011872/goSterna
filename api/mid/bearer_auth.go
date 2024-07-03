package mid

import (
	"fmt"
	"net/http"
	"strings"

	apiErr "github.com/wayne011872/goSterna/api/err"
	"github.com/wayne011872/goSterna/auth"
	"github.com/wayne011872/goSterna/log"
	"github.com/wayne011872/goSterna/util"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
)

type TokenParserResult interface {
	Host() string
	Perms() []string
	Account() string
	Name() string
	Sub() string
	Target() string
}

type AuthTokenParser func(token string) (TokenParserResult, error)

func NewBearerAuthMid(tokenParser AuthTokenParser, isMatchHost bool) AuthMidInter {
	return &bearAuthMiddle{
		parser:      tokenParser,
		authMap:     make(map[string]uint8),
		groupMap:    make(map[string][]auth.UserPerm),
		isMatchHost: isMatchHost,
	}
}

func NewGinBearAuthMid(service string, isMatchHost bool) AuthGinMidInter {
	return &bearAuthMiddle{
		service:     service,
		authMap:     make(map[string]uint8),
		groupMap:    make(map[string][]auth.UserPerm),
		isMatchHost: isMatchHost,
	}
}

func (lm *bearAuthMiddle) GetName() string {
	return "auth"
}

type bearAuthMiddle struct {
	service     string
	parser      AuthTokenParser
	log         log.Logger
	authMap     map[string]uint8
	groupMap    map[string][]auth.UserPerm
	isMatchHost bool
}

const (
	BearerAuthTokenKey = "Authorization"
)

func (lm *bearAuthMiddle) outputErr(c *gin.Context, err error) {
	apiErr.GinOutputErr(c, lm.service, err)
}

func (am *bearAuthMiddle) AddAuthPath(path string, method string, isAuth bool, group []auth.UserPerm) {
	value := uint8(0)
	if isAuth {
		value = value | authValue
	}
	key := getPathKey(path, method)
	am.authMap[key] = uint8(value)
	am.groupMap[key] = group
}

func (am *bearAuthMiddle) IsAuth(path string, method string) bool {
	key := getPathKey(path, method)
	value, ok := am.authMap[key]
	if ok {
		return (value & authValue) > 0
	}
	return false
}

func (am *bearAuthMiddle) HasPerm(path, method string, perm []string) bool {
	key := fmt.Sprintf("%s:%s", path, method)
	groupAry, ok := am.groupMap[key]
	if !ok || groupAry == nil || len(groupAry) == 0 {
		return true
	}
	for _, g := range groupAry {
		if util.IsStrInList(string(g), perm...) {
			return true
		}
	}
	return false
}

func (am *bearAuthMiddle) GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		// one time scope setup area for middleware
		return func(w http.ResponseWriter, r *http.Request) {
			path, err := mux.CurrentRoute(r).GetPathTemplate()
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(err.Error()))
				return
			}
			if am.IsAuth(path, r.Method) {
				authToken := r.Header.Get(BearerAuthTokenKey)
				if authToken == "" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("miss token"))
					return
				}
				if !strings.HasPrefix(authToken, "Bearer ") {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("invalid token: missing Bearer"))
					return
				}
				authToken = authToken[7:]
				result, err := am.parser(authToken)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("invalid token: " + err.Error()))
					return
				}
				if am.isMatchHost && result.Host() != util.GetHost(r) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(fmt.Sprintf("host not match: [%s] is not [%s]", result.Host(), r.Host)))
					return
				}

				if hasPerm := am.HasPerm(path, r.Method, result.Perms()); !hasPerm {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("permission error"))
					return
				}
				r = util.SetCtxKeyVal(r, auth.CtxUserInfoKey, auth.NewReqUser(
					result.Host(),
					result.Sub(),
					result.Account(),
					result.Name(),
					result.Perms(),
				))
			}
			f(w, r)
		}
	}
}

func (m *bearAuthMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		method := c.Request.Method
		if path == "" {
			m.outputErr(c, apiErr.NewApiError(http.StatusNotFound, "path not found"))
			return
		}
		if m.IsAuth(path, method) {
			authToken := c.GetHeader(BearerAuthTokenKey)
			if authToken == "" {
				m.outputErr(c, apiErr.NewApiError(http.StatusUnauthorized, "miss token"))
				return
			}

			if !strings.HasPrefix(authToken, "Bearer ") {
				m.outputErr(c, apiErr.NewApiError(http.StatusUnauthorized, "invalid token: missing Bearer"))
				return
			}

			u, ok := c.Get(string(auth.CtxUserInfoKey))
			if !ok {
				m.outputErr(c, apiErr.NewApiError(http.StatusBadRequest, "missing token"))
				c.Abort()
				return
			}
			reqUser := u.(auth.ReqUser)

			host := util.GetHost(c.Request)
			if m.isMatchHost && reqUser.Host() != host {
				m.outputErr(c, apiErr.NewApiError(http.StatusUnauthorized,
					fmt.Sprintf("host not match: [%s] is not [%s]", reqUser.Host(), host)))
				return
			}

			if hasPerm := m.HasPerm(path, method, reqUser.GetPerm()); !hasPerm {
				m.outputErr(c, apiErr.NewApiError(http.StatusUnauthorized, "permission error"))
				return
			}
		}
		c.Next()
	}
}
