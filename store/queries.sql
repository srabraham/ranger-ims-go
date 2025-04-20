-- name: QueryEventID :one
select ID from EVENT where NAME = ?;

-- name: SchemaVersion :one
select VERSION from SCHEMA_INFO;

-- name: Events :many
select ID, NAME from EVENT;

-- name: CreateEvent :execlastid
insert into EVENT (NAME) values (?);

-- name: EventAccess :many
select EXPRESSION, MODE, VALIDITY from EVENT_ACCESS
where EVENT = ?;

-- name: CreateIncident :execlastid
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
);

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
    i.EVENT = (select e.ID from EVENT e where e.NAME = ?)
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
select ID, NAME from CONCENTRIC_STREET
where EVENT = ?;

-- name: IncidentTypes :many
select NAME, HIDDEN from INCIDENT_TYPE;

-- name: FieldReports :many
select
    sqlc.embed(FIELD_REPORT)
from
    FIELD_REPORT
where
    EVENT = ?;

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
    and re.GENERATED <= ?;
