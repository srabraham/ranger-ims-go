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

type GetIncidents struct {
	imsDB *sql.DB
}

func (ga GetIncidents) getIncidents(w http.ResponseWriter, req *http.Request) {

	event := req.PathValue("eventName")

	eventID, err := queries.New(ga.imsDB).QueryEventID(req.Context(), event)
	if err != nil {
		log.Println(err)
		return
	}

	generatedLTE := false // false should mean to exclude

	reportEntries, err := queries.New(ga.imsDB).Incidents_ReportEntries(req.Context(), queries.Incidents_ReportEntriesParams{Event: eventID, Generated: generatedLTE})
	if err != nil {
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

	rows, err := queries.New(ga.imsDB).Incidents(req.Context(), event)
	if err != nil {
		log.Println(err)
		return
	}

	var resp []imsjson.Incident

	var garett = "garett"
	for _, r := range rows {
		//spew.Dump(r)
		var incidentTypes, rangerHandles []string
		var fieldReportNumbers []int32
		json.Unmarshal(r.IncidentTypes.([]byte), &incidentTypes)
		json.Unmarshal(r.RangerHandles.([]byte), &rangerHandles)
		json.Unmarshal(r.FieldReportNumbers.([]byte), &fieldReportNumbers)
		resp = append(resp, imsjson.Incident{
			// TODO: use event from the db, not the request
			Event:   ptr(event),
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
	//jjj, _ := json.MarshalIndent(resp, "", "  ")
	jjj, _ := json.Marshal(resp)
	w.Write(jjj)
	//
	//rows, _ := db.Query("select * from incident")
	//for rows.Next() {
	//	log.Println(rows.Scan())
	//}
}
