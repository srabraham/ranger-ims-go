package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/launchdarkly/eventsource"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type GetFieldReports struct {
	imsDB *sql.DB
}

func (action GetFieldReports) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := make(imsjson.FieldReports, 0)
	ctx := req.Context()

	if ok := parseForm(w, req); !ok {
		return
	}
	generatedLTE := req.Form.Get("exclude_system_entries") != "true" // false means to exclude

	event, ok := eventFromName(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}
	reportEntries, err := imsdb.New(action.imsDB).FieldReports_ReportEntries(ctx,
		imsdb.FieldReports_ReportEntriesParams{
			Event:     event.ID,
			Generated: generatedLTE,
		})
	if err != nil {
		slog.Error("Failed to get FR report entries", "error", err)
		http.Error(w, "Failed to fetch report entries", http.StatusInternalServerError)
		return
	}

	entriesByFR := make(map[int32][]imsjson.ReportEntry)
	for _, row := range reportEntries {
		re := row.ReportEntry
		entriesByFR[row.FieldReportNumber] = append(entriesByFR[row.FieldReportNumber], imsjson.ReportEntry{
			ID:            ptr(re.ID),
			Created:       ptr(time.Unix(int64(re.Created), 0)),
			Author:        ptr(re.Author),
			SystemEntry:   ptr(re.Generated),
			Text:          ptr(re.Text),
			Stricken:      ptr(re.Stricken),
			HasAttachment: ptr(re.AttachedFile.String != ""),
		})
	}

	rows, err := imsdb.New(action.imsDB).FieldReports(ctx, event.ID)
	if err != nil {
		log.Println(err)
		return
	}

	for _, r := range rows {
		fr := r.FieldReport
		entries := entriesByFR[fr.Number]
		resp = append(resp, imsjson.FieldReport{
			Event:         ptr(event.Name),
			Number:        ptr(fr.Number),
			Created:       ptr(time.Unix(int64(fr.Created), 0)),
			Summary:       stringOrNil(fr.Summary),
			Incident:      int32OrNil(fr.IncidentNumber),
			ReportEntries: entries,
		})
	}

	writeJSON(w, resp)
}

type GetFieldReport struct {
	imsDB *sql.DB
}

func (action GetFieldReport) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	response := imsjson.FieldReport{}
	ctx := req.Context()

	event, ok := eventFromName(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}
	fieldReportNumber, err := strconv.ParseInt(req.PathValue("fieldReportNumber"), 10, 32)
	if err != nil {
		slog.Error("Failed to parse field report number", "error", err)
		http.Error(w, "Failed to parse field report number", http.StatusBadRequest)
		return
	}

	reportEntryRows, err := imsdb.New(action.imsDB).FieldReport_ReportEntries(ctx,
		imsdb.FieldReport_ReportEntriesParams{
			Event:             event.ID,
			FieldReportNumber: int32(fieldReportNumber),
		})
	if err != nil {
		slog.Error("Failed to get FR report entries", "error", err)
		http.Error(w, "Failed to fetch report entries", http.StatusInternalServerError)
		return
	}

	frRow, err := imsdb.New(action.imsDB).FieldReport(ctx, imsdb.FieldReportParams{
		Event:  event.ID,
		Number: int32(fieldReportNumber),
	})
	if err != nil {
		slog.Error("Failed to get Field Report", "error", err)
		http.Error(w, "Failed to fetch Field Report", http.StatusInternalServerError)
		return
	}
	fr := frRow.FieldReport

	response = imsjson.FieldReport{
		Event:         ptr(event.Name),
		Number:        ptr(fr.Number),
		Created:       ptr(time.Unix(int64(fr.Created), 0)),
		Summary:       stringOrNil(fr.Summary),
		Incident:      int32OrNil(fr.IncidentNumber),
		ReportEntries: []imsjson.ReportEntry{},
	}
	entries := make([]imsjson.ReportEntry, 0)
	for _, rer := range reportEntryRows {
		re := rer.ReportEntry
		entries = append(entries, imsjson.ReportEntry{
			ID:            ptr(re.ID),
			Created:       ptr(time.Unix(int64(re.Created), 0)),
			Author:        ptr(re.Author),
			SystemEntry:   ptr(re.Generated),
			Text:          ptr(re.Text),
			Stricken:      ptr(re.Stricken),
			HasAttachment: ptr(re.AttachedFile.String != ""),
		})
	}
	response.ReportEntries = entries
	writeJSON(w, response)
}

type EditFieldReport struct {
	imsDB       *sql.DB
	eventSource *eventsource.Server
}

func (action EditFieldReport) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if ok := parseForm(w, req); !ok {
		return
	}

	event, ok := eventFromName(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}

	fieldReportNumber, _ := strconv.ParseInt(req.PathValue("fieldReportNumber"), 10, 32)

	queryAction := req.FormValue("action")
	if queryAction != "" {
		var incident sql.NullInt32

		switch queryAction {
		case "attach":
			num, _ := strconv.ParseInt(req.FormValue("incident"), 10, 32)
			incident = sql.NullInt32{Int32: int32(num), Valid: true}
		case "detach":
			incident = sql.NullInt32{Valid: false}
		default:
			slog.Error("Invalid action", "action", req.FormValue("action"))
			http.Error(w, "Invalid action", http.StatusBadRequest)
			return
		}
		_ = imsdb.New(action.imsDB).AttachFieldReportToIncident(ctx, imsdb.AttachFieldReportToIncidentParams{
			IncidentNumber: incident,
			Event:          event.ID,
			Number:         int32(fieldReportNumber),
		})
		slog.Info("attached FR to incident", "event", event.ID, "incident", incident.Int32, "FR", fieldReportNumber)
	}

	bod, _ := io.ReadAll(req.Body)
	defer req.Body.Close()
	requestFR := imsjson.FieldReport{}
	_ = json.Unmarshal(bod, &requestFR)

	slog.Info("unmarshalled", "requestFR", requestFR)

	frr, _ := imsdb.New(action.imsDB).FieldReport(ctx, imsdb.FieldReportParams{
		Event:  event.ID,
		Number: int32(fieldReportNumber),
	})

	storedFR := frr.FieldReport

	jwtCtx := req.Context().Value(JWTContextKey).(JWTContext)
	author := jwtCtx.Claims.RangerHandle()

	txn, _ := action.imsDB.Begin()
	defer txn.Rollback()
	dbTX := imsdb.New(action.imsDB).WithTx(txn)

	if requestFR.Summary != nil {
		storedFR.Summary = sqlNullString(requestFR.Summary)
		text := "Changed summary to: " + *requestFR.Summary
		err := addFRReportEntry(ctx, dbTX, event.ID, storedFR.Number, author, text, true)
		if err != nil {
			slog.Error("Error adding system fr report entry", "error", err)
			http.Error(w, "Error adding report entry", http.StatusInternalServerError)
			return
		}
	}
	_ = dbTX.UpdateFieldReport(ctx, imsdb.UpdateFieldReportParams{
		Summary:        storedFR.Summary,
		Event:          storedFR.Event,
		Number:         storedFR.Number,
		IncidentNumber: storedFR.IncidentNumber,
	})
	for _, entry := range requestFR.ReportEntries {
		if entry.Text == nil {
			continue
		}
		err := addFRReportEntry(ctx, dbTX, event.ID, storedFR.Number, author, *entry.Text, false)
		if err != nil {
			slog.Error("Error adding fr report entry", "error", err)
			http.Error(w, "Error adding report entry", http.StatusInternalServerError)
			return
		}
	}

	if err := txn.Commit(); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	http.Error(w, "Success", http.StatusNoContent)

	action.eventSource.Publish([]string{"imsevents"}, IMSEvent{
		EventID: idCounter.Add(1),
		EventData: IMSEventData{
			EventName:         event.Name,
			FieldReportNumber: storedFR.Number,
		},
	})

	//event, ok := eventFromName(w, req, req.PathValue("eventName"), action.imsDB)
	//if !ok {
	//	return
	//}
	//slog.Info("in fieldreport", "attach", req.FormValue("action"), "detach", req.FormValue("detach"))
}

type NewFieldReport struct {
	imsDB       *sql.DB
	eventSource *eventsource.Server
}

func (action NewFieldReport) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	//if ok := parseForm(w, req); !ok {
	//	return
	//}

	event, ok := eventFromName(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}

	bod, _ := io.ReadAll(req.Body)
	defer req.Body.Close()
	fr := imsjson.FieldReport{}
	_ = json.Unmarshal(bod, &fr)

	if fr.Incident != nil {
		slog.Error("New FR may not be attached to an incident", "incident", fr.Incident)
		http.Error(w, "New FR may not be attached to an incident", http.StatusBadRequest)
		return
	}

	jwtCtx := req.Context().Value(JWTContextKey).(JWTContext)
	author := jwtCtx.Claims.RangerHandle()

	numUntyped, _ := imsdb.New(action.imsDB).MaxFieldReportNumber(ctx, event.ID)
	newFrNum := numUntyped.(int64) + 1

	txn, _ := action.imsDB.Begin()
	defer txn.Rollback()
	dbTX := imsdb.New(action.imsDB).WithTx(txn)

	_ = dbTX.CreateFieldReport(ctx, imsdb.CreateFieldReportParams{
		Event:          event.ID,
		Number:         int32(newFrNum),
		Created:        float64(time.Now().Unix()),
		Summary:        sqlNullString(fr.Summary),
		IncidentNumber: sql.NullInt32{},
	})

	for _, entry := range fr.ReportEntries {
		if entry.Text == nil {
			continue
		}
		err := addFRReportEntry(ctx, dbTX, event.ID, int32(newFrNum), author, *entry.Text, false)
		if err != nil {
			slog.Error("Error adding system fr report entry", "error", err)
			http.Error(w, "Error adding report entry", http.StatusInternalServerError)
			return
		}
	}

	if fr.Summary != nil {
		text := "Changed summary to: " + *fr.Summary
		err := addFRReportEntry(ctx, dbTX, event.ID, int32(newFrNum), author, text, true)
		if err != nil {
			slog.Error("Error adding system fr report entry", "error", err)
			http.Error(w, "Error adding report entry", http.StatusInternalServerError)
			return
		}
	}

	if err := txn.Commit(); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-IMS-Field-Report-Number", strconv.FormatInt(newFrNum, 10))
	w.Header().Set("Location", "/ims/api/events/"+event.Name+"/field_reports/"+strconv.FormatInt(newFrNum, 10))
	http.Error(w, http.StatusText(http.StatusCreated), http.StatusCreated)

	action.eventSource.Publish([]string{"imsevents"}, IMSEvent{
		EventID: idCounter.Add(1),
		EventData: IMSEventData{
			EventName:         event.Name,
			FieldReportNumber: *fr.Number,
		},
	})
}

func addFRReportEntry(ctx context.Context, q *imsdb.Queries, eventID, frNum int32, author, text string, generated bool) error {
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
	err = q.AttachReportEntryToFieldReport(ctx, imsdb.AttachReportEntryToFieldReportParams{
		Event:             eventID,
		FieldReportNumber: frNum,
		ReportEntry:       int32(reID),
	})
	if err != nil {
		return fmt.Errorf("[AttachReportEntryToFieldReport]: %w", err)
	}
	return nil
}

func sqlNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
