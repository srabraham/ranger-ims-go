-- name: QueryEventID :one
select sqlc.embed(e) from EVENT e where e.NAME = ?;

-- name: SchemaVersion :one
select VERSION from SCHEMA_INFO;

-- name: Events :many
select sqlc.embed(e) from EVENT e;

-- name: CreateEvent :execlastid
insert into EVENT (NAME) values (?);

-- name: EventAccess :many
select sqlc.embed(ea)
from EVENT_ACCESS ea
where ea.EVENT = ?;

-- name: EventAccessAll :many
select sqlc.embed(ea)
from EVENT_ACCESS ea
;

-- name: ClearEventAccessForMode :exec
delete from EVENT_ACCESS
where EVENT = ? and MODE = ?;

-- name: ClearEventAccessForExpression :exec
delete from EVENT_ACCESS
where EVENT = ? and EXPRESSION = ?;

-- name: AddEventAccess :execlastid
insert into EVENT_ACCESS (EVENT, EXPRESSION, MODE, VALIDITY)
values (?, ?, ?, ?);

-- name: CreateIncident :execlastid
insert into INCIDENT (
    EVENT,
    NUMBER,
    CREATED,
    PRIORITY,
    STATE
)
values (
   ?,?,?,?,?
);

-- name: UpdateIncident :exec
update INCIDENT set
    CREATED = ?,
    PRIORITY = ?,
    STATE = ?,
    SUMMARY = ?,
    LOCATION_NAME = ?,
    LOCATION_CONCENTRIC = ?,
    LOCATION_RADIAL_HOUR = ?,
    LOCATION_RADIAL_MINUTE = ?,
    LOCATION_DESCRIPTION = ?
where
    EVENT = ?
    and NUMBER = ?
;

-- name: Incident :one
select
    sqlc.embed(i),
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
from INCIDENT i
where i.EVENT = ?
    and i.NUMBER = ?;

-- name: Incidents :many
select
    sqlc.embed(i),
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
    i.EVENT = ?
group by
    i.NUMBER;

-- name: Incidents_ReportEntries :many
select
    ire.INCIDENT_NUMBER,
    sqlc.embed(re)
from
    INCIDENT__REPORT_ENTRY ire
        join REPORT_ENTRY re
             on re.ID = ire.REPORT_ENTRY
where
    ire.EVENT = ?
    and re.GENERATED <= ?
;

-- name: Incident_ReportEntries :many
select
    ire.INCIDENT_NUMBER,
    sqlc.embed(re)
from
    INCIDENT__REPORT_ENTRY ire
        join REPORT_ENTRY re
             on re.ID = ire.REPORT_ENTRY
where
    ire.EVENT = ?
    and ire.INCIDENT_NUMBER = ?
;

-- name: ConcentricStreets :many
select sqlc.embed(cs)
from CONCENTRIC_STREET cs
where cs.EVENT = ?;

-- name: IncidentTypes :many
select sqlc.embed(it)
from INCIDENT_TYPE it;

-- name: FieldReports :many
select sqlc.embed(fr)
from FIELD_REPORT fr
where fr.EVENT = ?;

-- name: FieldReport :one
select sqlc.embed(fr)
from FIELD_REPORT fr
where fr.EVENT = ?
    and fr.NUMBER = ?;

-- name: FieldReports_ReportEntries :many
select
    irre.FIELD_REPORT_NUMBER,
    sqlc.embed(re)
from
    FIELD_REPORT__REPORT_ENTRY irre
        join REPORT_ENTRY re
             on irre.REPORT_ENTRY = re.ID
where
    irre.EVENT = ?
    and re.GENERATED <= ?
;

-- name: FieldReport_ReportEntries :many
select
    sqlc.embed(re)
from
    FIELD_REPORT__REPORT_ENTRY irre
        join REPORT_ENTRY re
             on irre.REPORT_ENTRY = re.ID
where
    irre.EVENT = ?
    and irre.FIELD_REPORT_NUMBER = ?
;

-- name: AttachFieldReportToIncident :exec
update FIELD_REPORT
set INCIDENT_NUMBER = ?
where EVENT = ? and NUMBER = ?
;

-- name: MaxFieldReportNumber :one
select coalesce(max(NUMBER), 0) from FIELD_REPORT
where EVENT = ?;

-- name: MaxIncidentNumber :one
select coalesce(max(NUMBER), 0) from INCIDENT
where EVENT = ?;

-- name: CreateFieldReport :exec
insert into FIELD_REPORT (
    EVENT, NUMBER, CREATED, SUMMARY, INCIDENT_NUMBER
)
values (?, ?, ?, ?, ?);

-- name: DetachedFieldReportNumbers :many
select NUMBER from FIELD_REPORT
where EVENT = ? and INCIDENT_NUMBER is null;

-- name: AttachedFieldReportNumbers :many
select NUMBER from FIELD_REPORT
where
    EVENT = ? and
    INCIDENT_NUMBER = ?;

-- name: UpdateFieldReport :exec
update FIELD_REPORT
set SUMMARY = ?, INCIDENT_NUMBER = ?
where EVENT = ? and NUMBER = ?;

-- name: CreateReportEntry :execlastid
insert into REPORT_ENTRY (
    AUTHOR, TEXT, CREATED, `GENERATED`, STRICKEN, ATTACHED_FILE
) values (
   ?, ?, ?, ?, ?, ?
);

-- name: AttachReportEntryToFieldReport :exec
insert into FIELD_REPORT__REPORT_ENTRY (
    EVENT, FIELD_REPORT_NUMBER, REPORT_ENTRY
) values (
    ?, ?, ?
);

-- name: AttachReportEntryToIncident :exec
insert into INCIDENT__REPORT_ENTRY (
    EVENT, INCIDENT_NUMBER, REPORT_ENTRY
) values (
    ?, ?, ?
);

/*
   The "stricken" queries seem bloated at first blush, because the whole
   "where ID in (..." could just be "where ID =". What it's doing though is
   ensuring that the provided eventID and incidentNumber actually align with
   the reportEntryID in question, and that's important for authorization purposes.
*/

-- name: SetIncidentReportEntryStricken :exec
update REPORT_ENTRY
set STRICKEN = ?
where ID IN (
    select REPORT_ENTRY
    from INCIDENT__REPORT_ENTRY
    where EVENT = ?
        and INCIDENT_NUMBER = ?
        and REPORT_ENTRY = ?
);

-- name: SetFieldReportReportEntryStricken :exec
update REPORT_ENTRY
set STRICKEN = ?
where ID IN (
    select REPORT_ENTRY
    from FIELD_REPORT__REPORT_ENTRY
    where EVENT = ?
      and FIELD_REPORT_NUMBER = ?
      and REPORT_ENTRY = ?
);

-- name: AttachRangerHandleToIncident :exec
insert into INCIDENT__RANGER (EVENT, INCIDENT_NUMBER, RANGER_HANDLE)
values (?, ?, ?);

-- name: DetachRangerHandleFromIncident :exec
delete from INCIDENT__RANGER
where
    EVENT = ?
    and INCIDENT_NUMBER = ?
    and RANGER_HANDLE = ?
;

-- name: AttachIncidentTypeToIncident :exec
insert into INCIDENT__INCIDENT_TYPE (
    EVENT, INCIDENT_NUMBER, INCIDENT_TYPE
) values (
    ?, ?, (select it.ID from INCIDENT_TYPE it where it.NAME = ?)
);

-- name: DetachIncidentTypeFromIncident :exec
delete from INCIDENT__INCIDENT_TYPE
where
    EVENT = ?
    and INCIDENT_NUMBER = ?
    and INCIDENT_TYPE = (select it.ID from INCIDENT_TYPE it where it.NAME = ?)
;

-- name: CreateIncidentTypeOrIgnore :exec
insert into INCIDENT_TYPE (NAME, HIDDEN)
values (?, ?)
    on duplicate key update NAME=NAME
;

-- name: HideShowIncidentType :exec
update INCIDENT_TYPE set HIDDEN = ?
where NAME = ?;

-- name: CreateConcentricStreet :exec
insert into CONCENTRIC_STREET (EVENT, ID, NAME)
values (?, ?, ?);
