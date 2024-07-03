package notify

import (
	"crypto/tls"
	"errors"
	"net/smtp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jordan-wright/email"
	"github.com/wayne011872/goSterna/util"
)

const(
	CtxMailKey = util.CtxKey("ctxMailKey")
)

func GetMailByGin(c *gin.Context) Mail {
	cltInter ,_:= c.Get(string(CtxMailKey))
	if mail, ok := cltInter.(Mail); ok {
		return mail
	}
	return nil
}

type Mail interface {
	SetMailTitle(string)
	SetMailBody(string)
	SendMail()
}

type MailDI interface {
	NewMail()Mail
}

type MailConf struct {
	UserName string 		`yaml:"username"`
	Server string			`yaml:"server"`
	Port int				`yaml:"port"`
	IsSSL bool				`yaml:"ssl"`
	FromMailAccount string	`yaml:"from"`
	Password string			`yaml:"password"`
	ToMailAccount []string	`yaml:"to"`
}

func (mc *MailConf)NewMail()Mail {
	mail := email.NewEmail()
	mail.From = mc.FromMailAccount
	mail.To = mc.ToMailAccount
	auth := LoginAuth(mc.UserName,mc.Password)
	return &mailImpl {
		Auth: auth,
		Email: mail,
		Server: mc.Server,
		Port: mc.Port,
		IsSSL: mc.IsSSL,
	}
}

type mailImpl struct {
	Auth smtp.Auth
	Server string
	Port int
	IsSSL bool
	Email *email.Email
}

func (mi *mailImpl) SetMailTitle(mailTitle string) {
	mi.Email.Subject = mailTitle
}

func (mi *mailImpl) SetMailBody(mailBody string) {
	mi.Email.HTML = []byte(mailBody)
}

func (mi *mailImpl) SendMail() {
	if mi.IsSSL {
		if err := mi.Email.SendWithTLS(mi.Server+":"+ strconv.Itoa(mi.Port),mi.Auth,&tls.Config{ServerName: mi.Server}); err != nil {
			panic("寄信時發生錯誤，原因為：" + err.Error())
		}
	}else{
		if err := mi.Email.Send(mi.Server+":"+ strconv.Itoa(mi.Port),mi.Auth); err != nil {
			panic("寄信時發生錯誤，原因為：" + err.Error())
		}
	}
}

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown fromServer")
		}
	}
	return nil, nil
}