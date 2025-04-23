package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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
			//ID:            ptr(re.ID),
			//Created:       ptr(time.Unix(int64(re.Created), 0)),
			//Author:        ptr(re.Author),
			//SystemEntry:   ptr(re.Generated),
			//Text:          ptr(re.Text),
			//Stricken:      ptr(re.Stricken),
			//HasAttachment: ptr(re.AttachedFile.String != ""),
			ID:            re.ID,
			Created:       time.Unix(int64(re.Created), 0),
			Author:        re.Author,
			SystemEntry:   re.Generated,
			Text:          re.Text,
			Stricken:      re.Stricken,
			HasAttachment: re.AttachedFile.String != "",
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
			Event:   event.Name,
			Number:  r.Incident.Number,
			Created: time.Unix(int64(r.Incident.Created), 0),
			// TODO: should look at report entries too
			LastModified: time.Unix(int64(r.Incident.Created), 0),
			State:        string(r.Incident.State),
			Priority:     r.Incident.Priority,
			Summary:      stringOrNil(r.Incident.Summary),
			Location: imsjson.Location{
				Name:         stringOrNil(r.Incident.LocationName),
				Concentric:   stringOrNil(r.Incident.LocationConcentric),
				RadialHour:   int16OrNil(r.Incident.LocationRadialHour),
				RadialMinute: int16OrNil(r.Incident.LocationRadialMinute),
				Description:  stringOrNil(r.Incident.LocationDescription),
				Type:         &garett,
			},
			IncidentTypes: incidentTypes,
			FieldReports:  fieldReportNumbers,
			RangerHandles: rangerHandles,
			ReportEntries: entriesByIncident[r.Incident.Number],
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
				ID:            re.ID,
				Created:       time.Unix(int64(re.Created), 0),
				Author:        re.Author,
				SystemEntry:   re.Generated,
				Text:          re.Text,
				Stricken:      re.Stricken,
				HasAttachment: re.AttachedFile.String != "",
				//ID:            ptr(re.ID),
				//Created:       ptr(time.Unix(int64(re.Created), 0)),
				//Author:        ptr(re.Author),
				//SystemEntry:   ptr(re.Generated),
				//Text:          ptr(re.Text),
				//Stricken:      ptr(re.Stricken),
				//HasAttachment: ptr(re.AttachedFile.String != ""),
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
		Event:   eventRow.Event.Name,
		Number:  r.Incident.Number,
		Created: time.Unix(int64(r.Incident.Created), 0),
		// TODO: should look at report entries too
		LastModified: time.Unix(int64(r.Incident.Created), 0),
		State:        string(r.Incident.State),
		Priority:     r.Incident.Priority,
		Summary:      stringOrNil(r.Incident.Summary),
		Location: imsjson.Location{
			Name:         stringOrNil(r.Incident.LocationName),
			Concentric:   stringOrNil(r.Incident.LocationConcentric),
			RadialHour:   int16OrNil(r.Incident.LocationRadialHour),
			RadialMinute: int16OrNil(r.Incident.LocationRadialMinute),
			Description:  stringOrNil(r.Incident.LocationDescription),
			Type:         &garett,
		},
		IncidentTypes: incidentTypes,
		FieldReports:  fieldReportNumbers,
		RangerHandles: rangerHandles,
		ReportEntries: resultEntries,
	}

	writeJSON(w, result)
}

func addIncidentReportEntry(ctx context.Context, q *imsdb.Queries, eventID, incidentNum int32, author, text string, generated bool) error {
	reID, err := q.CreateReportEntry(ctx, imsdb.CreateReportEntryParams{
		Author:       author,
		Text:         text,
		Created:      float64(time.Now().Unix()),
		Generated:    generated,
		Stricken:     false,
		AttachedFile: sql.NullString{},
	})
	if err != nil {
		return fmt.Errorf("[CreateReportEntry]: %w", err)
	}
	err = q.AttachReportEntryToIncident(ctx, imsdb.AttachReportEntryToIncidentParams{
		Event:          eventID,
		IncidentNumber: incidentNum,
		ReportEntry:    int32(reID),
	})
	if err != nil {
		return fmt.Errorf("[AttachReportEntryToIncident]: %w", err)
	}
	return nil
}

type NewIncident struct {
	imsDB *sql.DB
	es    *EventSourcerer
}

func (action NewIncident) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	event, ok := eventFromName(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}
	newIncident, ok := readBodyAs[imsjson.Incident](w, req)
	if !ok {
		return
	}

	jwtCtx := req.Context().Value(JWTContextKey).(JWTContext)
	author := jwtCtx.Claims.RangerHandle()

	numUntyped, _ := imsdb.New(action.imsDB).MaxIncidentNumber(ctx, event.ID)
	newIncidentNum := numUntyped.(int64) + 1

	txn, _ := action.imsDB.Begin()
	defer txn.Rollback()
	dbTX := imsdb.New(action.imsDB).WithTx(txn)

	createIncident := imsdb.CreateIncidentParams{
		Event:                event.ID,
		Number:               int32(newIncidentNum),
		Created:              float64(time.Now().Unix()),
		Priority:             imsjson.IncidentPriorityNormal,
		State:                imsdb.IncidentStateNew,
		Summary:              sql.NullString{},
		LocationName:         sql.NullString{},
		LocationConcentric:   sql.NullString{},
		LocationRadialHour:   sql.NullInt16{},
		LocationRadialMinute: sql.NullInt16{},
		LocationDescription:  sql.NullString{},
	}

	var logs []string

	if newIncident.Priority != imsjson.IncidentPriorityNormal {
		createIncident.Priority = newIncident.Priority
		logs = append(logs, fmt.Sprintf("priority: %v", newIncident.Priority))
	}
	if newIncident.State != string(imsdb.IncidentStateNew) {
		createIncident.State = imsdb.IncidentState(newIncident.State)
		logs = append(logs, fmt.Sprintf("state: %v", newIncident.State))
	}
	if newIncident.Summary != nil {
		createIncident.Summary = sqlNullString(newIncident.Summary)
		logs = append(logs, fmt.Sprintf("summary: %v", *newIncident.Summary))
	}
	if newIncident.Location.Name != nil {
		createIncident.LocationName = sqlNullString(newIncident.Location.Name)
		logs = append(logs, fmt.Sprintf("location name: %v", *newIncident.Location.Name))
	}
	if newIncident.Location.Concentric != nil {
		createIncident.LocationConcentric = sqlNullString(newIncident.Location.Concentric)
		logs = append(logs, fmt.Sprintf("location concentric: %v", *newIncident.Location.Concentric))
	}
	if newIncident.Location.RadialHour != nil {
		createIncident.LocationRadialHour = sql.NullInt16{Int16: *newIncident.Location.RadialHour, Valid: true}
		logs = append(logs, fmt.Sprintf("location radial hour: %v", *newIncident.Location.RadialHour))
	}
	if newIncident.Location.RadialMinute != nil {
		createIncident.LocationRadialMinute = sql.NullInt16{Int16: *newIncident.Location.RadialMinute, Valid: true}
		logs = append(logs, fmt.Sprintf("location radial minute: %v", *newIncident.Location.RadialMinute))
	}
	if newIncident.Location.Description != nil {
		createIncident.LocationDescription = sqlNullString(newIncident.Location.Description)
		logs = append(logs, fmt.Sprintf("location description: %v", *newIncident.Location.Description))
	}
	_, err := dbTX.CreateIncident(ctx, createIncident)
	if err != nil {
		slog.Error("Failed to create incident", "error", err)
		http.Error(w, "Failed to create incident", http.StatusInternalServerError)
		return
	}

	if len(logs) > 0 {
		err = addIncidentReportEntry(ctx, dbTX, event.ID, int32(newIncidentNum), author, fmt.Sprintf("Changed %v", strings.Join(logs, ", ")), true)
		if err != nil {
			slog.Error("Error adding incident ReportEntry", "error", err)
			http.Error(w, "Error adding incident ReportEntry", http.StatusInternalServerError)
			return
		}
	}

	for _, entry := range newIncident.ReportEntries {
		if entry.Text == "" {
			continue
		}
		err = addIncidentReportEntry(ctx, dbTX, event.ID, int32(newIncidentNum), author, entry.Text, false)
		if err != nil {
			slog.Error("Error adding incident ReportEntry", "error", err)
			http.Error(w, "Error adding incident ReportEntry", http.StatusInternalServerError)
			return
		}
	}

	// TODO: incident types, ranger handles, more?

	if err = txn.Commit(); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-IMS-Incident-Number", strconv.FormatInt(newIncidentNum, 10))
	w.Header().Set("Location", "/ims/api/events/"+event.Name+"/incidents/"+strconv.FormatInt(newIncidentNum, 10))
	http.Error(w, http.StatusText(http.StatusCreated), http.StatusCreated)

	action.es.notifyIncidentUpdate(event.Name, int32(newIncidentNum))
}
