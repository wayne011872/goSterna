package mid

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/wayne011872/goSterna/auth"
	"github.com/wayne011872/goSterna/log"
	"github.com/wayne011872/goSterna/util"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func NewAuthMid(token auth.JwtToken, kid string) AuthMidInter {
	return &authMiddle{
		token:    token,
		kid:      kid,
		authMap:  make(map[string]uint8),
		groupMap: make(map[string][]auth.UserPerm),
	}
}

func (lm *authMiddle) GetName() string {
	return "auth"
}

type authMiddle struct {
	token    auth.JwtToken
	kid      string
	log      log.Logger
	authMap  map[string]uint8
	groupMap map[string][]auth.UserPerm
}

const (
	PenddingMinute = 24 * 60 //閒置自動登出時間，單位分鐘
	authValue      = uint8(1 << iota)
	remoteValue    = uint8(1 << iota)

	AuthTokenKey   = "Auth-Token"
	RemoteTokenKey = "Remote-Token"
)

var ()

func getPathKey(path, method string) string {
	return fmt.Sprintf("%s:%s", path, method)
}

func (am *authMiddle) AddAuthPath(path string, method string, isAuth bool, group []auth.UserPerm) {
	value := uint8(0)
	if isAuth {
		value = value | authValue
	}
	key := getPathKey(path, method)
	am.authMap[key] = uint8(value)
	am.groupMap[key] = group
}

func (am *authMiddle) IsAuth(path string, method string) bool {
	key := getPathKey(path, method)
	value, ok := am.authMap[key]
	if ok {
		return (value & authValue) > 0
	}
	return false
}

func (am *authMiddle) HasGroup(path, method string, group string) bool {
	key := fmt.Sprintf("%s:%s", path, method)
	groupAry, ok := am.groupMap[key]
	if !ok || groupAry == nil || len(groupAry) == 0 {
		return true
	}
	for _, g := range groupAry {
		if string(g) == group {
			return true
		}
	}
	return false
}

func (am *authMiddle) GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc {
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
				authToken := r.Header.Get(AuthTokenKey)
				if authToken == "" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("miss token"))
					return
				}

				jwtToken, err := am.token.ParseToken(authToken)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(err.Error()))
					return
				}

				kid, ok := jwtToken.Header["kid"]
				if !ok || kid != am.kid {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("kid error"))
					return
				}

				mapClaims := jwtToken.Claims.(jwt.MapClaims)
				iss, ok := mapClaims["iss"].(string)
				if !ok || iss != r.Host {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("iss error"))
					return
				}
				permission, ok := mapClaims["per"].(string)
				if hasPerm := am.HasGroup(path, r.Method, permission); ok && !hasPerm {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("permission error"))
					return
				}
				r.Header.Set("isLogin", "true")

				usage, ok := jwtToken.Header["usa"]
				if !ok {
					reqUser := auth.NewReqUser(
						iss,
						mapClaims["sub"].(string),
						mapClaims["acc"].(string),
						mapClaims["nam"].(string),
						[]string{permission},
					)
					r = util.SetCtxKeyVal(r, auth.CtxUserInfoKey, reqUser)
				} else if usage == "access" {
					source := mapClaims["source"].(string)
					id := mapClaims["sourceId"].(string)
					if !strings.Contains(r.RequestURI, util.StrAppend(source, "/", id)) {
						w.WriteHeader(http.StatusForbidden)
						w.Write([]byte("token permision invalid"))
						return
					}
					reqUser := auth.NewAccessGuest(
						iss,
						source,
						id,
						r.RemoteAddr,
						"guest",
						mapClaims["db"].(string),
						[]string{permission},
					)
					r = util.SetCtxKeyVal(r, auth.CtxUserInfoKey, reqUser)
				} else if usage == "comp" {
					reqUser := auth.NewCompUser(
						iss,
						mapClaims["sub"].(string),
						mapClaims["acc"].(string),
						mapClaims["nam"].(string),
						mapClaims["compID"].(string),
						mapClaims["comp"].(string),
						[]string{permission},
					)
					r = util.SetCtxKeyVal(r, auth.CtxUserInfoKey, reqUser)
				}
			} else {
				ip, _, _ := net.SplitHostPort(r.RemoteAddr)
				reqUser := auth.NewGuestUser(r.Host, ip)
				r = util.SetCtxKeyVal(r, auth.CtxUserInfoKey, reqUser)
			}
			f(w, r)
		}
	}
}
