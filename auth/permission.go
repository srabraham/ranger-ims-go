package auth

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
