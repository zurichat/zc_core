package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"zuri.chat/zccore/utils"
)

type MailService interface {
	SetupSmtp(mailReq *Mail) (string, error)
	SetupSendgrid(mailReq *Mail) ([]byte, error)
	SendMail(mailReq *Mail) error
	NewMail(to []string, subject string, mailType MailType, data *MailData) *Mail
}

type MailType int

const (
	MailConfirmation MailType = iota + 1
	PasswordReset
	DownloadClient
	WorkspaceInvite
)

type MailData struct {
	Username   string
	Code       string
	OrgName    string
	InviteLink template.URL
}

type Mail struct {
	to      []string
	subject string
	body    string
	mtype   MailType
	data    *MailData
}

type ZcMailService struct {
	configs *utils.Configurations
}

func NewZcMailService(c *utils.Configurations) *ZcMailService {
	return &ZcMailService{configs: c}
}

// Gmail smtp setup
// To use this, Gmail need to set allowed unsafe app
func (ms *ZcMailService) SetupSmtp(mailReq *Mail) (string, error) {
	var templateFileName string

	if mailReq.mtype == MailConfirmation {
		templateFileName = ms.configs.ConfirmEmailTemplate
	} else if mailReq.mtype == PasswordReset {
		templateFileName = ms.configs.PasswordResetTemplate
	} else if mailReq.mtype == DownloadClient {
		templateFileName = ms.configs.DownloadClientTemplate
	} else if mailReq.mtype == WorkspaceInvite {
		templateFileName = ms.configs.WorkspaceInviteTemplate
	}

	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, mailReq.data); err != nil {
	}

	return buf.String(), nil
}

// sendgrid setup
func (ms *ZcMailService) SetupSendgrid(mailReq *Mail) ([]byte, error) {
	var templateFileName string

	if mailReq.mtype == MailConfirmation {
		templateFileName = ms.configs.ConfirmEmailTemplate
	} else if mailReq.mtype == PasswordReset {
		templateFileName = ms.configs.PasswordResetTemplate
	} else if mailReq.mtype == DownloadClient {
		templateFileName = ms.configs.DownloadClientTemplate
	} else if mailReq.mtype == WorkspaceInvite {
		templateFileName = ms.configs.WorkspaceInviteTemplate
	}

	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, mailReq.data); err != nil {
		return nil, err
	}

	x, a := mailReq.to[0], mailReq.to[1:]

	from := mail.NewEmail("name", ms.configs.SendgridEmail)
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
	switch esp := strings.ToLower(ms.configs.ESPType); esp {

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
		if err != nil {
			return err
		}

		fmt.Printf("mail sent successfully, with status code %d", response.StatusCode)
		return nil

	case "smtp":
		// switch to mailgun temp
		body, err := ms.SetupSmtp(mailReq)
		if err != nil {
			return err
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

func (ms *ZcMailService) NewMail(to []string, subject string, mailType MailType, data *MailData) *Mail {
	return &Mail{
		to:      to,
		subject: subject,
		mtype:   mailType,
		data:    data,
	}
}
