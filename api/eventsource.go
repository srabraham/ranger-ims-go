package api

import (
	"encoding/json"
	"github.com/launchdarkly/eventsource"
	"log/slog"
	"strconv"
	"sync/atomic"
)

var idCounter atomic.Int64

type IMSEventData struct {
	EventName         string `json:"event_id,omitzero"`
	IncidentNumber    int32  `json:"incident_number,omitzero"`
	FieldReportNumber int32  `json:"field_report_number,omitzero"`
	InitialEvent      bool   `json:"initial_event,omitzero"`
	Comment           string `json:"comment,omitzero"`
}

type IMSEvent struct {
	EventID   int64
	EventData IMSEventData
}

func (e IMSEvent) Id() string {
	return strconv.FormatInt(e.EventID, 10)
}

func (e IMSEvent) Event() string {
	if e.EventData.IncidentNumber > 0 {
		return "Incident"
	}
	if e.EventData.FieldReportNumber > 0 {
		return "FieldReport"
	}
	if e.EventData.InitialEvent {
		return "InitialEvent"
	}
	return "UnknownEvent"
}

func (e IMSEvent) Data() string {
	b, err := json.Marshal(e.EventData)
	slog.Info("in eventsource data", "err", err, "str", string(b), "id", e.EventID)
	return string(b)
}

type initialEventRepository struct {
}

func (r initialEventRepository) Replay(channel, id string) chan eventsource.Event {
	if channel != "imsevents" {
		return nil
	}
	out := make(chan eventsource.Event, 1)
	out <- IMSEvent{
		EventID: idCounter.Load(),
		EventData: IMSEventData{
			InitialEvent: true,
			Comment:      "The most recent SSE ID is provided in this message",
		},
	}
	close(out)
	return out
}
