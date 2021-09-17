package report

import (
	"time"
)

const(
	ReportCollectionName  = "reports" 
)

type Report struct {
	ID                string		`bson:"_id,omitempty" json:"_id,omitempty"`
	ReporterName      string        `bson:"reporter_name" validate:"required,min=2,max=100" json:"reporter_name"`
	ReporteeName      string        `bson:"reportee_name" validate:"required,min=2,max=100" json:"reportee_name"`
	Organization      string        `bson:"organization_id" validate:"required" json:"organization_id"`
	Date              time.Time     `bson:"date" json:"date" validate:"required"`
	Anonymous         bool          `bson:"anonymous" default:"false" json:"anonymous"`
	Subject           string        `bson:"subject" json:"subject"`
	Body              string        `bson:"body" json:"body"`
}