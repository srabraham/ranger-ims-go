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

const (
	IncidentPriorityHigh   = 5
	IncidentPriorityNormal = 3
	IncidentPriorityLow    = 1
)

type Incident struct {
	Event         string        `json:"event"`
	Number        int32         `json:"number"`
	Created       time.Time     `json:"created,omitzero"`
	LastModified  time.Time     `json:"last_modified,omitzero"`
	State         string        `json:"state"`
	Priority      int8          `json:"priority"`
	Summary       *string       `json:"summary"`
	Location      Location      `json:"location"`
	IncidentTypes IncidentTypes `json:"incident_types"`
	FieldReports  []int32       `json:"field_reports"`
	RangerHandles []string      `json:"ranger_handles"`
	ReportEntries []ReportEntry `json:"report_entries"`
}
