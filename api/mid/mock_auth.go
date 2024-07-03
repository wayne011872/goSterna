package mid

import (
	"strings"

	"github.com/wayne011872/goSterna/auth"
	"github.com/wayne011872/goSterna/util"
	"github.com/gin-gonic/gin"
)

func NewMockAuthMid() AuthGinMidInter {
	return &mockAuthMiddle{}
}

type mockAuthMiddle struct {
}

func (lm *mockAuthMiddle) GetName() string {
	return "mockAuth"
}

func (am *mockAuthMiddle) AddAuthPath(path string, method string, isAuth bool, group []auth.UserPerm) {
}

func (am *mockAuthMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("Mock_User_UID")
		if userID == "" {
			userID = "mock-id"
		}
		userAcc := c.GetHeader("Mock_User_ACC")
		if userAcc == "" {
			userAcc = "mock-account"
		}
		userName := c.GetHeader("Mock_User_NAM")
		if userName == "" {
			userName = "mock-name"
		}
		roles := strings.Split(c.GetHeader("Mock_User_Roles"), ",")
		if len(roles) == 0 {
			roles = []string{"mock"}
		}
		c.Set(
			string(auth.CtxUserInfoKey),
			auth.NewReqUser(util.GetHost(c.Request), userID, userAcc, userName, roles),
		)
		c.Next()
	}
}
