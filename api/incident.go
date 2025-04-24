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
	if !mustParseForm(w, req) {
		return
	}
	generatedLTE := req.Form.Get("exclude_system_entries") != "true" // false means to exclude

	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}

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
				Name:         ptr(r.Incident.LocationName.String),
				Concentric:   ptr(r.Incident.LocationConcentric.String),
				RadialHour:   formatInt16(r.Incident.LocationRadialHour),
				RadialMinute: formatInt16(r.Incident.LocationRadialMinute),
				Description:  ptr(r.Incident.LocationDescription.String),
				Type:         garett,
			},
			IncidentTypes: incidentTypes,
			FieldReports:  fieldReportNumbers,
			RangerHandles: rangerHandles,
			ReportEntries: entriesByIncident[r.Incident.Number],
		})
	}

	writeJSON(w, resp)
}

func formatInt16(i sql.NullInt16) *string {
	if i.Valid {
		result := strconv.FormatInt(int64(i.Int16), 10)
		return &result
	}
	return nil
}

func parseInt16(s *string) sql.NullInt16 {
	if s == nil {
		return sql.NullInt16{}
	}
	parsed, err := strconv.ParseInt(*s, 10, 16)
	if err != nil {
		return sql.NullInt16{}
	}
	return sql.NullInt16{
		Int16: int16(parsed),
		Valid: true,
	}
}

type GetIncident struct {
	imsDB *sql.DB
}

func (action GetIncident) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}
	incident := req.PathValue("incidentNumber")

	incidentNumber, err := strconv.ParseInt(incident, 10, 32)
	if err != nil {
		slog.Error("Got nonnumeric incident number", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r, err := imsdb.New(action.imsDB).Incident(req.Context(), imsdb.IncidentParams{
		Event:  event.ID,
		Number: int32(incidentNumber),
	})
	if err != nil {
		slog.Error("Failed to read incident", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	reportEntries, err := imsdb.New(action.imsDB).Incident_ReportEntries(req.Context(),
		imsdb.Incident_ReportEntriesParams{
			Event:          event.ID,
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
		Event:   event.Name,
		Number:  r.Incident.Number,
		Created: time.Unix(int64(r.Incident.Created), 0),
		// TODO: should look at report entries too
		LastModified: time.Unix(int64(r.Incident.Created), 0),
		State:        string(r.Incident.State),
		Priority:     r.Incident.Priority,
		Summary:      stringOrNil(r.Incident.Summary),
		Location: imsjson.Location{
			Name:         ptr(r.Incident.LocationName.String),
			Concentric:   ptr(r.Incident.LocationConcentric.String),
			RadialHour:   formatInt16(r.Incident.LocationRadialHour),
			RadialMinute: formatInt16(r.Incident.LocationRadialMinute),
			Description:  ptr(r.Incident.LocationDescription.String),
			Type:         garett,
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
	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}
	newIncident, ok := mustReadBodyAs[imsjson.Incident](w, req)
	if !ok {
		return
	}

	jwtCtx := req.Context().Value(JWTContextKey).(JWTContext)
	author := jwtCtx.Claims.RangerHandle()

	// First create the incident, to lock in the incident number reservation
	numUntyped, _ := imsdb.New(action.imsDB).MaxIncidentNumber(ctx, event.ID)
	newIncident.EventID = event.ID
	newIncident.Number = int32(numUntyped.(int64)) + 1

	createTheIncident := imsdb.CreateIncidentParams{
		Event:    newIncident.EventID,
		Number:   newIncident.Number,
		Created:  float64(time.Now().Unix()),
		Priority: imsjson.IncidentPriorityNormal,
		State:    imsdb.IncidentStateNew,
	}
	code, err := imsdb.New(action.imsDB).CreateIncident(ctx, createTheIncident)
	slog.Info("in newincident got", "code", code, "err", err)
	if err != nil {
		slog.Error("error creating incident", "err", err)
		http.Error(w, "Failed to create incident", http.StatusInternalServerError)
		return
	}

	if err = updateIncident(ctx, action.imsDB, newIncident, author); err != nil {
		slog.Error("error updating incident", "err", err)
		http.Error(w, "Failed to update incident", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-IMS-Incident-Number", fmt.Sprint(newIncident.Number))
	w.Header().Set("Location", "/ims/api/events/"+event.Name+"/incidents/"+fmt.Sprint(newIncident.Number))
	http.Error(w, http.StatusText(http.StatusCreated), http.StatusCreated)

	action.es.notifyIncidentUpdate(event.Name, newIncident.Number)
	for _, fr := range newIncident.FieldReports {
		action.es.notifyFieldReportUpdate(event.Name, fr)
	}
}

func updateIncident(ctx context.Context, imsDB *sql.DB, newIncident imsjson.Incident, author string) error {
	storedIncidentRow, err := imsdb.New(imsDB).Incident(ctx, imsdb.IncidentParams{
		Event:  newIncident.EventID,
		Number: newIncident.Number,
	})
	if err != nil {
		panic(err)
	}
	storedIncident := storedIncidentRow.Incident

	txn, _ := imsDB.Begin()
	defer txn.Rollback()
	dbTX := imsdb.New(imsDB).WithTx(txn)

	update := imsdb.UpdateIncidentParams{
		Event:                storedIncident.Event,
		Number:               storedIncident.Number,
		Created:              storedIncident.Created,
		Priority:             storedIncident.Priority,
		State:                storedIncident.State,
		Summary:              storedIncident.Summary,
		LocationName:         storedIncident.LocationName,
		LocationConcentric:   storedIncident.LocationConcentric,
		LocationRadialHour:   storedIncident.LocationRadialHour,
		LocationRadialMinute: storedIncident.LocationRadialMinute,
		LocationDescription:  storedIncident.LocationDescription,
	}

	var logs []string

	if newIncident.Priority != 0 {
		update.Priority = newIncident.Priority
		logs = append(logs, fmt.Sprintf("priority: %v", update.Priority))
	}
	if imsdb.IncidentState(newIncident.State).Valid() {
		update.State = imsdb.IncidentState(newIncident.State)
		logs = append(logs, fmt.Sprintf("state: %v", update.State))
	}
	if newIncident.Summary != nil {
		update.Summary = sqlNullString(newIncident.Summary)
		logs = append(logs, fmt.Sprintf("summary: %v", update.Summary.String))
	}
	if newIncident.Location.Name != nil {
		update.LocationName = sqlNullString(newIncident.Location.Name)
		logs = append(logs, fmt.Sprintf("location name: %v", update.LocationName.String))
	}
	if newIncident.Location.Concentric != nil {
		update.LocationConcentric = sqlNullString(newIncident.Location.Concentric)
		logs = append(logs, fmt.Sprintf("location concentric: %v", update.LocationConcentric.String))
	}
	if newIncident.Location.RadialHour != nil {
		update.LocationRadialHour = parseInt16(newIncident.Location.RadialHour)
		logs = append(logs, fmt.Sprintf("location radial hour: %v", update.LocationRadialHour.Int16))
	}
	if newIncident.Location.RadialMinute != nil {
		update.LocationRadialMinute = parseInt16(newIncident.Location.RadialMinute)
		logs = append(logs, fmt.Sprintf("location radial minute: %v", update.LocationRadialMinute.Int16))
	}
	if newIncident.Location.Description != nil {
		update.LocationDescription = sqlNullString(newIncident.Location.Description)
		logs = append(logs, fmt.Sprintf("location description: %v", update.LocationDescription.String))
	}
	err = dbTX.UpdateIncident(ctx, update)
	if err != nil {
		return fmt.Errorf("[UpdateIncident]: %w", err)
	}

	if len(newIncident.RangerHandles) > 0 {
		logs = append(logs, fmt.Sprintf("Rangers: %v", newIncident.RangerHandles))
		for _, rh := range newIncident.RangerHandles {
			err = dbTX.AttachRangerHandleToIncident(ctx, imsdb.AttachRangerHandleToIncidentParams{
				Event:          newIncident.EventID,
				IncidentNumber: newIncident.Number,
				RangerHandle:   rh,
			})
			if err != nil {
				return fmt.Errorf("[AttachRangerHandleToIncident]: %w", err)
			}
		}
	}

	if len(newIncident.IncidentTypes) > 0 {
		logs = append(logs, fmt.Sprintf("incident types: %v", newIncident.IncidentTypes))
		for _, itype := range newIncident.IncidentTypes {
			err = dbTX.AttachIncidentTypeToIncident(ctx, imsdb.AttachIncidentTypeToIncidentParams{
				Event:          newIncident.EventID,
				IncidentNumber: newIncident.Number,
				Name:           itype,
			})
			if err != nil {
				return fmt.Errorf("[AttachIncidentTypeToIncident]: %w", err)
			}
		}
	}

	if len(newIncident.FieldReports) > 0 {
		logs = append(logs, fmt.Sprintf("field reports: %v", newIncident.FieldReports))
		for _, fr := range newIncident.FieldReports {
			err = dbTX.AttachFieldReportToIncident(ctx, imsdb.AttachFieldReportToIncidentParams{
				Event:          newIncident.EventID,
				IncidentNumber: sql.NullInt32{Int32: int32(newIncident.Number), Valid: true},
				Number:         fr,
			})
			if err != nil {
				return fmt.Errorf("[AttachFieldReportToIncident]: %w", err)
			}
		}
	}

	if len(logs) > 0 {
		err = addIncidentReportEntry(ctx, dbTX, newIncident.EventID, newIncident.Number, author, fmt.Sprintf("Changed %v", strings.Join(logs, ", ")), true)
		if err != nil {
			return fmt.Errorf("[addIncidentReportEntry]: %w", err)
		}
	}

	for _, entry := range newIncident.ReportEntries {
		if entry.Text == "" {
			continue
		}
		err = addIncidentReportEntry(ctx, dbTX, newIncident.EventID, newIncident.Number, author, entry.Text, false)
		if err != nil {
			return fmt.Errorf("[addIncidentReportEntry]: %w", err)
		}
	}

	if err = txn.Commit(); err != nil {
		return fmt.Errorf("[Commit]: %w", err)
	}
	return nil
}

type EditIncident struct {
	imsDB *sql.DB
	es    *EventSourcerer
}

func (action EditIncident) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}
	incidentNumber, err := strconv.ParseInt(req.PathValue("incidentNumber"), 10, 32)
	if err != nil {
		slog.Error("Got nonnumeric incident number", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newIncident, ok := mustReadBodyAs[imsjson.Incident](w, req)
	if !ok {
		return
	}
	newIncident.EventID = event.ID
	newIncident.Number = int32(incidentNumber)

	jwtCtx := req.Context().Value(JWTContextKey).(JWTContext)
	author := jwtCtx.Claims.RangerHandle()

	//r, err := imsdb.New(action.imsDB).Incident(req.Context(), imsdb.IncidentParams{
	//	Event:  event.ID,
	//	Number: int32(incidentNumber),
	//})
	//if err != nil {
	//	slog.Error("Failed to read incident", "error", err)
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	//
	//var incidentTypes imsjson.IncidentTypes
	//var rangerHandles []string
	//var fieldReportNumbers []int32
	//json.Unmarshal(r.IncidentTypes.([]byte), &incidentTypes)
	//json.Unmarshal(r.RangerHandles.([]byte), &rangerHandles)
	//json.Unmarshal(r.FieldReportNumbers.([]byte), &fieldReportNumbers)

	if err = updateIncident(ctx, action.imsDB, newIncident, author); err != nil {
		slog.Error("error updating incident", "err", err)
		http.Error(w, "Failed to update incident", http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)

	action.es.notifyIncidentUpdate(event.Name, newIncident.Number)
	// TODO: need to update detached FRs too
	for _, fr := range newIncident.FieldReports {
		action.es.notifyFieldReportUpdate(event.Name, fr)
	}

	// TODO: much stuff...try not to repeat a ton of code from the "new" action
	return

	//	ctx := req.Context()
	//	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
	//	if !ok {
	//		return
	//	}
	//
	//	newIncident, ok := mustReadBodyAs[imsjson.Incident](w, req)
	//	if !ok {
	//		return
	//	}
	//
	//}
}
