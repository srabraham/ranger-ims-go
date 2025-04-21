package auth

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"maps"
	"slices"
	"strings"
)

type Role string

const (
	EventReporter Role = "EventReporter"
	EventReader   Role = "EventReader"
	EventWriter   Role = "EventWriter"
	Administrator Role = "Administrator"
)

type Permission string

const (
	ReadEventName        Permission = "ReadEventName"
	ReadIncidents        Permission = "ReadIncidents"
	WriteIncidents       Permission = "WriteIncidents"
	ReadAllFieldReports  Permission = "ReadAllFieldReports"
	WriteAllFieldReports Permission = "WriteAllFieldReports"
	WriteOwnFieldReports Permission = "WriteOwnFieldReports"
	ReadOwnFieldReports  Permission = "ReadOwnFieldReports"
	ReadPersonnel        Permission = "ReadPersonnel"
	AdminIMS             Permission = "AdminIMS"
)

var RolesToPerms = map[Role]map[Permission]bool{
	EventReporter: {
		ReadEventName:        true,
		WriteOwnFieldReports: true,
		ReadOwnFieldReports:  true,
	},
	EventReader: {
		ReadEventName:       true,
		ReadIncidents:       true,
		ReadAllFieldReports: true,
		ReadOwnFieldReports: true,
		ReadPersonnel:       true,
	},
	EventWriter: {
		ReadEventName:        true,
		ReadIncidents:        true,
		WriteIncidents:       true,
		ReadAllFieldReports:  true,
		WriteAllFieldReports: true,
		WriteOwnFieldReports: true,
		ReadOwnFieldReports:  true,
		ReadPersonnel:        true,
	},
	Administrator: {
		AdminIMS: true,
	},
}

func UserPermissions2(
	ctx context.Context,
	eventName string, // or "" for no event
	imsDB *sql.DB,
	imsAdmins []string,
	claims IMSClaims,
) (map[Permission]bool, error) {
	eventRow, err := queries.New(imsDB).QueryEventID(ctx, eventName)
	if err != nil {
		return nil, fmt.Errorf("QueryEventID: %w", err)
	}
	accessRows, err := queries.New(imsDB).EventAccess(ctx, eventRow.Event.ID)
	if err != nil {
		return nil, fmt.Errorf("EventAccess: %w", err)
	}
	var eventAccesses []queries.EventAccess
	for _, ea := range accessRows {
		eventAccesses = append(eventAccesses, ea.EventAccess)
	}

	permissions := UserPermissions(eventAccesses, imsAdmins, claims.RangerHandle(), claims.RangerOnSite(), claims.RangerPositions(), claims.RangerTeams())
	return permissions, nil
}

func UserPermissions(
	eventAccesses []queries.EventAccess, // all for the same event, or nil for no event
	imsAdmins []string,
	handle string,
	onsite bool,
	positions, teams []string,
) map[Permission]bool {

	translate := map[queries.EventAccessMode]Role{
		queries.EventAccessModeRead:   EventReader,
		queries.EventAccessModeWrite:  EventWriter,
		queries.EventAccessModeReport: EventReporter,
	}

	perms := make(map[Permission]bool)

	if slices.Contains(imsAdmins, handle) {
		maps.Copy(perms, RolesToPerms[Administrator])
	}

	for _, ea := range eventAccesses {
		matchExpr := false
		if ea.Expression == "*" {
			matchExpr = true
		}
		if strings.HasPrefix(ea.Expression, "person:") &&
			strings.TrimPrefix(ea.Expression, "person:") == handle {
			matchExpr = true
		}
		if strings.HasPrefix(ea.Expression, "position:") &&
			slices.Contains(positions, strings.TrimPrefix(ea.Expression, "position:")) {
			matchExpr = true
		}
		if strings.HasPrefix(ea.Expression, "team:") &&
			slices.Contains(teams, strings.TrimPrefix(ea.Expression, "team:")) {
			matchExpr = true
		}
		matchValidity := false
		if ea.Validity == queries.EventAccessValidityAlways {
			matchValidity = true
		}
		if ea.Validity == queries.EventAccessValidityOnsite && onsite {
			matchValidity = true
		}
		if matchExpr && matchValidity {
			maps.Copy(perms, RolesToPerms[translate[ea.Mode]])
		}
	}
	return perms
}
