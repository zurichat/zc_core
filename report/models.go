package report

import (
	"time"
)

const(
	ReportCollectionName  = "reports" 
)

type Report struct {
	ID                string		`bson:"_id,omitempty" json:"_id,omitempty"`
	ReporterEmail     string        `bson:"reporter_email" validate:"required,email" json:"reporter_email"`
	OffenderEmail     string        `bson:"offender_email" validate:"required,email" json:"offender_email"`
	Organization      string        `bson:"organization_id" validate:"required" json:"organization_id"`
	Date              time.Time     `bson:"date" json:"date" validate:"required"`
	Anonymous         bool          `bson:"anonymous" default:"false" json:"anonymous"`
	Subject           string        `bson:"subject" json:"subject"`
	Body              string        `bson:"body" json:"body"`
}