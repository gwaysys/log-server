package mail

import (
	"testing"

	"gopkg.in/gomail.v2"
)

func TestMailClient(t *testing.T) {
	host := "smtp.lib10.cn"
	userName := "anonymous@lib10.cn"
	userPwd := "anonymous" // 测试时请填写真实密码
	msg := gomail.NewMessage()
	msg.SetHeader("From", userName)
	msg.SetHeader("To", "testing@lib10.cn")
	msg.SetHeader("Subject", "Hello!")
	msg.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")

	//mc := NewMailClient(host, 587, userName, userPwd)
	mc := NewMailClient(host, 465, userName, userPwd)
	defer mc.Close()
	if err := mc.SendMail(msg); err != nil {
		t.Fatal(err)
	}
}
