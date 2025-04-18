// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries.sql

package queries

import (
	"context"
	"database/sql"
)

const rangersById = `-- name: RangersById :many
select
    id,
    callsign,
    email,
    status,
    on_site,
    password
from person
where status in ('active', 'inactive', 'inactive extension', 'auditor')
`

type RangersByIdRow struct {
	ID       int64
	Callsign string
	Email    sql.NullString
	Status   PersonStatus
	OnSite   bool
	Password sql.NullString
}

func (q *Queries) RangersById(ctx context.Context) ([]RangersByIdRow, error) {
	rows, err := q.db.QueryContext(ctx, rangersById)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RangersByIdRow
	for rows.Next() {
		var i RangersByIdRow
		if err := rows.Scan(
			&i.ID,
			&i.Callsign,
			&i.Email,
			&i.Status,
			&i.OnSite,
			&i.Password,
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
