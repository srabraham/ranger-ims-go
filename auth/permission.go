package auth

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"maps"
	"slices"
	"strings"
)

type Role string

const (
	AnyAuthenticatedUser Role = "AnyAuthenticatedUser"
	EventReporter        Role = "EventReporter"
	EventReader          Role = "EventReader"
	EventWriter          Role = "EventWriter"
	Administrator        Role = "Administrator"
)

type Permission string

const (
	// Event-specific permissions

	ReadIncidents     Permission = "ReadIncidents"
	WriteIncidents    Permission = "WriteIncidents"
	ReadFieldReports  Permission = "ReadFieldReports"
	WriteFieldReports Permission = "WriteFieldReports"
	ReadEventName     Permission = "ReadEventName"
	ReadEventStreets  Permission = "ReadEventStreets"

	// Permissions that aren't event-specific

	ReadIncidentTypes         Permission = "ReadIncidentTypes"
	ReadPersonnel             Permission = "ReadPersonnel"
	AdministrateEvents        Permission = "AdministrateEvents"
	AdministrateStreets       Permission = "AdministrateStreets"
	AdministrateIncidentTypes Permission = "AdministrateIncidentTypes"
)

var RolesToPerms = map[Role]map[Permission]bool{
	AnyAuthenticatedUser: {
		ReadIncidentTypes: true,
		ReadPersonnel:     true,
	},
	EventReporter: {
		ReadEventName:     true,
		ReadIncidentTypes: true,
		ReadEventStreets:  true,
		ReadFieldReports:  true,
		WriteFieldReports: true,
	},
	EventReader: {
		ReadEventName:     true,
		ReadEventStreets:  true,
		ReadIncidents:     true,
		ReadIncidentTypes: true,
		ReadFieldReports:  true,
		ReadPersonnel:     true,
	},
	EventWriter: {
		ReadEventName:     true,
		ReadEventStreets:  true,
		ReadIncidents:     true,
		ReadIncidentTypes: true,
		WriteIncidents:    true,
		ReadFieldReports:  true,
		WriteFieldReports: true,
		ReadPersonnel:     true,
	},
	Administrator: {
		AdministrateEvents:        true,
		AdministrateStreets:       true,
		AdministrateIncidentTypes: true,
	},
}

func UserPermissions2(
	ctx context.Context,
	eventName string, // or "" for no event
	imsDB *sql.DB,
	imsAdmins []string,
	claims IMSClaims,
) (map[Permission]bool, error) {
	var eventAccesses []imsdb.EventAccess
	if eventName != "" {
		eventRow, err := imsdb.New(imsDB).QueryEventID(ctx, eventName)
		if err != nil {
			return nil, fmt.Errorf("QueryEventID: %w", err)
		}
		accessRows, err := imsdb.New(imsDB).EventAccess(ctx, eventRow.Event.ID)
		if err != nil {
			return nil, fmt.Errorf("EventAccess: %w", err)
		}
		for _, ea := range accessRows {
			eventAccesses = append(eventAccesses, ea.EventAccess)
		}
	}
	permissions := UserPermissions(eventAccesses, imsAdmins, claims.RangerHandle(), claims.RangerOnSite(), claims.RangerPositions(), claims.RangerTeams())
	return permissions, nil
}

func UserPermissions(
	eventAccesses []imsdb.EventAccess, // all for the same event, or nil for no event
	imsAdmins []string,
	handle string,
	onsite bool,
	positions, teams []string,
) map[Permission]bool {

	translate := map[imsdb.EventAccessMode]Role{
		imsdb.EventAccessModeRead:   EventReader,
		imsdb.EventAccessModeWrite:  EventWriter,
		imsdb.EventAccessModeReport: EventReporter,
	}

	perms := make(map[Permission]bool)

	if handle != "" {
		maps.Copy(perms, RolesToPerms[AnyAuthenticatedUser])
	}

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
		if ea.Validity == imsdb.EventAccessValidityAlways {
			matchValidity = true
		}
		if ea.Validity == imsdb.EventAccessValidityOnsite && onsite {
			matchValidity = true
		}
		if matchExpr && matchValidity {
			maps.Copy(perms, RolesToPerms[translate[ea.Mode]])
		}
	}
	return perms
}
