package mid

import (
	"net/http"
	"reflect"

	"github.com/wayne011872/goSterna"
	apiErr "github.com/wayne011872/goSterna/api/err"
	myStorage "github.com/wayne011872/goSterna/storage"
	"github.com/gin-gonic/gin"
)

type DevDIMiddle string

func NewGinDevDiMid(fileStorage myStorage.Storage, di interface{}, service string) GinMiddle {
	return &devDiMiddle{
		service: service,
		storage: fileStorage,
		di:      di,
	}
}

type devDiMiddle struct {
	service string
	storage myStorage.Storage
	di      interface{}
}

func (lm *devDiMiddle) outputErr(c *gin.Context, err error) {
	apiErr.GinOutputErr(c, lm.service, err)
}

func (lm *devDiMiddle) GetName() string {
	return "di"
}

func (am *devDiMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		val := reflect.ValueOf(am.di)
		if val.Kind() == reflect.Ptr {
			val = reflect.Indirect(val)
		}
		newValue := reflect.New(val.Type()).Interface()
		confByte, err := am.storage.Get("config.yml")
		if err != nil {
			am.outputErr(c, apiErr.NewApiError(http.StatusInternalServerError, err.Error()))
			c.Abort()
			return
		}
		goSterna.InitConfByByte(confByte, newValue)
		c.Set(string(goSterna.CtxServDiKey),newValue)

		c.Next()
	}
}
