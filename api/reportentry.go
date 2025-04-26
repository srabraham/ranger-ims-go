package api

import (
	"fmt"
	"github.com/srabraham/ranger-ims-go/auth"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"log"
	"log/slog"
	"net/http"
	"strconv"
)

type EditFieldReportReportEntry struct {
	imsDB       *store.DB
	eventSource *EventSourcerer
	imsAdmins   []string
}

func (action EditFieldReportReportEntry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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

	author := jwtCtx.Claims.RangerHandle()

	var fieldReportNumber int32
	if isInt32(req.PathValue("fieldReportNumber")) {
		fieldReportNumber = toInt32(req.PathValue("fieldReportNumber"))
	} else {
		log.Panicf("wanted int32 FRN, got %v", req.PathValue("fieldReportNumber"))
	}
	var reportEntryId int32
	if isInt32(req.PathValue("reportEntryId")) {
		reportEntryId = toInt32(req.PathValue("reportEntryId"))
	} else {
		log.Panicf("wanted int32 REID, got %v", req.PathValue("reportEntryId"))
	}

	re, ok := mustReadBodyAs[imsjson.ReportEntry](w, req)
	if !ok {
		return
	}

	txn, _ := action.imsDB.Begin()
	defer txn.Rollback()
	dbTxn := imsdb.New(txn)

	err := dbTxn.SetFieldReportReportEntryStricken(ctx, imsdb.SetFieldReportReportEntryStrickenParams{
		Stricken:          re.Stricken,
		Event:             event.ID,
		FieldReportNumber: fieldReportNumber,
		ReportEntry:       reportEntryId,
	})
	if err != nil {
		slog.Error("Error setting field report entry", "error", err)
		http.Error(w, "Error setting stricken value", http.StatusInternalServerError)
		return
	}
	struckVerb := "Struck"
	if !re.Stricken {
		struckVerb = "Unstruck"
	}
	err = addFRReportEntry(ctx, dbTxn, event.ID, fieldReportNumber, author, fmt.Sprintf("%v reportEntry %v", struckVerb, reportEntryId), true)
	if err != nil {
		slog.Error("Error adding report entry", "error", err)
		http.Error(w, "Error adding report entry", http.StatusInternalServerError)
		return
	}
	if err = txn.Commit(); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)

	action.eventSource.notifyFieldReportUpdate(event.Name, fieldReportNumber)
}

type EditIncidentReportEntry struct {
	imsDB       *store.DB
	eventSource *EventSourcerer
	imsAdmins   []string
}

func (action EditIncidentReportEntry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	event, jwtCtx, eventPermissions, ok := mustGetEventPermissions(w, req, action.imsDB, action.imsAdmins)
	if !ok {
		return
	}
	if eventPermissions&(auth.EventWriteIncidents) == 0 {
		slog.Error("The requestor does not have permission to write Report Entries on this Event")
		http.Error(w, "The requestor does not have permission to write Report Entries on this Event", http.StatusForbidden)
		return
	}
	ctx := req.Context()

	author := jwtCtx.Claims.RangerHandle()

	var incidentNumber int32
	if isInt32(req.PathValue("incidentNumber")) {
		incidentNumber = toInt32(req.PathValue("incidentNumber"))
	} else {
		log.Panicf("wanted int32 IN, got %v", req.PathValue("incidentNumber"))
	}
	var reportEntryId int32
	if isInt32(req.PathValue("reportEntryId")) {
		reportEntryId = toInt32(req.PathValue("reportEntryId"))
	} else {
		log.Panicf("wanted int32 REID, got %v", req.PathValue("reportEntryId"))
	}

	re, ok := mustReadBodyAs[imsjson.ReportEntry](w, req)
	if !ok {
		return
	}

	txn, _ := action.imsDB.Begin()
	defer txn.Rollback()
	dbTxn := imsdb.New(txn)

	err := dbTxn.SetIncidentReportEntryStricken(ctx, imsdb.SetIncidentReportEntryStrickenParams{
		Stricken:       re.Stricken,
		Event:          event.ID,
		IncidentNumber: incidentNumber,
		ReportEntry:    reportEntryId,
	})
	if err != nil {
		slog.Error("Error setting incident report entry", "error", err)
		http.Error(w, "Error setting stricken value", http.StatusInternalServerError)
		return
	}
	struckVerb := "Struck"
	if !re.Stricken {
		struckVerb = "Unstruck"
	}
	err = addIncidentReportEntry(ctx, dbTxn, event.ID, incidentNumber, author, fmt.Sprintf("%v reportEntry %v", struckVerb, reportEntryId), true)
	if err != nil {
		slog.Error("Error adding report entry", "error", err)
		http.Error(w, "Error adding report entry", http.StatusInternalServerError)
		return
	}
	if err = txn.Commit(); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)

	action.eventSource.notifyIncidentUpdate(event.Name, incidentNumber)
}

func isInt32(s string) bool {
	_, err := strconv.ParseInt(s, 10, 32)
	return err == nil
}

func toInt32(s string) int32 {
	i, _ := strconv.ParseInt(s, 10, 32)
	return int32(i)
}
