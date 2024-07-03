package mid

import(
	"runtime"
	"net/http"
	"github.com/wayne011872/goSterna/notify"
	goSterna "github.com/wayne011872/goSterna"
	apiErr "github.com/wayne011872/goSterna/api/err"
	"github.com/gin-gonic/gin"
)

type MailMidDI interface {
	notify.MailDI
}

type MailMiddle string

func NewGinMailMid(service string) GinMiddle {
	return &mailMiddle{
		service: service,
	}
}

type mailMiddle struct {
	service string
}

func (mm *mailMiddle) GetName() string{
	return "mail"
}

func(mm *mailMiddle) outputErr(c *gin.Context, err error) {
	apiErr.GinOutputErr(c,mm.service,err)
}

func (mm *mailMiddle) Handler() gin.HandlerFunc{
	return func(c *gin.Context) {
		servDi, ok := c.Get(string(goSterna.CtxServDiKey))
		if !ok || servDi == nil {
			mm.outputErr(c,apiErr.NewApiError(http.StatusInternalServerError, "can not get di"))
			c.Abort()
			return
		}
		if maildi,ok := servDi.(MailMidDI); ok {
			m := maildi.NewMail()
			c.Set(string(notify.CtxMailKey),m)
			c.Next()
			runtime.GC()
		}else{
			mm.outputErr(c,apiErr.NewApiError(http.StatusInternalServerError, "invalid di"))
			c.Abort()
			return
		}
	}
}