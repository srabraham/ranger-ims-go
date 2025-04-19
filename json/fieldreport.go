package json

import "time"

type FieldReport struct {
	Event         *string        `json:"event"`
	Number        *int32         `json:"number"`
	Created       *time.Time     `json:"created"`
	Summary       *string        `json:"summary"`
	Incident      *int32         `json:"incident"`
	ReportEntries *[]ReportEntry `json:"report_entries"`
}
