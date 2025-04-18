-- name: QueryEventID :one
select ID from EVENT where NAME = ?;

-- name: SchemaVersion :one
select VERSION from SCHEMA_INFO;

-- name: Events :many
select NAME from EVENT;

-- name: CreateEvent :exec
insert into EVENT (NAME) values (?);

-- name: EventAccess :many
select EXPRESSION, VALIDITY from EVENT_ACCESS
where EVENT = (select ID from EVENT where NAME = ?) and MODE = ?;

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
;