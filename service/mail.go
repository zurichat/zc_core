package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"zuri.chat/zccore/utils"
)

type MailService interface {
	LoadTemplate(mailReq *Mail) (string, error)
	SendMail(mailReq *Mail) error
	NewCustomMail(to []string, subject string, mailBody string) *Mail
	NewMail(to []string, subject string, mailType MailType, data map[string]interface{}) *Mail
}

type MailType int

const (
	MailConfirmation MailType = iota + 1
	PasswordReset
	EmailSubscription
	DownloadClient
	WorkspaceInvite
	TokenBillingNotice
	WorkSpaceInvite
	WorkSpaceWelcome
)

var MailTypes = map[MailType]MailType{
	MailConfirmation:   MailConfirmation,
	PasswordReset:      PasswordReset,
	EmailSubscription:  EmailSubscription,
	DownloadClient:     DownloadClient,
	WorkspaceInvite:    WorkspaceInvite,
	TokenBillingNotice: TokenBillingNotice,
	WorkSpaceInvite: 	WorkSpaceInvite,
	WorkSpaceWelcome: 	WorkSpaceWelcome,
}

type Mail struct {
	to      		[]string
	subject 		string
	body    		string
	customTmpl		bool
	mtype  			MailType
	data    		map[string]interface{}
}

type ZcMailService struct {
	configs *utils.Configurations
}

func NewZcMailService(c *utils.Configurations) *ZcMailService {
	return &ZcMailService{configs: c}
}

// Gmail smtp setup
// To use this, Gmail need to set allowed unsafe app
func (ms *ZcMailService) LoadTemplate(mailReq *Mail) (string, error) {

	// include your email template here
	m := map[MailType]string{
		MailConfirmation:   ms.configs.ConfirmEmailTemplate,
		PasswordReset:      ms.configs.PasswordResetTemplate,
		EmailSubscription:  ms.configs.EmailSubscriptionTemplate,
		DownloadClient:     ms.configs.DownloadClientTemplate,
		WorkspaceInvite:    ms.configs.WorkspaceInviteTemplate,
		TokenBillingNotice: ms.configs.TokenBillingNoticeTemplate,
		WorkSpaceInvite:	ms.configs.WorkSpaceInviteTemplate,
		WorkSpaceWelcome:	ms.configs.WorkSpaceWelcomeTemplate,
	}

	templateFileName, ok := m[mailReq.mtype]
	if !ok {
		return "", errors.New("Invalid email type, email template does not exists!")
	}

	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, mailReq.data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (ms *ZcMailService) SendMail(mailReq *Mail) error {
	// if ms.configs.ESPType == "sendgrid"
	switch esp := strings.ToLower(ms.configs.ESPType); esp {

	case "sendgrid":
		// SENDGRID
		var body string
		var err error

		if body = mailReq.body; mailReq.customTmpl != true {
			body, err = ms.LoadTemplate(mailReq)
			if err != nil { return err }
		}

		request := sendgrid.GetRequest(
			ms.configs.SendGridApiKey,
			"/v3/mail/send",
			"https://api.sendgrid.com",
		)

		request.Method = "POST"

		x, a := mailReq.to[0], mailReq.to[1:]
		reziever := strings.Split(x, "@")

		from := mail.NewEmail("Zuri Chat", ms.configs.SendgridEmail)
		to := mail.NewEmail(reziever[0], x)

		content := mail.NewContent("text/html", body)

		m := mail.NewV3MailInit(from, mailReq.subject, to, content)
		if len(a) > 0 {

			tos := make([]*mail.Email, 0)
			for _, to := range mailReq.to {
				user := strings.Split(to, "@")
				tos = append(tos, mail.NewEmail(user[0], to))
			}

			m.Personalizations[0].AddTos(tos...)
		}

		request.Body = mail.GetRequestBody(m)

		response, err := sendgrid.API(request)
		if err != nil {
			return err
		}

		fmt.Printf("mail sent successfully, with status code %d", response.StatusCode)
		return nil

	case "smtp":
		// SMTP -> use gmail
		var body string
		var err error

		if body = mailReq.body; mailReq.customTmpl != true {
			body, err = ms.LoadTemplate(mailReq)
			if err != nil { return err }
		}

		auth := smtp.PlainAuth(
			"",
			ms.configs.SmtpUsername,
			ms.configs.SmtpPassword,
			"smtp.gmail.com",
		)

		subject := fmt.Sprintf("Subject: %s\n", mailReq.subject)
		mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		msg := []byte(subject + mime + body)

		addr := "smtp.gmail.com:587"
		if err := smtp.SendMail(addr, auth, ms.configs.SmtpUsername, mailReq.to, msg); err != nil {
			return err
		}
		return nil

	case "mailgun":
		// switch to mailgun temp
		var body string
		var err error

		if body = mailReq.body; mailReq.customTmpl != true {
			body, err = ms.LoadTemplate(mailReq)
			if err != nil { return err }
		}

		mg := mailgun.NewMailgun(ms.configs.MailGunDomain, ms.configs.MailGunKey)
		message := mg.NewMessage(ms.configs.MailGunSenderEmail, mailReq.subject, "", mailReq.to...)
		message.SetHtml(body)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if _, _, err := mg.Send(ctx, message); err != nil {
			return err
		}
		return nil

	default:
		msg := fmt.Sprintf("%s is not included in the list of email service providers", esp)
		return errors.New(msg)
	}

}

func (ms *ZcMailService) NewCustomMail(to []string, subject string, mailBody string) *Mail {
	return &Mail{
		to:      to,
		subject: subject,
		body:    mailBody,
		customTmpl:  true,
	}
}

func (ms *ZcMailService) NewMail(to []string, subject string, mailType MailType, data map[string]interface{}) *Mail {
	return &Mail{
		to:      to,
		subject: subject,
		mtype:   mailType,
		data:    data,
		customTmpl: false,
	}
}
