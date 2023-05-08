package mail

import (
	"crypto/tls"
	"net/smtp"
	"strconv"
	"errors"
	"github.com/jordan-wright/email"
	"github.com/spf13/viper"
)

type Mail struct {
	Auth smtp.Auth
	UserName string
	Server string
	Port int
	IsSSL bool
	FromMailAccount string
	Password string
	ToMailAccount []string
	Email *email.Email
}
func (m *Mail) MailInit() {
	m.GetConfig()
	m.SetMailMessage()
	m.SetAuth()
}

func (m *Mail) GetConfig(){
	
	vip:=viper.New()
	vip.AddConfigPath("./conf")
	vip.SetConfigName("config.yml")
	vip.SetConfigType("yaml")
	err := vip.ReadInConfig()
	if err != nil {
		panic("讀取設定檔出現錯誤，原因為：" + err.Error())
	}
	m.Server = vip.GetString("mail.server")
	m.Port = vip.GetInt("mail.port")
	m.IsSSL = vip.GetBool("mail.ssl")
	m.UserName = vip.GetString("mail.username")
	m.FromMailAccount = vip.GetString("mail.from")
	m.Password = vip.GetString("mail.password")
	m.ToMailAccount = vip.GetStringSlice("mail.to")
}

func (m *Mail) SetMailMessage() {
	m.Email = email.NewEmail()
	m.Email.From = m.FromMailAccount
	m.Email.To = m.ToMailAccount
}

func (m *Mail) SetMailTitle(mailTitle string) {
	m.Email.Subject = mailTitle
}

func (m *Mail) SetMailBody(mailBody string) {
	m.Email.HTML = []byte(mailBody)
}

func(m *Mail) SetAuth() {
	m.Auth = LoginAuth(m.UserName,m.Password)
}

func (m *Mail) SendMail() {
	if m.IsSSL {
		if err := m.Email.SendWithTLS(m.Server+":"+ strconv.Itoa(m.Port),m.Auth,&tls.Config{ServerName: m.Server}); err != nil {
			panic("寄信時發生錯誤，原因為：" + err.Error())
		}
	}else{
		if err := m.Email.Send(m.Server+":"+ strconv.Itoa(m.Port),m.Auth); err != nil {
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