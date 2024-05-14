package mid

import (
	"net/http"
	"strings"

	apiErr "github.com/wayne011872/goSterna/api/err"
	"github.com/wayne011872/goSterna/auth"
	"github.com/gin-gonic/gin"
)

func NewGinTokenParserMid(service string, parser AuthTokenParser) GinMiddle {
	return &reqUserTokenMiddle{
		service: service,
		parser:  parser,
	}
}

type reqUserTokenMiddle struct {
	service string
	parser  AuthTokenParser
}

func (lm *reqUserTokenMiddle) GetName() string {
	return "token"
}

func (lm *reqUserTokenMiddle) outputErr(c *gin.Context, err error) {
	apiErr.GinOutputErr(c, lm.service, err)
}

func (m *reqUserTokenMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {

		authToken := c.GetHeader(BearerAuthTokenKey)
		if authToken == "" {
			return
		}

		if !strings.HasPrefix(authToken, "Bearer ") {
			m.outputErr(c, apiErr.NewApiError(http.StatusUnauthorized, "invalid token: missing Bearer"))
			return
		}
		authToken = authToken[7:]
		result, err := m.parser(authToken)
		if err != nil {
			m.outputErr(c, apiErr.NewApiError(http.StatusUnauthorized, "invalid token: "+err.Error()))
			return
		}

		c.Set(string(auth.CtxUserInfoKey), auth.NewReqUser(
			result.Host(), result.Sub(), result.Account(),
			result.Name(), result.Perms()))
	}
}
