package api

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/srabraham/ranger-ims-go/auth"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type GetFieldReports struct {
	imsDB     *store.DB
	imsAdmins []string
}

func (action GetFieldReports) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := make(imsjson.FieldReports, 0)
	event, jwtCtx, eventPermissions, ok := mustGetEventPermissions(w, req, action.imsDB, action.imsAdmins)
	if !ok {
		return
	}
	if eventPermissions&(auth.EventReadAllFieldReports|auth.EventReadOwnFieldReports) == 0 {
		slog.Error("The requestor does not have permission to read Field Reports on this Event")
		http.Error(w, "The requestor does not have permission to read Field Reports on this Event", http.StatusForbidden)
		return
	}
	// i.e. they have EventReadOwnFieldReports, but not EventReadAllFieldReports
	limitedAccess := eventPermissions&auth.EventReadAllFieldReports == 0

	if ok := mustParseForm(w, req); !ok {
		return
	}
	generatedLTE := req.Form.Get("exclude_system_entries") != "true" // false means to exclude

	ctx := req.Context()
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
		http.Error(w, "Failed to fetch Field Reports", http.StatusInternalServerError)
		return
	}

	for _, r := range rows {
		fr := r.FieldReport
		entries := entriesByFR[fr.Number]
		if !limitedAccess || containsAuthor(entries, jwtCtx.Claims.RangerHandle()) {
			resp = append(resp, imsjson.FieldReport{
				Event:         event.Name,
				Number:        fr.Number,
				Created:       time.Unix(int64(fr.Created), 0),
				Summary:       stringOrNil(fr.Summary),
				Incident:      fr.IncidentNumber.Int32,
				ReportEntries: entries,
			})
		}
	}

	mustWriteJSON(w, resp)
}

func containsAuthor(entries []imsjson.ReportEntry, author string) bool {
	for _, e := range entries {
		if e.Author == author {
			return true
		}
	}
	return false
}

type GetFieldReport struct {
	imsDB *store.DB
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
		})
	}
	response.ReportEntries = entries
	mustWriteJSON(w, response)
}

type EditFieldReport struct {
	imsDB       *store.DB
	eventSource *EventSourcerer
	imsAdmins   []string
}

func (action EditFieldReport) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	event, jwtCtx, eventPermissions, ok := mustGetEventPermissions(w, req, action.imsDB, action.imsAdmins)
	if !ok {
		return
	}
	if eventPermissions&(auth.EventWriteAllFieldReports|auth.EventWriteOwnFieldReports) == 0 {
		slog.Error("The requestor does not have permission to write Field Reports on this Event")
		http.Error(w, "The requestor does not have permission to write Field Reports on this Event", http.StatusForbidden)
		return
	}
	// i.e. they have EventWriteOwnFieldReports, but not EventWriteAllFieldReports
	limitedAccess := eventPermissions&auth.EventWriteAllFieldReports == 0

	ctx := req.Context()
	if ok := mustParseForm(w, req); !ok {
		return
	}
	fieldReportNumber, _ := strconv.ParseInt(req.PathValue("fieldReportNumber"), 10, 32)
	author := jwtCtx.Claims.RangerHandle()
	if limitedAccess {
		if ok := action.mustCheckIfPreviousAuthor(w, ctx, event.ID, int32(fieldReportNumber), author); !ok {
			return
		}
	}
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
	dbTxn := imsdb.New(txn) //.WithTx(txn)

	if requestFR.Summary != nil {
		storedFR.Summary = sqlNullString(requestFR.Summary)
		text := "Changed summary to: " + *requestFR.Summary
		err := addFRReportEntry(ctx, dbTxn, event.ID, storedFR.Number, author, text, true)
		if err != nil {
			slog.Error("Error adding system fr report entry", "error", err)
			http.Error(w, "Error adding report entry", http.StatusInternalServerError)
			return
		}
	}
	_ = dbTxn.UpdateFieldReport(ctx, imsdb.UpdateFieldReportParams{
		Event:          storedFR.Event,
		Number:         storedFR.Number,
		Summary:        storedFR.Summary,
		IncidentNumber: storedFR.IncidentNumber,
	})
	for _, entry := range requestFR.ReportEntries {
		if entry.Text == "" {
			continue
		}
		err := addFRReportEntry(ctx, dbTxn, event.ID, storedFR.Number, author, entry.Text, false)
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

func (action EditFieldReport) mustCheckIfPreviousAuthor(
	w http.ResponseWriter,
	ctx context.Context,
	eventID int32,
	fieldReportNumber int32,
	author string,
) (isPreviousAuthor bool) {
	entries, err := imsdb.New(action.imsDB).FieldReport_ReportEntries(ctx,
		imsdb.FieldReport_ReportEntriesParams{
			Event:             eventID,
			FieldReportNumber: fieldReportNumber,
		})
	if err != nil {
		slog.Error("Failed to get FR report entries", "error", err)
		http.Error(w, "Failed to get FR report entries", http.StatusInternalServerError)
		return false
	}
	authorMatch := false
	for _, entry := range entries {
		if entry.ReportEntry.Author == author {
			authorMatch = true
			break
		}
	}
	if !authorMatch {
		slog.Error("EditFieldReport denied to user who is not a previous author on this FieldReport")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}
	return true
}

type NewFieldReport struct {
	imsDB       *store.DB
	eventSource *EventSourcerer
	imsAdmins   []string
}

func (action NewFieldReport) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	event, jwtCtx, eventPermissions, ok := mustGetEventPermissions(w, req, action.imsDB, action.imsAdmins)
	if !ok {
		return
	}
	if eventPermissions&(auth.EventWriteAllFieldReports|auth.EventWriteOwnFieldReports) == 0 {
		slog.Error("The requestor does not have permission to write Field Reports on this Event")
		http.Error(w, "The requestor does not have permission to write Field Reports on this Event", http.StatusForbidden)
		return
	}
	ctx := req.Context()

	fr, ok := mustReadBodyAs[imsjson.FieldReport](w, req)
	if !ok {
		return
	}

	if fr.Incident != 0 {
		slog.Error("New FR may not be attached to an incident", "incident", fr.Incident)
		http.Error(w, "New FR may not be attached to an incident", http.StatusBadRequest)
		return
	}

	author := jwtCtx.Claims.RangerHandle()
	numUntyped, _ := imsdb.New(action.imsDB).MaxFieldReportNumber(ctx, event.ID)
	newFrNum := numUntyped.(int64) + 1

	txn, _ := action.imsDB.Begin()
	defer txn.Rollback()
	dbTxn := imsdb.New(txn)

	_ = dbTxn.CreateFieldReport(ctx, imsdb.CreateFieldReportParams{
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
		err := addFRReportEntry(ctx, dbTxn, event.ID, int32(newFrNum), author, entry.Text, false)
		if err != nil {
			slog.Error("Error adding system fr report entry", "error", err)
			http.Error(w, "Error adding report entry", http.StatusInternalServerError)
			return
		}
	}

	if fr.Summary != nil {
		text := "Changed summary to: " + *fr.Summary
		err := addFRReportEntry(ctx, dbTxn, event.ID, int32(newFrNum), author, text, true)
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
