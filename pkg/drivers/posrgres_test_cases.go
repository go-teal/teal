package drivers

import (
	"context"
	"database/sql"
)

func (d *PostgresDBEngine) SimpleTest(sqlQuery string) (string, error) {
	var count sql.NullString
	err := d.db.QueryRow(context.Background(), sqlQuery).Scan(&count)

	if err == sql.ErrNoRows {
		return "", nil
	}

	return count.String, err
}
