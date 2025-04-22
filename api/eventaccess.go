package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"log/slog"
	"net/http"
	"sync"
)

type GetEventAccesses struct {
	imsDB *sql.DB
}

func (action GetEventAccesses) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := imsjson.EventsAccess{}
	ctx := req.Context()

	resp, err := GetEventsAccess(ctx, action.imsDB, "")
	if err != nil {
		slog.Error("GetEventsAccess failed", "error", err)
		http.Error(w, "Failed to get events access", http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp)
}

func GetEventsAccess(ctx context.Context, imsDB *sql.DB, eventName string) (imsjson.EventsAccess, error) {
	var events []imsdb.Event
	//if eventName != "" {
	//	eventRow, err := imsdb.New(imsDB).QueryEventID(ctx, eventName)
	//	if err != nil {
	//		return nil, fmt.Errorf("[QueryEventID]: %w", err)
	//	}
	//	events = append(events, eventRow.Event)
	//} else {
	allEventRows, err := imsdb.New(imsDB).Events(ctx)
	if err != nil {
		return nil, fmt.Errorf("[Events]: %w", err)
	}
	for _, aer := range allEventRows {
		events = append(events, aer.Event)
	}
	//}

	result := make(imsjson.EventsAccess)

	for _, e := range events {
		accessRows, err := imsdb.New(imsDB).EventAccess(ctx, e.ID)
		if err != nil {
			return nil, fmt.Errorf("[EventAccess]: %w", err)
		}
		ea := imsjson.EventAccess{
			Readers:   []imsjson.AccessRule{},
			Writers:   []imsjson.AccessRule{},
			Reporters: []imsjson.AccessRule{},
		}
		for _, accessRow := range accessRows {
			access := accessRow.EventAccess
			rule := imsjson.AccessRule{Expression: access.Expression, Validity: string(access.Validity)}
			switch access.Mode {
			case imsdb.EventAccessModeRead:
				ea.Readers = append(ea.Readers, rule)
			case imsdb.EventAccessModeWrite:
				ea.Writers = append(ea.Writers, rule)
			case imsdb.EventAccessModeReport:
				ea.Reporters = append(ea.Reporters, rule)
			}
		}
		result[e.Name] = ea
	}
	return result, nil
}

type PostEventAccess struct {
	imsDB *sql.DB
}

var eventAccessWriteMu sync.Mutex

func (action PostEventAccess) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	eventAccessWriteMu.Lock()
	defer eventAccessWriteMu.Unlock()

	ctx := req.Context()
	if ok := parseForm(w, req); !ok {
		return
	}
	bodyBytes, ok := readBody(w, req)
	if !ok {
		return
	}
	var eventsAccess imsjson.EventsAccess
	if err := json.Unmarshal(bodyBytes, &eventsAccess); err != nil {
		slog.Error("PostEventAccess failed to parse body", "error", err)
		http.Error(w, "Failed to parse body", http.StatusBadRequest)
		return
	}
	var errs []error
	for eventName, access := range eventsAccess {
		event, success := eventFromName(w, req, eventName, action.imsDB)
		if !success {
			return
		}
		errs = append(errs, action.maybeSetAccess(ctx, event, access.Readers, imsdb.EventAccessModeRead))
		errs = append(errs, action.maybeSetAccess(ctx, event, access.Writers, imsdb.EventAccessModeWrite))
		errs = append(errs, action.maybeSetAccess(ctx, event, access.Reporters, imsdb.EventAccessModeReport))
	}
	if err := errors.Join(errs...); err != nil {
		slog.Error("PostEventAccess", "error", err)
		http.Error(w, "Failed to set event access", http.StatusInternalServerError)
		return
	}
	http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
}

func (action PostEventAccess) maybeSetAccess(ctx context.Context, event imsdb.Event, rules []imsjson.AccessRule, mode imsdb.EventAccessMode) error {
	if len(rules) == 0 {
		return nil
	}
	txn, err := action.imsDB.BeginTx(ctx, nil)
	defer txn.Rollback()
	if err != nil {
		return fmt.Errorf("[BeginTx]: %w", err)
	}
	err = imsdb.New(action.imsDB).WithTx(txn).ClearEventAccessForMode(ctx, imsdb.ClearEventAccessForModeParams{
		Event: event.ID,
		Mode:  mode,
	})
	if err != nil {
		return fmt.Errorf("[ClearEventAccessForMode]: %w", err)
	}
	for _, rule := range rules {
		err = imsdb.New(action.imsDB).WithTx(txn).ClearEventAccessForExpression(ctx, imsdb.ClearEventAccessForExpressionParams{
			Event:      event.ID,
			Expression: rule.Expression,
		})
		if err != nil {
			return fmt.Errorf("[ClearEventAccessForExpression]: %w", err)
		}
		_, err = imsdb.New(action.imsDB).WithTx(txn).AddEventAccess(ctx, imsdb.AddEventAccessParams{
			Event:      event.ID,
			Expression: rule.Expression,
			Mode:       mode,
			Validity:   imsdb.EventAccessValidity(rule.Validity),
		})
		if err != nil {
			return fmt.Errorf("[AddEventAccess]: %w", err)
		}
	}
	if err = txn.Commit(); err != nil {
		return fmt.Errorf("[Commit]: %w", err)
	}
	return nil
}
