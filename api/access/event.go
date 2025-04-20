package access

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/srabraham/ranger-ims-go/auth"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"maps"
	"slices"
	"strings"
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
	log.Printf("event access: %v", resp)
	return resp, nil
}

func UserPermissions(
	ea []queries.EventAccessRow, // all for the same event, or nil for no event
	imsAdmins []string,
	handle string,
	onsite bool,
	positions, teams []string,
) map[auth.Permission]bool {

	translate := map[queries.EventAccessMode]auth.Role{
		queries.EventAccessModeRead:   auth.EventReader,
		queries.EventAccessModeWrite:  auth.EventWriter,
		queries.EventAccessModeReport: auth.EventReporter,
	}

	perms := make(map[auth.Permission]bool)

	if slices.Contains(imsAdmins, handle) {
		maps.Copy(perms, auth.RolesToPerms[auth.Administrator])
	}

	for _, access := range ea {
		matchExpr := false
		if access.Expression == "*" {
			matchExpr = true
		}
		if strings.HasPrefix(access.Expression, "person:") &&
			strings.TrimPrefix(access.Expression, "person:") == handle {
			matchExpr = true
		}
		if strings.HasPrefix(access.Expression, "position:") &&
			slices.Contains(positions, strings.TrimPrefix(access.Expression, "position:")) {
			matchExpr = true
		}
		if strings.HasPrefix(access.Expression, "team:") &&
			slices.Contains(teams, strings.TrimPrefix(access.Expression, "team:")) {
			matchExpr = true
		}
		matchValidity := false
		if access.Validity == queries.EventAccessValidityAlways {
			matchValidity = true
		}
		if access.Validity == queries.EventAccessValidityOnsite && onsite {
			matchValidity = true
		}
		if matchExpr && matchValidity {
			maps.Copy(perms, auth.RolesToPerms[translate[access.Mode]])
		}
	}
	return perms
}
