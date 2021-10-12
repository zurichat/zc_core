package report

import (
	"time"

	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

const (
	ReportCollectionName       = "reports"
	OrganizationCollectionName = "organizations"
	MemberCollectionName       = "members"
)

type Report struct {
	ID            string    `bson:"_id,omitempty" json:"_id,omitempty"`
	ReporterEmail string    `bson:"reporter_email" validate:"required,email" json:"reporter_email"`
	OffenderEmail string    `bson:"offender_email" validate:"required,email" json:"offender_email"`
	Organization  string    `bson:"organization_id" validate:"required" json:"organization_id"`
	Date          time.Time `bson:"date" json:"date" validate:"required"`
	Anonymous     bool      `bson:"anonymous" default:"false" json:"anonymous"`
	Subject       string    `bson:"subject" json:"subject"`
	Body          string    `bson:"body" json:"body"`
}


type ReportHandler struct {
	configs     *utils.Configurations
	mailService service.MailService
}

func NewReportHandler(c *utils.Configurations, mail service.MailService) *ReportHandler {
	return &ReportHandler{configs: c, mailService: mail}
}
