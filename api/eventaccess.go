package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log/slog"
	"net/http"
)

type GetEventAccesses struct {
	imsDB *sql.DB
}

func (hand GetEventAccesses) getEventAccesses(w http.ResponseWriter, req *http.Request) {
	resp := imsjson.EventsAccess{}

	eventName := req.PathValue("eventName")

	var err error
	resp, err = GetEventsAccess(req.Context(), hand.imsDB, eventName)
	if err != nil {
		slog.Error("GetEventsAccess failed", "error", err)
		http.Error(w, "Failed to get events access", http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp)
}

func GetEventsAccess(ctx context.Context, imsDB *sql.DB, eventName string) (imsjson.EventsAccess, error) {
	var events []queries.Event
	if eventName != "" {
		eventRow, err := queries.New(imsDB).QueryEventID(ctx, eventName)
		if err != nil {
			return nil, fmt.Errorf("[QueryEventID]: %w", err)
		}
		events = append(events, eventRow.Event)
	} else {
		allEventRows, err := queries.New(imsDB).Events(ctx)
		if err != nil {
			return nil, fmt.Errorf("[Events]: %w", err)
		}
		for _, aer := range allEventRows {
			events = append(events, aer.Event)
		}
	}

	resp := make(imsjson.EventsAccess)

	for _, e := range events {
		accessRows, err := queries.New(imsDB).EventAccess(ctx, e.ID)
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
			case queries.EventAccessModeRead:
				ea.Readers = append(ea.Readers, rule)
			case queries.EventAccessModeWrite:
				ea.Writers = append(ea.Writers, rule)
			case queries.EventAccessModeReport:
				ea.Reporters = append(ea.Reporters, rule)
			}
		}
		resp[e.Name] = ea
	}
	return resp, nil
}

type PostEventAccess struct {
	imsDB *sql.DB
}

func (handler PostEventAccess) postEventAccess(w http.ResponseWriter, req *http.Request) {
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
		event, success := eventFromName(w, req, eventName, handler.imsDB)
		if !success {
			return
		}
		errs = append(errs, handler.maybeSetAccess(req.Context(), event, access.Readers, queries.EventAccessModeRead))
		errs = append(errs, handler.maybeSetAccess(req.Context(), event, access.Writers, queries.EventAccessModeWrite))
		errs = append(errs, handler.maybeSetAccess(req.Context(), event, access.Reporters, queries.EventAccessModeReport))
	}
	if err := errors.Join(errs...); err != nil {
		slog.Error("PostEventAccess", "error", err)
		http.Error(w, "Failed to set event access", http.StatusInternalServerError)
		return
	}
	http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
}

func (handler PostEventAccess) maybeSetAccess(ctx context.Context, event queries.Event, rules []imsjson.AccessRule, mode queries.EventAccessMode) error {
	if len(rules) == 0 {
		return nil
	}
	txn, err := handler.imsDB.BeginTx(ctx, nil)
	defer txn.Rollback()
	if err != nil {
		return fmt.Errorf("[BeginTx]: %w", err)
	}
	err = queries.New(handler.imsDB).WithTx(txn).ClearEventAccessForMode(ctx, queries.ClearEventAccessForModeParams{
		Event: event.ID,
		Mode:  mode,
	})
	if err != nil {
		return fmt.Errorf("[ClearEventAccessForMode]: %w", err)
	}
	for _, rule := range rules {
		err = queries.New(handler.imsDB).WithTx(txn).ClearEventAccessForExpression(ctx, queries.ClearEventAccessForExpressionParams{
			Event:      event.ID,
			Expression: rule.Expression,
		})
		if err != nil {
			return fmt.Errorf("[ClearEventAccessForExpression]: %w", err)
		}
		_, err = queries.New(handler.imsDB).WithTx(txn).AddEventAccess(ctx, queries.AddEventAccessParams{
			Event:      event.ID,
			Expression: rule.Expression,
			Mode:       mode,
			Validity:   queries.EventAccessValidity(rule.Validity),
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
