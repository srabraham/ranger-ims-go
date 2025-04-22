package api

import (
	"database/sql"
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

func (handler GetFieldReports) getFieldReports(w http.ResponseWriter, req *http.Request) {
	if ok := parseForm(w, req); !ok {
		return
	}
	generatedLTE := req.Form.Get("exclude_system_entries") != "true" // false means to exclude

	event, ok := eventFromName(w, req, req.PathValue("eventName"), handler.imsDB)
	if !ok {
		return
	}
	reportEntries, err := imsdb.New(handler.imsDB).FieldReports_ReportEntries(req.Context(),
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

	rows, err := imsdb.New(handler.imsDB).FieldReports(req.Context(), event.ID)
	if err != nil {
		log.Println(err)
		return
	}

	resp := make(imsjson.FieldReports, 0)

	for _, r := range rows {
		fr := r.FieldReport
		resp = append(resp, imsjson.FieldReport{
			Event:    ptr(event.Name),
			Number:   ptr(fr.Number),
			Created:  ptr(time.Unix(int64(fr.Created), 0)),
			Summary:  stringOrNil(fr.Summary),
			Incident: int32OrNil(fr.IncidentNumber),
		})
	}

	writeJSON(w, resp)
}

type GetFieldReport struct {
	imsDB *sql.DB
}

func (handler GetFieldReport) getFieldReport(w http.ResponseWriter, req *http.Request) {
	response := imsjson.FieldReport{}

	event, ok := eventFromName(w, req, req.PathValue("eventName"), handler.imsDB)
	if !ok {
		return
	}
	fieldReportNumber, err := strconv.ParseInt(req.PathValue("fieldReportNumber"), 10, 32)
	if err != nil {
		slog.Error("Failed to parse field report number", "error", err)
		http.Error(w, "Failed to parse field report number", http.StatusBadRequest)
		return
	}

	reportEntryRows, err := imsdb.New(handler.imsDB).FieldReport_ReportEntries(req.Context(),
		imsdb.FieldReport_ReportEntriesParams{
			Event:             event.ID,
			FieldReportNumber: int32(fieldReportNumber),
		})
	if err != nil {
		slog.Error("Failed to get FR report entries", "error", err)
		http.Error(w, "Failed to fetch report entries", http.StatusInternalServerError)
		return
	}

	frRow, err := imsdb.New(handler.imsDB).FieldReport(req.Context(), imsdb.FieldReportParams{
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
		ReportEntries: &[]imsjson.ReportEntry{},
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
	response.ReportEntries = &entries
	writeJSON(w, response)
}
