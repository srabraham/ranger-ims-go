// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries.sql

package queries

import (
	"context"
	"database/sql"
)

const createEvent = `-- name: CreateEvent :exec
insert into EVENT (NAME) values (?)
`

func (q *Queries) CreateEvent(ctx context.Context, name string) error {
	_, err := q.db.ExecContext(ctx, createEvent, name)
	return err
}

const eventAccess = `-- name: EventAccess :many
select EXPRESSION, VALIDITY from EVENT_ACCESS
where EVENT = (select ID from EVENT where NAME = ?) and MODE = ?
`

type EventAccessParams struct {
	Name string
	Mode EventAccessMode
}

type EventAccessRow struct {
	Expression string
	Validity   EventAccessValidity
}

func (q *Queries) EventAccess(ctx context.Context, arg EventAccessParams) ([]EventAccessRow, error) {
	rows, err := q.db.QueryContext(ctx, eventAccess, arg.Name, arg.Mode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []EventAccessRow
	for rows.Next() {
		var i EventAccessRow
		if err := rows.Scan(&i.Expression, &i.Validity); err != nil {
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
select NAME from EVENT
`

func (q *Queries) Events(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, events)
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
    re.ID,
    ire.INCIDENT_NUMBER,
    re.AUTHOR,
    re.TEXT,
    re.CREATED,
    re.GENERATED,
    re.STRICKEN,
    re.ATTACHED_FILE
from
    INCIDENT__REPORT_ENTRY ire
        join REPORT_ENTRY re
             on re.ID = ire.REPORT_ENTRY
where
    ire.EVENT = (select e.ID from EVENT e where e.NAME = ?)
  and re.GENERATED <= ?
`

type Incidents_ReportEntriesParams struct {
	Name      string
	Generated bool
}

type Incidents_ReportEntriesRow struct {
	ID             int32
	IncidentNumber int32
	Author         string
	Text           string
	Created        float64
	Generated      bool
	Stricken       bool
	AttachedFile   sql.NullString
}

func (q *Queries) Incidents_ReportEntries(ctx context.Context, arg Incidents_ReportEntriesParams) ([]Incidents_ReportEntriesRow, error) {
	rows, err := q.db.QueryContext(ctx, incidents_ReportEntries, arg.Name, arg.Generated)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Incidents_ReportEntriesRow
	for rows.Next() {
		var i Incidents_ReportEntriesRow
		if err := rows.Scan(
			&i.ID,
			&i.IncidentNumber,
			&i.Author,
			&i.Text,
			&i.Created,
			&i.Generated,
			&i.Stricken,
			&i.AttachedFile,
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
