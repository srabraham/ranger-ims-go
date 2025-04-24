package api

import (
	"context"
	"database/sql"
	"fmt"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
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

	if ok := mustParseForm(w, req); !ok {
		return
	}
	generatedLTE := req.Form.Get("exclude_system_entries") != "true" // false means to exclude

	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
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
			ID:            re.ID,
			Created:       time.Unix(int64(re.Created), 0),
			Author:        re.Author,
			SystemEntry:   re.Generated,
			Text:          re.Text,
			Stricken:      re.Stricken,
			HasAttachment: re.AttachedFile.String != "",
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
			Event:         event.Name,
			Number:        fr.Number,
			Created:       time.Unix(int64(fr.Created), 0),
			Summary:       stringOrNil(fr.Summary),
			Incident:      fr.IncidentNumber.Int32,
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

	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
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
		Event:         event.Name,
		Number:        fr.Number,
		Created:       time.Unix(int64(fr.Created), 0),
		Summary:       stringOrNil(fr.Summary),
		Incident:      fr.IncidentNumber.Int32,
		ReportEntries: []imsjson.ReportEntry{},
	}
	entries := make([]imsjson.ReportEntry, 0)
	for _, rer := range reportEntryRows {
		re := rer.ReportEntry
		entries = append(entries, imsjson.ReportEntry{
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
		})
	}
	response.ReportEntries = entries
	writeJSON(w, response)
}

type EditFieldReport struct {
	imsDB       *sql.DB
	eventSource *EventSourcerer
}

func (action EditFieldReport) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if ok := mustParseForm(w, req); !ok {
		return
	}

	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}

	jwtCtx := req.Context().Value(JWTContextKey).(JWTContext)
	author := jwtCtx.Claims.RangerHandle()

	fieldReportNumber, _ := strconv.ParseInt(req.PathValue("fieldReportNumber"), 10, 32)

	queryAction := req.FormValue("action")
	if queryAction != "" {
		var incident sql.NullInt32

		entryText := ""
		switch queryAction {
		case "attach":
			num, _ := strconv.ParseInt(req.FormValue("incident"), 10, 32)
			incident = sql.NullInt32{Int32: int32(num), Valid: true}
			entryText = fmt.Sprintf("Attached to incident %v", num)
		case "detach":
			incident = sql.NullInt32{Valid: false}
			entryText = "Detached from incident"
		default:
			slog.Error("Invalid action", "action", req.FormValue("action"))
			http.Error(w, "Invalid action", http.StatusBadRequest)
			return
		}
		err := imsdb.New(action.imsDB).AttachFieldReportToIncident(ctx, imsdb.AttachFieldReportToIncidentParams{
			IncidentNumber: incident,
			Event:          event.ID,
			Number:         int32(fieldReportNumber),
		})
		if err != nil {
			panic(err)
		}
		err = addFRReportEntry(ctx, imsdb.New(action.imsDB), event.ID, int32(fieldReportNumber), author, entryText, true)
		if err != nil {
			panic(err)
		}
		slog.Info("attached FR to incident", "event", event.ID, "incident", incident.Int32, "FR", fieldReportNumber)
	}

	requestFR, ok := mustReadBodyAs[imsjson.FieldReport](w, req)
	if !ok {
		return
	}
	// This is fine, as it may be that only an attach/detach was requested
	if requestFR.Number == 0 {
		slog.Debug("No field report number provided")
		http.Error(w, "OK", http.StatusNoContent)
		return
	}

	//slog.Info("unmarshalled", "requestFR", requestFR)

	frr, _ := imsdb.New(action.imsDB).FieldReport(ctx, imsdb.FieldReportParams{
		Event:  event.ID,
		Number: int32(fieldReportNumber),
	})

	storedFR := frr.FieldReport

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
		Event:          storedFR.Event,
		Number:         storedFR.Number,
		Summary:        storedFR.Summary,
		IncidentNumber: storedFR.IncidentNumber,
	})
	for _, entry := range requestFR.ReportEntries {
		if entry.Text == "" {
			continue
		}
		err := addFRReportEntry(ctx, dbTX, event.ID, storedFR.Number, author, entry.Text, false)
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

	action.eventSource.notifyFieldReportUpdate(event.Name, storedFR.Number)
}

type NewFieldReport struct {
	imsDB       *sql.DB
	eventSource *EventSourcerer
}

func (action NewFieldReport) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), action.imsDB)
	if !ok {
		return
	}
	fr, ok := mustReadBodyAs[imsjson.FieldReport](w, req)
	if !ok {
		return
	}

	if fr.Incident != 0 {
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
		if entry.Text == "" {
			continue
		}
		err := addFRReportEntry(ctx, dbTX, event.ID, int32(newFrNum), author, entry.Text, false)
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

	action.eventSource.notifyFieldReportUpdate(event.Name, int32(newFrNum))
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
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
