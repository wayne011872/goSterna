package mid

import (
	"reflect"

	"github.com/gin-gonic/gin"
	goSterna "github.com/wayne011872/goSterna"
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
		val := reflect.ValueOf(am.di)
		if val.Kind() == reflect.Ptr {
			val = reflect.Indirect(val)
		}
		newValue := reflect.New(val.Type()).Interface()
		goSterna.InitDefaultConf(".",newValue)
		c.Set(string(goSterna.CtxServDiKey),newValue)

		c.Next()
	}
}