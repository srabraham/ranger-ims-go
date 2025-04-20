package access

import (
	"context"
	"database/sql"
	"fmt"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
)

func GetEventsAccess(ctx context.Context, imsDB *sql.DB, eventName string) (imsjson.EventsAccess, error) {
	var events []queries.Event
	if eventName != "" {
		eventID, err := queries.New(imsDB).QueryEventID(ctx, eventName)
		if err != nil {
			return nil, fmt.Errorf("[QueryEventID]: %w", err)
		}
		events = append(events, queries.Event{
			ID:   eventID,
			Name: eventName,
		})
	} else {
		allEvents, err := queries.New(imsDB).Events(ctx)
		if err != nil {
			return nil, fmt.Errorf("[Events]: %w", err)
		}
		events = append(events, allEvents...)
	}

	resp := make(imsjson.EventsAccess)

	for _, e := range events {
		accesses, err := queries.New(imsDB).EventAccess(ctx, e.ID)
		if err != nil {
			return nil, fmt.Errorf("[EventAccess]: %w", err)
		}
		ea := imsjson.EventAccess{
			Readers:   []imsjson.AccessRule{},
			Writers:   []imsjson.AccessRule{},
			Reporters: []imsjson.AccessRule{},
		}
		for _, access := range accesses {
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
