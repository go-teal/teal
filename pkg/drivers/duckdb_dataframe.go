package drivers

import (
	"database/sql"

	"github.com/go-teal/gota/dataframe"
	"github.com/go-teal/gota/series"
	"github.com/rs/zerolog/log"
)

// ToDataFrame implements DBDriver.
func (d *DuckDBEngine) ToDataFrame(sqlQuery string) (*dataframe.DataFrame, error) {
	rows, err := d.db.Query(sqlQuery)
	if err != nil {
		log.Error().Stack().Err(err).Msg(sqlQuery)
		return nil, err
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		log.Error().Stack().Err(err).Msg("Can not extract column types")
		return nil, err
	}
	seriesData := make([]interface{}, len(columnTypes))
	for i, c := range columnTypes {
		switch c.DatabaseTypeName() {
		case "VARCHAR":
			seriesData[i] = make([]string, 0)
		case "DOUBLE":
			seriesData[i] = make([]float64, 0)
		case "HUGEINT":
			// TODO: Add this type to gota series
			seriesData[i] = make([]string, 0)
		case "INTEGER":
			seriesData[i] = make([]int, 0)
		case "TIMESTAMP":
			// TODO: Add this type to gota series
			seriesData[i] = make([]string, 0)
		case "BIGINT":
			// TODO: Add this type to gota series
			seriesData[i] = make([]string, 0)
		case "BOOLEAN":
			seriesData[i] = make([]string, 0)
		case "DATE":
			// TODO: Add this type to gota series
			seriesData[i] = make([]string, 0)
		default:
			seriesData[i] = make([]string, 0)
			log.Warn().Str("type", c.DatabaseTypeName()).Str("field", c.Name()).Msg("type not implemented")
		}
	}

	var rowNumber int = 0
	for rows.Next() {
		rowNumber++
		safeData := make([]interface{}, len(columnTypes))
		for i, c := range columnTypes {
			switch c.DatabaseTypeName() {
			case "VARCHAR":
				safeData[i] = &sql.NullString{}
			case "DOUBLE":
				safeData[i] = &sql.NullFloat64{}
			case "HUGEINT":
				safeData[i] = &sql.NullString{}
			case "INTEGER":
				safeData[i] = &sql.NullInt32{}
			case "TIMESTAMP":
				safeData[i] = &sql.NullString{}
			case "BIGINT":
				safeData[i] = &sql.NullInt64{}
			case "BOOLEAN":
				safeData[i] = &sql.NullBool{}
			case "DATE":
				safeData[i] = &sql.NullString{}
			default:
				safeData[i] = &sql.NullString{}
			}
		}
		err := rows.Scan(safeData...)
		if err != nil {
			log.Error().Stack().Err(err).Msg("DuckDB Scan error")
			return nil, err
		}

		for i, c := range columnTypes {
			switch c.DatabaseTypeName() {

			case "VARCHAR":
				sd := seriesData[i].([]string)
				val := safeData[i].(*sql.NullString)
				sd = append(sd, val.String)
				seriesData[i] = sd

			case "DOUBLE":
				sd := seriesData[i].([]float64)
				val := safeData[i].(*sql.NullFloat64)
				sd = append(sd, val.Float64)
				seriesData[i] = sd

			case "HUGEINT":
				sd := seriesData[i].([]string)
				val := safeData[i].(*sql.NullString)
				sd = append(sd, val.String)
				seriesData[i] = sd

			case "INTEGER":
				sd := seriesData[i].([]int)
				val := safeData[i].(*sql.NullInt32)
				sd = append(sd, int(val.Int32))
				seriesData[i] = sd

			case "TIMESTAMP":
				sd := seriesData[i].([]string)
				val := safeData[i].(*sql.NullString)
				sd = append(sd, val.String)
				seriesData[i] = sd

			case "BIGINT":
				sd := seriesData[i].([]int64)
				val := safeData[i].(*sql.NullInt64)
				sd = append(sd, val.Int64)
				seriesData[i] = sd

			case "BOOLEAN":
				sd := seriesData[i].([]bool)
				val := safeData[i].(*sql.NullBool)
				sd = append(sd, val.Bool)
				seriesData[i] = sd

			case "DATE":
				sd := seriesData[i].([]string)
				val := safeData[i].(*sql.NullString)
				sd = append(sd, val.String)
				seriesData[i] = sd

			default:
				sd := seriesData[i].([]string)
				val := safeData[i].(*sql.NullString)
				sd = append(sd, val.String)
				seriesData[i] = sd
			}
		}
	}

	dFseries := make([]series.Series, len(columnTypes))
	for i, c := range columnTypes {
		switch c.DatabaseTypeName() {
		case "VARCHAR":
			dFseries[i] = series.New(seriesData[i], series.String, c.Name())
		case "DOUBLE":
			dFseries[i] = series.New(seriesData[i], series.Float, c.Name())
		case "HUGEINT":
			// TODO: Add this type to gota series
			dFseries[i] = series.New(seriesData[i], series.String, c.Name())
		case "INTEGER":
			dFseries[i] = series.New(seriesData[i], series.Int, c.Name())
		case "TIMESTAMP":
			// TODO: Add this type to gota series
			dFseries[i] = series.New(seriesData[i], series.String, c.Name())
		case "BIGINT":
			// TODO: Add this type to gota series
			dFseries[i] = series.New(seriesData[i], series.String, c.Name())
		case "BOOLEAN":
			dFseries[i] = series.New(seriesData[i], series.String, c.Name())
		case "DATE":
			dFseries[i] = series.New(seriesData[i], series.String, c.Name())
		default:
			log.Warn().Str("type", c.DatabaseTypeName()).Str("field", c.Name()).Msg("type not implemented")
			dFseries[i] = series.New(seriesData[i], series.String, c.Name())
		}
	}

	df := dataframe.New(dFseries...)
	// log.Debug().Msg(df.String())
	return &df, nil
}
