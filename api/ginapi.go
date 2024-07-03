package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wayne011872/goSterna/api/mid"
	"github.com/wayne011872/goSterna/auth"
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
	Auth    bool
	Group   []auth.UserPerm
}

type GinApiServer interface {
	AddAPIs(handlers ...GinAPI)GinApiServer
	Middles(mids ...mid.GinMiddle) GinApiServer
	SetAuth(authmid mid.AuthGinMidInter) GinApiServer
	Static(relativePath string, root string)GinApiServer
	Run(port string) error
}

type apiService struct {
	*gin.Engine
	authMid mid.AuthGinMidInter
}

func (serv *apiService) Static(relativePath, root string) GinApiServer {
	serv.Engine.Static(relativePath, root)
	return serv
}

func (serv *apiService) SetAuth(authMid mid.AuthGinMidInter) GinApiServer {
	serv.authMid = authMid
	return serv
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
			if serv.authMid != nil {
				serv.authMid.AddAuthPath(h.Path, h.Method, h.Auth, h.Group)
			}
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