package api

import (
	"database/sql"
	"encoding/json"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"log/slog"
	"net/http"
	"time"
)

type GetFieldReports struct {
	imsDB *sql.DB
}

type GetFieldReportsResponse []imsjson.FieldReport

func (hand GetFieldReports) getFieldReports(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		slog.Error("Failed to parse form", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	generatedLTE := req.Form.Get("exclude_system_entries") != "true" // false means to exclude

	eventRow, err := queries.New(hand.imsDB).QueryEventID(req.Context(), req.PathValue("eventName"))
	if err != nil {
		slog.Error("Failed to get event ID", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	reportEntries, err := queries.New(hand.imsDB).FieldReports_ReportEntries(req.Context(),
		queries.FieldReports_ReportEntriesParams{
			Event:     eventRow.Event.ID,
			Generated: generatedLTE,
		})
	if err != nil {
		log.Println(err)
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

	rows, err := queries.New(hand.imsDB).FieldReports(req.Context(), eventRow.Event.ID)
	if err != nil {
		log.Println(err)
		return
	}

	resp := make(GetFieldReportsResponse, 0)

	for _, r := range rows {
		fr := r.FieldReport
		resp = append(resp, imsjson.FieldReport{
			Event:    ptr(eventRow.Event.Name),
			Number:   ptr(fr.Number),
			Created:  ptr(time.Unix(int64(fr.Created), 0)),
			Summary:  stringOrNil(fr.Summary),
			Incident: int32OrNil(fr.IncidentNumber),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	jjj, _ := json.Marshal(resp)
	w.Write(jjj)
}
