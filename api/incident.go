package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/srabraham/ranger-ims-go/auth"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"golang.org/x/exp/slices"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type GetIncidents struct {
	imsDB *store.DB
}

func (action GetIncidents) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	perms, ok := mustGetPermissionsCtx(w, req)
	if !ok {
		return
	}
	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}
	if perms.EventPermissions[event.ID]&auth.EventReadIncidents == 0 {
		slog.Error("The requestor does not have EventReadIncidents permission on this Event")
		http.Error(w, "The requestor does not have EventReadIncidents permission on this Event", http.StatusForbidden)
		return
	}

	if !mustParseForm(w, req) {
		return
	}
	generatedLTE := req.Form.Get("exclude_system_entries") != "true" // false means to exclude

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
			IncidentTypes: &incidentTypes,
			FieldReports:  &fieldReportNumbers,
			RangerHandles: &rangerHandles,
			ReportEntries: entriesByIncident[r.Incident.Number],
		})
	}

	mustWriteJSON(w, resp)
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
	imsDB *store.DB
}

func (action GetIncident) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	perms, ok := mustGetPermissionsCtx(w, req)
	if !ok {
		return
	}
	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}
	if perms.EventPermissions[event.ID]&auth.EventReadIncidents == 0 {
		slog.Error("The requestor does not have EventReadIncidents permission on this Event")
		http.Error(w, "The requestor does not have EventReadIncidents permission on this Event", http.StatusForbidden)
		return
	}
	ctx := req.Context()

	incidentNumber, err := strconv.ParseInt(req.PathValue("incidentNumber"), 10, 32)
	if err != nil {
		slog.Error("Got nonnumeric incident number", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	storedRow, reportEntries, err := fetchIncident(ctx, action.imsDB, event.ID, int32(incidentNumber))
	if err != nil {
		slog.Error("Failed to read incident", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resultEntries := make([]imsjson.ReportEntry, 0)
	for _, re := range reportEntries {
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
	json.Unmarshal(storedRow.IncidentTypes.([]byte), &incidentTypes)
	json.Unmarshal(storedRow.RangerHandles.([]byte), &rangerHandles)
	json.Unmarshal(storedRow.FieldReportNumbers.([]byte), &fieldReportNumbers)

	var garett = "garett"
	result := imsjson.Incident{
		Event:   event.Name,
		Number:  storedRow.Incident.Number,
		Created: time.Unix(int64(storedRow.Incident.Created), 0),
		// TODO: should look at report entries too
		LastModified: time.Unix(int64(storedRow.Incident.Created), 0),
		State:        string(storedRow.Incident.State),
		Priority:     storedRow.Incident.Priority,
		Summary:      stringOrNil(storedRow.Incident.Summary),
		Location: imsjson.Location{
			Name:         ptr(storedRow.Incident.LocationName.String),
			Concentric:   ptr(storedRow.Incident.LocationConcentric.String),
			RadialHour:   formatInt16(storedRow.Incident.LocationRadialHour),
			RadialMinute: formatInt16(storedRow.Incident.LocationRadialMinute),
			Description:  ptr(storedRow.Incident.LocationDescription.String),
			Type:         garett,
		},
		IncidentTypes: &incidentTypes,
		FieldReports:  &fieldReportNumbers,
		RangerHandles: &rangerHandles,
		ReportEntries: resultEntries,
	}

	mustWriteJSON(w, result)
}

func fetchIncident(ctx context.Context, imsDB *store.DB, eventID, incidentNumber int32) (
	incident imsdb.IncidentRow, reportEntries []imsdb.ReportEntry, err error,
) {
	//var incidentTypes imsjson.IncidentTypes
	//var rangerHandles []string
	//var fieldReportNumbers []int32

	incidentRow, err := imsdb.New(imsDB).Incident(ctx, imsdb.IncidentParams{
		Event:  eventID,
		Number: incidentNumber,
	})
	if err != nil {
		return imsdb.IncidentRow{}, nil, fmt.Errorf("[Incident]: %w", err)
	}

	reportEntryRows, err := imsdb.New(imsDB).Incident_ReportEntries(ctx,
		imsdb.Incident_ReportEntriesParams{
			Event:          eventID,
			IncidentNumber: incidentNumber,
		},
	)
	if err != nil {
		return imsdb.IncidentRow{}, nil, fmt.Errorf("[Incident_ReportEntries]: %w", err)
	}
	for _, rer := range reportEntryRows {
		reportEntries = append(reportEntries, rer.ReportEntry)
	}

	return incidentRow, reportEntries, nil

	//var incidentTypes imsjson.IncidentTypes
	//var rangerHandles []string
	//var fieldReportNumbers []int32
	//json.Unmarshal(r.IncidentTypes.([]byte), &incidentTypes)
	//json.Unmarshal(r.RangerHandles.([]byte), &rangerHandles)
	//json.Unmarshal(r.FieldReportNumbers.([]byte), &fieldReportNumbers)
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
	imsDB *store.DB
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
	newIncident.Event = event.Name
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

	if err = updateIncident(ctx, action.imsDB, action.es, newIncident, author); err != nil {
		slog.Error("error updating incident", "err", err)
		http.Error(w, "Failed to update incident", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-IMS-Incident-Number", fmt.Sprint(newIncident.Number))
	w.Header().Set("Location", "/ims/api/events/"+event.Name+"/incidents/"+fmt.Sprint(newIncident.Number))
	http.Error(w, http.StatusText(http.StatusCreated), http.StatusCreated)

	//action.es.notifyIncidentUpdate(event.Name, newIncident.Number)
	//if newIncident.FieldReports != nil {
	//	for _, fr := range *newIncident.FieldReports {
	//		action.es.notifyFieldReportUpdate(event.Name, fr)
	//	}
	//}
}

func unmarshalByteSlice[T any](isByteSlice any) (T, error) {
	var result T
	b, ok := isByteSlice.([]byte)
	if !ok {
		return result, fmt.Errorf("could not read object as []bytes. Was actually %T", b)
	}
	err := json.Unmarshal(b, &result)
	if err != nil {
		return result, fmt.Errorf("[Unmarshal]: %w", err)
	}
	return result, nil
}

func readExtraIncidentRowFields(row imsdb.IncidentRow) (incidentTypes, rangerHandles []string, fieldReportNumbers []int32, err error) {
	incidentTypes, err = unmarshalByteSlice[[]string](row.IncidentTypes)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("[unmarshalByteSlice]: %w", err)
	}
	rangerHandles, err = unmarshalByteSlice[[]string](row.RangerHandles)
	if err != nil {

		return nil, nil, nil, fmt.Errorf("[unmarshalByteSlice]: %w", err)
	}
	fieldReportNumbers, err = unmarshalByteSlice[[]int32](row.FieldReportNumbers)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("[unmarshalByteSlice]: %w", err)
	}
	return incidentTypes, rangerHandles, fieldReportNumbers, nil
}

func updateIncident(ctx context.Context, imsDB *store.DB, es *EventSourcerer, newIncident imsjson.Incident, author string) error {

	storedIncidentRow, err := imsdb.New(imsDB).Incident(ctx, imsdb.IncidentParams{
		Event:  newIncident.EventID,
		Number: newIncident.Number,
	})
	if err != nil {
		return fmt.Errorf("[Incident]: %w", err)
	}
	storedIncident := storedIncidentRow.Incident

	incidentTypes, rangerHandles, fieldReportNumbers, err := readExtraIncidentRowFields(storedIncidentRow)
	if err != nil {
		return fmt.Errorf("[readExtraIncidentRowFields]: %w", err)
	}
	_, _ = incidentTypes, fieldReportNumbers

	txn, _ := imsDB.Begin()
	defer txn.Rollback()
	dbTxn := imsdb.New(txn)

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
	err = dbTxn.UpdateIncident(ctx, update)
	if err != nil {
		return fmt.Errorf("[UpdateIncident]: %w", err)
	}

	if newIncident.RangerHandles != nil {
		add := sliceSubtract(*newIncident.RangerHandles, rangerHandles)
		sub := sliceSubtract(rangerHandles, *newIncident.RangerHandles)
		if len(add) > 0 {
			logs = append(logs, fmt.Sprintf("Rangers added: %v", add))
			for _, rh := range add {
				err = dbTxn.AttachRangerHandleToIncident(ctx, imsdb.AttachRangerHandleToIncidentParams{
					Event:          newIncident.EventID,
					IncidentNumber: newIncident.Number,
					RangerHandle:   rh,
				})
				if err != nil {
					return fmt.Errorf("[AttachRangerHandleToIncident]: %w", err)
				}
			}
		}
		if len(sub) > 0 {
			logs = append(logs, fmt.Sprintf("Rangers removed: %v", sub))
			for _, rh := range sub {
				err = dbTxn.DetachRangerHandleFromIncident(ctx, imsdb.DetachRangerHandleFromIncidentParams{
					Event:          newIncident.EventID,
					IncidentNumber: newIncident.Number,
					RangerHandle:   rh,
				})
				if err != nil {
					return fmt.Errorf("[DetachRangerHandleFromIncident]: %w", err)
				}
			}
		}
	}

	if newIncident.IncidentTypes != nil {
		add := sliceSubtract(*newIncident.IncidentTypes, incidentTypes)
		sub := sliceSubtract(incidentTypes, *newIncident.IncidentTypes)
		if len(add) > 0 {
			logs = append(logs, fmt.Sprintf("type added: %v", add))
			for _, itype := range add {
				err = dbTxn.AttachIncidentTypeToIncident(ctx, imsdb.AttachIncidentTypeToIncidentParams{
					Event:          newIncident.EventID,
					IncidentNumber: newIncident.Number,
					Name:           itype,
				})
				if err != nil {
					return fmt.Errorf("[AttachIncidentTypeToIncident]: %w", err)
				}
			}
		}
		if len(sub) > 0 {
			logs = append(logs, fmt.Sprintf("type removed: %v", sub))
			for _, rh := range sub {
				err = dbTxn.DetachIncidentTypeFromIncident(ctx, imsdb.DetachIncidentTypeFromIncidentParams{
					Event:          newIncident.EventID,
					IncidentNumber: newIncident.Number,
					Name:           rh,
				})
				if err != nil {
					return fmt.Errorf("[DetachIncidentTypeFromIncident]: %w", err)
				}
			}
		}
	}
	var updatedFieldReports []int32
	if newIncident.FieldReports != nil {
		add := sliceSubtract(*newIncident.FieldReports, fieldReportNumbers)
		sub := sliceSubtract(fieldReportNumbers, *newIncident.FieldReports)
		updatedFieldReports = append(updatedFieldReports, add...)
		updatedFieldReports = append(updatedFieldReports, sub...)

		if len(add) > 0 {
			logs = append(logs, fmt.Sprintf("Field Report added: %v", add))
			for _, frNum := range add {
				err = dbTxn.AttachFieldReportToIncident(ctx, imsdb.AttachFieldReportToIncidentParams{
					Event:          newIncident.EventID,
					Number:         frNum,
					IncidentNumber: sql.NullInt32{Int32: newIncident.Number, Valid: true},
				})
				if err != nil {
					return fmt.Errorf("[AttachIncidentTypeToIncident]: %w", err)
				}
			}
		}
		if len(sub) > 0 {
			logs = append(logs, fmt.Sprintf("Field Report removed: %v", sub))
			for _, frNum := range sub {
				err = dbTxn.AttachFieldReportToIncident(ctx, imsdb.AttachFieldReportToIncidentParams{
					Event:          newIncident.EventID,
					Number:         frNum,
					IncidentNumber: sql.NullInt32{},
				})
				if err != nil {
					return fmt.Errorf("[AttachFieldReportToIncident]: %w", err)
				}
			}
		}
	}

	if len(logs) > 0 {
		err = addIncidentReportEntry(ctx, dbTxn, newIncident.EventID, newIncident.Number, author, fmt.Sprintf("Changed %v", strings.Join(logs, ", ")), true)
		if err != nil {
			return fmt.Errorf("[addIncidentReportEntry]: %w", err)
		}
	}

	for _, entry := range newIncident.ReportEntries {
		if entry.Text == "" {
			continue
		}
		err = addIncidentReportEntry(ctx, dbTxn, newIncident.EventID, newIncident.Number, author, entry.Text, false)
		if err != nil {
			return fmt.Errorf("[addIncidentReportEntry]: %w", err)
		}
	}

	if err = txn.Commit(); err != nil {
		return fmt.Errorf("[Commit]: %w", err)
	}

	es.notifyIncidentUpdate(newIncident.Event, newIncident.Number)
	for _, fr := range updatedFieldReports {
		es.notifyFieldReportUpdate(newIncident.Event, fr)
	}

	return nil
}

func sliceSubtract[T comparable](a, b []T) []T {
	var ret []T
	for _, item := range a {
		if !slices.Contains(b, item) {
			ret = append(ret, item)
		}
	}
	return ret
}

type EditIncident struct {
	imsDB *store.DB
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
	newIncident.Event = event.Name
	newIncident.EventID = event.ID
	newIncident.Number = int32(incidentNumber)

	jwtCtx := req.Context().Value(JWTContextKey).(JWTContext)
	author := jwtCtx.Claims.RangerHandle()

	if err = updateIncident(ctx, action.imsDB, action.es, newIncident, author); err != nil {
		slog.Error("error updating incident", "err", err)
		http.Error(w, "Failed to update incident", http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)

	//action.es.notifyIncidentUpdate(event.Name, newIncident.Number)
	//// TODO: need to update detached FRs too
	//if newIncident.FieldReports != nil {
	//	for _, fr := range *newIncident.FieldReports {
	//		action.es.notifyFieldReportUpdate(event.Name, fr)
	//	}
	//}

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
