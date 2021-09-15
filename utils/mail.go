package utils

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type MailService interface {
	SetupSmtp(mailReq *Mail) ([]byte, error)
	SetupSendgrid(mailReq *Mail) ([]byte, error)
	SendMail(mailReq *Mail) error
	NewMail(from string, to []string, subject string, mailType MailType, data *MailData) *Mail
}

type MailType int

const (
	MailConfirmation MailType = iota + 1
	PasswordReset
)

type MailData struct {
	Username string
	Code	 string
}

type Mail struct {
	from  	string
	to    	[]string
	subject string
	body 	string
	mtype 	MailType
	data 	*MailData
}

type ZcMailService struct {
	configs		*Configurations
}

func NewZcMailService(c *Configurations) *ZcMailService {
	return &ZcMailService{ configs: c}
}

// Gmail smtp setup
// To use this, Gmail need to set allowed unsafe app
func (ms *ZcMailService) SetupSmtp(mailReq *Mail) ([]byte, error) {
	var templateFileName string

	if mailReq.mtype == MailConfirmation {
		templateFileName = ms.configs.ConfirmEmailTemplate
	} else if mailReq.mtype == PasswordReset {
		templateFileName = ms.configs.PasswordResetTemplate
	}

	t, err := template.ParseFiles(templateFileName)
	if err != nil {return nil, err }

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, mailReq.data); err != nil {}

	return []byte(buf.String()), nil
}

// sendgrid setup
func (ms *ZcMailService) SetupSendgrid(mailReq *Mail) ([]byte, error) {
	var templateFileName string

	if mailReq.mtype == MailConfirmation {
		templateFileName = ms.configs.ConfirmEmailTemplate
	} else if mailReq.mtype == PasswordReset {
		templateFileName = ms.configs.PasswordResetTemplate
	}

	t, err := template.ParseFiles(templateFileName)
	if err != nil { return nil, err }

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, mailReq.data); err != nil {
		return nil, err
	}	

	x, a := mailReq.to[0], mailReq.to[1:]

	from := mail.NewEmail("name", mailReq.from)
	to := mail.NewEmail("name", x)

	content := mail.NewContent("text/html", buf.String())

	m := mail.NewV3MailInit(from, mailReq.subject, to, content)
	if len(a) > 0 {

		tos := make([]*mail.Email, 0)
		for _, to := range mailReq.to {
			tos = append(tos, mail.NewEmail("user", to))
		}
				
		m.Personalizations[0].AddTos(tos...)
	}
	return mail.GetRequestBody(m), nil
}

func (ms *ZcMailService) SendMail(mailReq *Mail) error {
	// if ms.configs.ESPType == "sendgrid"
	switch esp := ms.configs.ESPType; esp {

	case "sendgrid":
		request := sendgrid.GetRequest(
			ms.configs.SendGridApiKey, 
			"/v3/mail/send", 
			"https://api.sendgrid.com",
		)

		request.Method = "POST"
		
		body, err := ms.SetupSendgrid(mailReq)
		request.Body = body

		response, err := sendgrid.API(request)
		if err != nil { return err }

		fmt.Printf("mail sent successfully, with status code %d", response.StatusCode)
		return nil

	case "smtp":
		auth := smtp.PlainAuth(
			"",
			ms.configs.SmtpUsername,
			ms.configs.SmtpPassword,
			"smtp.gmail.com",
		)

		body, err  := ms.SetupSmtp(mailReq)
		if err != nil { return err }

		addr := "smtp.gmail.com:587"
		if err := smtp.SendMail(addr, auth, mailReq.from, mailReq.to, body); err != nil {
			return err
		}
		return nil
		
	default:
		msg := fmt.Sprintf("%s is not included in the list of email service providers", ms.configs.ESPType)
		return errors.New(msg)
	}

}

func (ms *ZcMailService) NewMail(from string, to []string, subject string, mailType MailType, data *MailData) *Mail {
	return &Mail{
		from: 		from,
		to: 		to,
		subject: 	subject,
		mtype: 		mailType,
		data: 		data,
	}
}