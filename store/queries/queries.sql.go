// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries.sql

package queries

import (
	"context"
	"database/sql"
)

const concentricStreets = `-- name: ConcentricStreets :many
select ID, NAME from CONCENTRIC_STREET
where EVENT = ?
`

type ConcentricStreetsRow struct {
	ID   string
	Name string
}

func (q *Queries) ConcentricStreets(ctx context.Context, event int32) ([]ConcentricStreetsRow, error) {
	rows, err := q.db.QueryContext(ctx, concentricStreets, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ConcentricStreetsRow
	for rows.Next() {
		var i ConcentricStreetsRow
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const createEvent = `-- name: CreateEvent :execlastid
insert into EVENT (NAME) values (?)
`

func (q *Queries) CreateEvent(ctx context.Context, name string) (int64, error) {
	result, err := q.db.ExecContext(ctx, createEvent, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createIncident = `-- name: CreateIncident :execlastid
insert into INCIDENT (
    EVENT,
    NUMBER,
    CREATED,
    PRIORITY,
    STATE,
    SUMMARY,
    LOCATION_NAME,
    LOCATION_CONCENTRIC,
    LOCATION_RADIAL_HOUR,
    LOCATION_RADIAL_MINUTE,
    LOCATION_DESCRIPTION
)
values (
   ?,?,?,?,?,?,?,?,?,?,?
)
`

type CreateIncidentParams struct {
	Event                int32
	Number               int32
	Created              float64
	Priority             int8
	State                IncidentState
	Summary              sql.NullString
	LocationName         sql.NullString
	LocationConcentric   sql.NullString
	LocationRadialHour   sql.NullInt16
	LocationRadialMinute sql.NullInt16
	LocationDescription  sql.NullString
}

func (q *Queries) CreateIncident(ctx context.Context, arg CreateIncidentParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createIncident,
		arg.Event,
		arg.Number,
		arg.Created,
		arg.Priority,
		arg.State,
		arg.Summary,
		arg.LocationName,
		arg.LocationConcentric,
		arg.LocationRadialHour,
		arg.LocationRadialMinute,
		arg.LocationDescription,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const eventAccess = `-- name: EventAccess :many
select EXPRESSION, MODE, VALIDITY from EVENT_ACCESS
where EVENT = ?
`

type EventAccessRow struct {
	Expression string
	Mode       EventAccessMode
	Validity   EventAccessValidity
}

func (q *Queries) EventAccess(ctx context.Context, event int32) ([]EventAccessRow, error) {
	rows, err := q.db.QueryContext(ctx, eventAccess, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []EventAccessRow
	for rows.Next() {
		var i EventAccessRow
		if err := rows.Scan(&i.Expression, &i.Mode, &i.Validity); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const events = `-- name: Events :many
select ID, NAME from EVENT
`

func (q *Queries) Events(ctx context.Context) ([]Event, error) {
	rows, err := q.db.QueryContext(ctx, events)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const fieldReports = `-- name: FieldReports :many
select
    field_report.event, field_report.number, field_report.created, field_report.summary, field_report.incident_number
from
    FIELD_REPORT
where
    EVENT = ?
`

type FieldReportsRow struct {
	FieldReport FieldReport
}

func (q *Queries) FieldReports(ctx context.Context, event int32) ([]FieldReportsRow, error) {
	rows, err := q.db.QueryContext(ctx, fieldReports, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FieldReportsRow
	for rows.Next() {
		var i FieldReportsRow
		if err := rows.Scan(
			&i.FieldReport.Event,
			&i.FieldReport.Number,
			&i.FieldReport.Created,
			&i.FieldReport.Summary,
			&i.FieldReport.IncidentNumber,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const fieldReports_ReportEntries = `-- name: FieldReports_ReportEntries :many
select
    irre.FIELD_REPORT_NUMBER,
    re.id, re.author, re.text, re.created, re.` + "`" + `generated` + "`" + `, re.stricken, re.attached_file
from
    FIELD_REPORT__REPORT_ENTRY irre
        join REPORT_ENTRY re
             on irre.REPORT_ENTRY = re.ID
where
    irre.EVENT = ?
    and re.GENERATED <= ?
`

type FieldReports_ReportEntriesParams struct {
	Event     int32
	Generated bool
}

type FieldReports_ReportEntriesRow struct {
	FieldReportNumber int32
	ReportEntry       ReportEntry
}

func (q *Queries) FieldReports_ReportEntries(ctx context.Context, arg FieldReports_ReportEntriesParams) ([]FieldReports_ReportEntriesRow, error) {
	rows, err := q.db.QueryContext(ctx, fieldReports_ReportEntries, arg.Event, arg.Generated)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FieldReports_ReportEntriesRow
	for rows.Next() {
		var i FieldReports_ReportEntriesRow
		if err := rows.Scan(
			&i.FieldReportNumber,
			&i.ReportEntry.ID,
			&i.ReportEntry.Author,
			&i.ReportEntry.Text,
			&i.ReportEntry.Created,
			&i.ReportEntry.Generated,
			&i.ReportEntry.Stricken,
			&i.ReportEntry.AttachedFile,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const incidentTypes = `-- name: IncidentTypes :many
select NAME from INCIDENT_TYPE
`

func (q *Queries) IncidentTypes(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, incidentTypes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const incidents = `-- name: Incidents :many
select
    i.event, i.number, i.created, i.priority, i.state, i.summary, i.location_name, i.location_concentric, i.location_radial_hour, i.location_radial_minute, i.location_description,
    (
        select coalesce(json_arrayagg(it.NAME), "[]")
        from INCIDENT__INCIDENT_TYPE iit
        join INCIDENT_TYPE it
            on i.EVENT = iit.EVENT
            and i.NUMBER = iit.INCIDENT_NUMBER
            and iit.INCIDENT_TYPE = it.ID
    ) as INCIDENT_TYPES,
    (
        select coalesce(json_arrayagg(irep.NUMBER), "[]")
        from FIELD_REPORT irep
        where i.EVENT = irep.EVENT
            and i.NUMBER = irep.INCIDENT_NUMBER
    ) as FIELD_REPORT_NUMBERS,
    (
        select coalesce(json_arrayagg(ir.RANGER_HANDLE), "[]")
        from INCIDENT__RANGER ir
        where i.EVENT = ir.EVENT
            and i.NUMBER = ir.INCIDENT_NUMBER
    ) as RANGER_HANDLES
from
    INCIDENT i
where
    i.EVENT = (select e.ID from EVENT e where e.NAME = ?)
group by
    i.NUMBER
`

type IncidentsRow struct {
	Incident           Incident
	IncidentTypes      interface{}
	FieldReportNumbers interface{}
	RangerHandles      interface{}
}

func (q *Queries) Incidents(ctx context.Context, name string) ([]IncidentsRow, error) {
	rows, err := q.db.QueryContext(ctx, incidents, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []IncidentsRow
	for rows.Next() {
		var i IncidentsRow
		if err := rows.Scan(
			&i.Incident.Event,
			&i.Incident.Number,
			&i.Incident.Created,
			&i.Incident.Priority,
			&i.Incident.State,
			&i.Incident.Summary,
			&i.Incident.LocationName,
			&i.Incident.LocationConcentric,
			&i.Incident.LocationRadialHour,
			&i.Incident.LocationRadialMinute,
			&i.Incident.LocationDescription,
			&i.IncidentTypes,
			&i.FieldReportNumbers,
			&i.RangerHandles,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const incidents_ReportEntries = `-- name: Incidents_ReportEntries :many
select
    ire.INCIDENT_NUMBER,
    re.id, re.author, re.text, re.created, re.` + "`" + `generated` + "`" + `, re.stricken, re.attached_file
from
    INCIDENT__REPORT_ENTRY ire
        join REPORT_ENTRY re
             on re.ID = ire.REPORT_ENTRY
where
    ire.EVENT = ?
    and re.GENERATED <= ?
`

type Incidents_ReportEntriesParams struct {
	Event     int32
	Generated bool
}

type Incidents_ReportEntriesRow struct {
	IncidentNumber int32
	ReportEntry    ReportEntry
}

func (q *Queries) Incidents_ReportEntries(ctx context.Context, arg Incidents_ReportEntriesParams) ([]Incidents_ReportEntriesRow, error) {
	rows, err := q.db.QueryContext(ctx, incidents_ReportEntries, arg.Event, arg.Generated)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Incidents_ReportEntriesRow
	for rows.Next() {
		var i Incidents_ReportEntriesRow
		if err := rows.Scan(
			&i.IncidentNumber,
			&i.ReportEntry.ID,
			&i.ReportEntry.Author,
			&i.ReportEntry.Text,
			&i.ReportEntry.Created,
			&i.ReportEntry.Generated,
			&i.ReportEntry.Stricken,
			&i.ReportEntry.AttachedFile,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const queryEventID = `-- name: QueryEventID :one
select ID from EVENT where NAME = ?
`

func (q *Queries) QueryEventID(ctx context.Context, name string) (int32, error) {
	row := q.db.QueryRowContext(ctx, queryEventID, name)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const schemaVersion = `-- name: SchemaVersion :one
select VERSION from SCHEMA_INFO
`

func (q *Queries) SchemaVersion(ctx context.Context) (int16, error) {
	row := q.db.QueryRowContext(ctx, schemaVersion)
	var version int16
	err := row.Scan(&version)
	return version, err
}
