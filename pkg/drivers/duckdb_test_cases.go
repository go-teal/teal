package drivers

import "database/sql"

func (d *DuckDBEngine) SimpleTest(sqlQuery string) (string, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	var count sql.NullString
	err := d.db.QueryRow(sqlQuery).Scan(&count)

	if err == sql.ErrNoRows {
		return "", nil
	}

	return count.String, err
}
