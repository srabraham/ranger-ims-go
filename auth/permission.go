package auth

import (
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
	ReadIncidents            Permission = "ReadIncidents"
	WriteIncidents           Permission = "WriteIncidents"
	ReadAllFieldReports      Permission = "ReadAllFieldReports"
	WriteAllFieldReports     Permission = "WriteAllFieldReports"
	ReadWriteOwnFieldReports Permission = "ReadWriteOwnFieldReports"
	ReadPersonnel            Permission = "ReadPersonnel"
	AdminIMS                 Permission = "AdminIMS"
)

var RolesToPerms = map[Role]map[Permission]bool{
	EventReporter: {
		ReadWriteOwnFieldReports: true,
	},
	EventReader: {
		ReadIncidents:       true,
		ReadAllFieldReports: true,
		ReadPersonnel:       true,
	},
	EventWriter: {
		ReadIncidents:            true,
		WriteIncidents:           true,
		ReadAllFieldReports:      true,
		WriteAllFieldReports:     true,
		ReadWriteOwnFieldReports: true,
		ReadPersonnel:            true,
	},
	Administrator: {
		AdminIMS: true,
	},
}

func UserPermissions(
	ea []queries.EventAccessRow, // all for the same event, or nil for no event
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
			maps.Copy(perms, RolesToPerms[translate[access.Mode]])
		}
	}
	return perms
}
