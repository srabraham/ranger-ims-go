-- name: RangersById :many
select
    id,
    callsign,
    email,
    status,
    on_site,
    password
from person
where status in ('active', 'inactive', 'inactive extension', 'auditor')