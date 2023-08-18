package mid

import(
	"runtime"
	"net/http"
	"github.com/wayne011872/goSterna/notify"
	goSterna "github.com/wayne011872/goSterna"
	apiErr "github.com/wayne011872/goSterna/api/err"
	"github.com/gin-gonic/gin"
)

type LineMidDI interface {
	notify.LineDI
}

type LineMiddle string

func NewGinLineMid(service string) GinMiddle {
	return &lineMiddle{
		service: service,
	}
}

type lineMiddle struct {
	service string
}

func (lm *lineMiddle) GetName() string{
	return "line"
}

func(lm *lineMiddle) outputErr(c *gin.Context, err error) {
	apiErr.GinOutputErr(c,lm.service,err)
}

func (lm *lineMiddle) Handler() gin.HandlerFunc{
	return func(c *gin.Context) {
		servDi, ok := c.Get(string(goSterna.CtxServDiKey))
		if !ok || servDi == nil {
			lm.outputErr(c,apiErr.NewApiError(http.StatusInternalServerError, "can not get di"))
			c.Abort()
			return
		}
		if linedi,ok := servDi.(LineMidDI); ok {
			m := linedi.NewLine()
			c.Set(string(notify.CtxLineKey),m)
			c.Next()
			runtime.GC()
		}else{
			lm.outputErr(c,apiErr.NewApiError(http.StatusInternalServerError, "invalid di"))
			c.Abort()
			return
		}
	}
}