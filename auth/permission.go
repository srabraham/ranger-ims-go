package auth

// per-event

type Role string

var (
	EventReader   Role = "EventReader"
	EventWriter   Role = "EventWriter"
	EventReporter Role = "EventReporter"
)

type Permission string

var (
	ReadIncidents        Permission = "ReadIncidents"
	WriteIncidents       Permission = "WriteIncidents"
	ReadAllFieldReports  Permission = "ReadAllFieldReports"
	WriteAllFieldReports Permission = "WriteAllFieldReports"
	ReadOwnFieldReports  Permission = "ReadOwnFieldReports"
	WriteFieldReports    Permission = "WriteOwnFieldReports"
	ReadPersonnel        Permission = "ReadPersonnel"
	Admin                Permission = "Admin"
)

// global

type Rule struct {
	Event    int32
	Subjects []string
	Actions  []string
}
