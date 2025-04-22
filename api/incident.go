package api

import (
	"database/sql"
	"encoding/json"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type GetIncidents struct {
	imsDB *sql.DB
}

func (action GetIncidents) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !parseForm(w, req) {
		return
	}
	generatedLTE := req.Form.Get("exclude_system_entries") != "true" // false means to exclude

	event, ok := eventFromName(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}

	//eventName := req.PathValue("eventName")
	//
	//eventRow, err := queries.New(action.imsDB).QueryEventID(req.Context(), eventName)
	//if err != nil {
	//	slog.Error("Failed to get event ID", "error", err)
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}

	reportEntries, err := imsdb.New(action.imsDB).Incidents_ReportEntries(req.Context(),
		imsdb.Incidents_ReportEntriesParams{
			Event:     event.ID,
			Generated: generatedLTE,
		})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	entriesByIncident := make(map[int32][]imsjson.ReportEntry)
	for _, row := range reportEntries {
		re := row.ReportEntry
		entriesByIncident[row.IncidentNumber] = append(entriesByIncident[row.IncidentNumber], imsjson.ReportEntry{
			ID:            ptr(re.ID),
			Created:       ptr(time.Unix(int64(re.Created), 0)),
			Author:        ptr(re.Author),
			SystemEntry:   ptr(re.Generated),
			Text:          ptr(re.Text),
			Stricken:      ptr(re.Stricken),
			HasAttachment: ptr(re.AttachedFile.String != ""),
		})
	}

	rows, err := imsdb.New(action.imsDB).Incidents(req.Context(), event.ID)
	if err != nil {
		log.Println(err)
		return
	}

	resp := make(imsjson.Incidents, 0)

	var garett = "garett"
	for _, r := range rows {
		var incidentTypes imsjson.IncidentTypes
		var rangerHandles []string
		var fieldReportNumbers []int32
		json.Unmarshal(r.IncidentTypes.([]byte), &incidentTypes)
		json.Unmarshal(r.RangerHandles.([]byte), &rangerHandles)
		json.Unmarshal(r.FieldReportNumbers.([]byte), &fieldReportNumbers)
		resp = append(resp, imsjson.Incident{
			Event:   ptr(event.Name),
			Number:  ptr(r.Incident.Number),
			Created: ptr(time.Unix(int64(r.Incident.Created), 0)),
			// TODO: should look at report entries too
			LastModified: ptr(time.Unix(int64(r.Incident.Created), 0)),
			State:        ptr(string(r.Incident.State)),
			Priority:     ptr(r.Incident.Priority),
			Summary:      stringOrNil(r.Incident.Summary),
			Location: &imsjson.Location{
				Name:         stringOrNil(r.Incident.LocationName),
				Concentric:   stringOrNil(r.Incident.LocationConcentric),
				RadialHour:   int16OrNil(r.Incident.LocationRadialHour),
				RadialMinute: int16OrNil(r.Incident.LocationRadialMinute),
				Description:  stringOrNil(r.Incident.LocationDescription),
				Type:         &garett,
			},
			IncidentTypes: ptr(incidentTypes),
			FieldReports:  ptr(fieldReportNumbers),
			RangerHandles: ptr(rangerHandles),
			ReportEntries: ptr(entriesByIncident[r.Incident.Number]),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	jjj, _ := json.Marshal(resp)
	w.Write(jjj)
}

type GetIncident struct {
	imsDB *sql.DB
}

func (action GetIncident) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	eventName := req.PathValue("eventName")
	incident := req.PathValue("incidentNumber")

	eventRow, err := imsdb.New(action.imsDB).QueryEventID(req.Context(), eventName)
	if err != nil {
		slog.Error("Failed to get eventRow ID", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	incidentNumber, err := strconv.ParseInt(incident, 10, 32)
	if err != nil {
		slog.Error("Got nonnumeric incident number", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r, err := imsdb.New(action.imsDB).Incident(req.Context(), imsdb.IncidentParams{
		Event:  eventRow.Event.ID,
		Number: int32(incidentNumber),
	})
	if err != nil {
		slog.Error("Failed to read incident", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	reportEntries, err := imsdb.New(action.imsDB).Incident_ReportEntries(req.Context(),
		imsdb.Incident_ReportEntriesParams{
			Event:          eventRow.Event.ID,
			IncidentNumber: int32(incidentNumber),
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	resultEntries := make([]imsjson.ReportEntry, 0)
	for _, entry := range reportEntries {
		re := entry.ReportEntry
		resultEntries = append(resultEntries,
			imsjson.ReportEntry{
				ID:            ptr(re.ID),
				Created:       ptr(time.Unix(int64(re.Created), 0)),
				Author:        ptr(re.Author),
				SystemEntry:   ptr(re.Generated),
				Text:          ptr(re.Text),
				Stricken:      ptr(re.Stricken),
				HasAttachment: ptr(re.AttachedFile.String != ""),
			},
		)
	}

	var incidentTypes imsjson.IncidentTypes
	var rangerHandles []string
	var fieldReportNumbers []int32
	json.Unmarshal(r.IncidentTypes.([]byte), &incidentTypes)
	json.Unmarshal(r.RangerHandles.([]byte), &rangerHandles)
	json.Unmarshal(r.FieldReportNumbers.([]byte), &fieldReportNumbers)

	var garett = "garett"
	result := imsjson.Incident{
		Event:   ptr(eventRow.Event.Name),
		Number:  ptr(r.Incident.Number),
		Created: ptr(time.Unix(int64(r.Incident.Created), 0)),
		// TODO: should look at report entries too
		LastModified: ptr(time.Unix(int64(r.Incident.Created), 0)),
		State:        ptr(string(r.Incident.State)),
		Priority:     ptr(r.Incident.Priority),
		Summary:      stringOrNil(r.Incident.Summary),
		Location: &imsjson.Location{
			Name:         stringOrNil(r.Incident.LocationName),
			Concentric:   stringOrNil(r.Incident.LocationConcentric),
			RadialHour:   int16OrNil(r.Incident.LocationRadialHour),
			RadialMinute: int16OrNil(r.Incident.LocationRadialMinute),
			Description:  stringOrNil(r.Incident.LocationDescription),
			Type:         &garett,
		},
		IncidentTypes: ptr(incidentTypes),
		FieldReports:  ptr(fieldReportNumbers),
		RangerHandles: ptr(rangerHandles),
		ReportEntries: ptr(resultEntries),
	}

	j, _ := json.Marshal(result)

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
