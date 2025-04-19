package api

import (
	"database/sql"
	"encoding/json"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"net/http"
	"time"
)

type GetFieldReports struct {
	imsDB *sql.DB
}

func (gfr GetFieldReports) getFieldReports(w http.ResponseWriter, req *http.Request) {

	event := req.PathValue("eventName")

	eventID, err := queries.New(gfr.imsDB).QueryEventID(req.Context(), event)
	if err != nil {
		log.Println(err)
		return
	}

	generatedLTE := false // false should mean to exclude

	reportEntries, err := queries.New(gfr.imsDB).FieldReports_ReportEntries(req.Context(), queries.FieldReports_ReportEntriesParams{Event: eventID, Generated: generatedLTE})
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

	rows, err := queries.New(gfr.imsDB).FieldReports(req.Context(), eventID)
	if err != nil {
		log.Println(err)
		return
	}

	var resp []imsjson.FieldReport

	for _, r := range rows {
		fr := r.FieldReport
		resp = append(resp, imsjson.FieldReport{
			// TODO: use event from the db, not the request
			Event:    ptr(event),
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
