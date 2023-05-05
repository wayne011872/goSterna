package mid

import (
	"reflect"

	"github.com/gin-gonic/gin"
	golanggeneral "github.com/wayne011872/goSterna"
	"github.com/wayne011872/goSterna/util"
)

func NewServiceDiMid(di interface{},service string) GinMiddle {
	return &serviceDiMiddle{
		di:   di,
		service: service,
	}
}

type serviceDiMiddle struct {
	service string
	di interface{}
}

func (lm *serviceDiMiddle) GetName() string {
	return "di"
}

func (am *serviceDiMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		CtxServDiKey := util.CtxKey("ServiceDi")
		val := reflect.ValueOf(am.di)
		if val.Kind() == reflect.Ptr {
			val = reflect.Indirect(val)
		}
		newValue := reflect.New(val.Type()).Interface()
		golanggeneral.InitConfByEnv(newValue)
		c.Request = util.SetCtxKeyVal(c.Request, CtxServDiKey,newValue)

		c.Next()
	}
}