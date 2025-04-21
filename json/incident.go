package json

import (
	"time"
)

type Incidents []Incident

type Location struct {
	Name         *string `json:"name"`
	Concentric   *string `json:"concentric"`
	RadialHour   *int16  `json:"radial_hour"`
	RadialMinute *int16  `json:"radial_minute"`
	Description  *string `json:"description"`
	Type         *string `json:"type"`
}

type ReportEntry struct {
	ID            *int32     `json:"id"`
	Created       *time.Time `json:"created"`
	Author        *string    `json:"author"`
	SystemEntry   *bool      `json:"system_entry"`
	Text          *string    `json:"text"`
	Stricken      *bool      `json:"stricken"`
	HasAttachment *bool      `json:"has_attachment"`
}

type Incident struct {
	Event         *string        `json:"event"`
	Number        *int32         `json:"number"`
	Created       *time.Time     `json:"created"`
	LastModified  *time.Time     `json:"last_modified"`
	State         *string        `json:"state"`
	Priority      *int8          `json:"priority"`
	Summary       *string        `json:"summary"`
	Location      *Location      `json:"location"`
	IncidentTypes *IncidentTypes `json:"incident_types"`
	FieldReports  *[]int32       `json:"field_reports"`
	RangerHandles *[]string      `json:"ranger_handles"`
	ReportEntries *[]ReportEntry `json:"report_entries"`
}
