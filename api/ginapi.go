package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wayne011872/goSterna/api/mid"
)

type GinAPI interface {
	GetAPIs() []*GinApiHandler
	GetName() string
}

func NewGinApiServer(mode string) GinApiServer {
	gin.SetMode(mode)
	return &apiService{
		Engine: gin.New(),
	}
}

type GinApiHandler struct {
	Method   string
	Path     string
	Handler  func(c *gin.Context)
}

type GinApiServer interface {
	AddAPIs(handlers ...GinAPI)GinApiServer
	Middles(mids ...mid.GinMiddle) GinApiServer
	Run(port string) error
}

type apiService struct {
	*gin.Engine
}

func (serv *apiService) Middles(mids ...mid.GinMiddle) GinApiServer {
	for _, m := range mids {
		serv.Engine.Use(m.Handler())
	}

	return serv
}

func (serv *apiService) AddAPIs(apis ...GinAPI) GinApiServer {
	for _,api := range apis {
		for _,h := range api.GetAPIs() {
			switch h.Method {
			case "GET":
				serv.Engine.GET(h.Path,h.Handler)
			case "POST":
				serv.Engine.POST(h.Path,h.Handler)
			case "PUT":
				serv.Engine.PUT(h.Path,h.Handler)
			case "DELETE":
				serv.Engine.DELETE(h.Path,h.Handler)
			}
		}
	}
	return serv
}

func (serv *apiService) Run(port string) error {
	return serv.Engine.Run(":"+port)
}
func GinOutputErr(c *gin.Context, service string, err error) {
	if err == nil {
		return
	}
	if apiErr, ok := err.(ApiError); ok {
		c.AbortWithStatusJSON(apiErr.GetStatus(),
			map[string]interface{}{
				"status":   apiErr.GetStatus(),
				"title":    apiErr.GetErrorMsg(),
				"service":  service,
				"errorKey": apiErr.GetErrorKey(),
			})
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			map[string]interface{}{
				"status":   http.StatusInternalServerError,
				"title":    err.Error(),
				"service":  service,
				"errorKey": "",
			})
	}
}

type ErrorOutputAPI interface {
	GinOutputErr(c *gin.Context, err error)
}

func NewErrorOutputAPI(service string) ErrorOutputAPI {
	return &commonAPI{
		service: service,
	}
}

type commonAPI struct {
	service string
}

func (capi *commonAPI) GinOutputErr(c *gin.Context, err error) {
	GinOutputErr(c, capi.service, err)
}


type ApiError interface {
	GetStatus() int
	GetErrorKey() string
	GetErrorMsg() string
	error
}

type myApiError struct {
	statusCode int
	message    string
	key        string
}

func (e myApiError) GetStatus() int {
	return e.statusCode
}

func (e myApiError) GetErrorKey() string {
	return e.key
}

func (e myApiError) GetErrorMsg() string {
	return e.message
}

func (e myApiError) Error() string {
	return fmt.Sprintf("%v: %v", e.statusCode, e.message)
}

func NewApiError(status int, msg string) ApiError {
	return myApiError{statusCode: status, message: msg}
}