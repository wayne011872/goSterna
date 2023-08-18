package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wayne011872/goSterna/util"
)

const(
	CtxLineKey = util.CtxKey("ctxLineKey")
)

func GetLineByGin(c *gin.Context) Line {
	cltInter ,_:= c.Get(string(CtxLineKey))
	if line, ok := cltInter.(Line); ok {
		return line
	}
	return nil
}

type Line interface {
	SetLineBody(string)
	SendLine()
}

type LineDI interface {
	NewLine() Line
}

type LineConf struct {
	Url 		string 	`yaml:"notifyUrl"`
	AccessToken string	`yaml:"token"`
}

func (lc *LineConf) NewLine() Line{
	return &lineImpl{
		Url: lc.Url,
		AccessToken: lc.AccessToken,
	}
}

type lineImpl struct {
	Body 		*RequestBody
	Url 		string
	AccessToken string
}

func (li *lineImpl) SetLineBody(lineMessage string) {
	li.Body = &RequestBody{
		Message: lineMessage,
	}
}

func (li *lineImpl) SendLine(){
	jSysInfo, err := json.Marshal(li.Body)
	if err != nil {
		panic(err)
	}
	req,_:=http.NewRequest("POST",li.Url,bytes.NewReader(jSysInfo))
	req.Header.Set("Authorization",li.AccessToken)
	resp,err := (&http.Client{}).Do(req)
	fmt.Println(resp)
	if err != nil {
		panic(err)
	}
}

type RequestBody struct {
	Message string
}