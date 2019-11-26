package send

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mail2ics/clean"
	"mail2ics/config"
	"mail2ics/toics"
	"net"
	"net/smtp"
)

// refernce from https://blog.csdn.net/xcl168/article/details/51340272
// and https://help.aliyun.com/document_detail/29457.html?spm=a2c4g.11186623.6.645.10001f23jBAnkR
func SendEmail(msg *clean.Message) error {
	if err := toics.ToIcs(msg); err != nil {
		return err
	}

	to := msg.From
	host := config.Sender.Addr
	port := 465
	from := config.Sender.Email
	password := config.Sender.Password
	//sendTo := strings.Split(to, ";")
	subject := msg.Subject
	boundary := "boundary" // boundary 用于分割邮件内容，可自定义. 注意它的开始和结束格式

	mime := bytes.NewBuffer(nil)

	// 设置邮件
	mime.WriteString(fmt.Sprintf(
		"From: %s<%s>\nTo: %s\nSubject: %s\nMIME-Version: 1.0\n",
		from, from, to, subject))
	mime.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", boundary))
	mime.WriteString("Content-Description: This is an email with .ics formate attachment\n")

	// 邮件普通Text正文
	mime.WriteString(fmt.Sprintf("--%s\n", boundary))
	mime.WriteString("Content-Type: text/plain; charset=utf-8\n")
	mime.WriteString("This is an email with .ics formate attachment，" +
		"witch can be auto recognized as an activity by google calendar.")

	// 附件
	mime.WriteString(fmt.Sprintf("\n--%s\n", boundary))
	mime.WriteString("Content-Type: application/octet-stream\n")
	mime.WriteString("Content-Description: ics formate attachment\n")
	mime.WriteString("Content-Transfer-Encoding: base64\n")
	mime.WriteString("Content-Disposition: attachment; filename=\"" + msg.Filename + "\"\n\n")
	// 读取并编码文件内容
	attaData, err := ioutil.ReadFile(msg.Filename)
	if err != nil {
		return err
	}
	b := make([]byte, base64.StdEncoding.EncodedLen(len(attaData)))
	base64.StdEncoding.Encode(b, attaData)
	mime.Write(b)

	// 邮件结束
	mime.WriteString("\n--" + boundary + "--\n\n")

	// 发送相关
	auth := smtp.PlainAuth("", from, password, host)
	if err = SendMailUsingTLS(
		fmt.Sprintf("%s:%d", host, port),
		auth, from, []string{to}, mime.Bytes()); err != nil {
		return err
	}

	return nil
}

//return a smtp client
func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}
	//分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

//参考net/smtp的func SendMail()
//使用net.Dial连接tls(ssl)端口时,smtp.NewClient()会卡住且不提示err
//len(to)>1时,to[1]开始提示是密送
func SendMailUsingTLS(addr string, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {
	//create smtp client
	c, err := Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
