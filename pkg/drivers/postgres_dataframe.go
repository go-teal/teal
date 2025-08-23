package drivers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-teal/gota/dataframe"
	"github.com/go-teal/gota/series"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var pgOIDToType = map[int]string{
	16:    "bool",
	17:    "bytea",
	18:    "char",
	19:    "name",
	20:    "int8",
	21:    "int2",
	22:    "int2vector",
	23:    "int4",
	24:    "regproc",
	25:    "text",
	26:    "oid",
	27:    "tid",
	28:    "xid",
	29:    "cid",
	30:    "oidvector",
	114:   "json",
	142:   "xml",
	5069:  "xid8",
	600:   "point",
	601:   "lseg",
	602:   "path",
	603:   "box",
	604:   "polygon",
	628:   "line",
	700:   "float4",
	701:   "float8",
	705:   "unknown",
	718:   "circle",
	790:   "money",
	829:   "macaddr",
	869:   "inet",
	650:   "cidr",
	774:   "macaddr8",
	1033:  "aclitem",
	1042:  "bpchar",
	1043:  "varchar",
	1082:  "date",
	1083:  "time",
	1114:  "timestamp",
	1184:  "timestamptz",
	1186:  "interval",
	1266:  "timetz",
	1560:  "bit",
	1562:  "varbit",
	1700:  "numeric",
	1790:  "refcursor",
	2202:  "regprocedure",
	2203:  "regoper",
	2204:  "regoperator",
	2205:  "regclass",
	4191:  "regcollation",
	2206:  "regtype",
	4096:  "regrole",
	4089:  "regnamespace",
	2950:  "uuid",
	3614:  "tsvector",
	3642:  "gtsvector",
	3615:  "tsquery",
	3734:  "regconfig",
	3769:  "regdictionary",
	3802:  "jsonb",
	4072:  "jsonpath",
	3904:  "int4range",
	3906:  "numrange",
	3908:  "tsrange",
	3910:  "tstzrange",
	3912:  "daterange",
	3926:  "int8range",
	4451:  "int4multirange",
	4532:  "nummultirange",
	4533:  "tsmultirange",
	4534:  "tstzmultirange",
	4535:  "datemultirange",
	4536:  "int8multirange",
	2249:  "record",
	2275:  "cstring",
	2276:  "any",
	2277:  "anyarray",
	2278:  "void",
	2279:  "trigger",
	2281:  "internal",
	2283:  "anyelement",
	2776:  "anynonarray",
	3500:  "anyenum",
	3831:  "anyrange",
	5077:  "anycompatible",
	5078:  "anycompatiblearray",
	5079:  "anycompatiblenonarray",
	5080:  "anycompatiblerange",
	4537:  "anymultirange",
	4538:  "anycompatiblemultirange",
	13259: "attributes",
	13279: "collations",
	13309: "columns",
	13333: "domains",
	13347: "parameters",
	13390: "routines",
	13395: "schemata",
	13399: "sequences",
	13438: "tables",
	13443: "transforms",
	13453: "triggers",
	13496: "views",
}

// ToDataFrame implements PGDriver.
func (d *PostgresDBEngine) ToDataFrame(sqlQuery string) (*dataframe.DataFrame, error) {
	rows, err := d.db.Query(context.Background(), sqlQuery)
	if err != nil {
		log.Error().Caller().Stack().Err(err).Msg(sqlQuery)
		return nil, err
	}
	columnTypes := rows.FieldDescriptions()
	seriesData := make([]interface{}, len(columnTypes))
	for i, c := range columnTypes {
		pgType := pgOIDToType[int(c.DataTypeOID)]
		switch pgType {
		case "varchar", "text":
			seriesData[i] = make([]string, 0)
		case "numeric", "float4", "float8":
			seriesData[i] = make([]float64, 0)
		case "int2":
			seriesData[i] = make([]int, 0)
		case "int4":
			seriesData[i] = make([]int, 0)
		case "bool":
			seriesData[i] = make([]string, 0)
		default:
			seriesData[i] = make([]string, 0)
			log.Warn().Str("type", pgType).Str("field", c.Name).Msg("type not implemented in Dataframe")
		}
	}

	// var rowNumber int = 0
	for rows.Next() {
		// rowNumber++
		safeData := make([]interface{}, len(columnTypes))
		for i, c := range columnTypes {
			pgType := pgOIDToType[int(c.DataTypeOID)]
			switch pgType {
			case "varchar", "text":
				safeData[i] = &sql.NullString{}
			case "numeric", "float4", "float8":
				safeData[i] = &sql.NullFloat64{}
			case "int2":
				safeData[i] = &sql.NullInt16{}
			case "int4":
				safeData[i] = &sql.NullInt32{}
			case "bool":
				safeData[i] = &sql.NullBool{}
			default:
				safeData[i] = &sql.NullString{}
			}
		}
		err := rows.Scan(safeData...)
		if err != nil {
			log.Error().Caller().Stack().Err(err).Msg("PostgreSQL Scan error")
			return nil, err
		}

		for i, c := range columnTypes {
			pgType := pgOIDToType[int(c.DataTypeOID)]
			switch pgType {

			case "varchar", "text":
				sd := seriesData[i].([]string)
				val := safeData[i].(*sql.NullString)
				sd = append(sd, val.String)
				seriesData[i] = sd

			case "numeric", "float4", "float8":
				sd := seriesData[i].([]float64)
				val := safeData[i].(*sql.NullFloat64)
				sd = append(sd, val.Float64)
				seriesData[i] = sd

			case "int2":
				sd := seriesData[i].([]int)
				val := safeData[i].(*sql.NullInt16)
				sd = append(sd, int(val.Int16))
				seriesData[i] = sd
			case "int4":
				sd := seriesData[i].([]int)
				val := safeData[i].(*sql.NullInt32)
				sd = append(sd, int(val.Int32))
				seriesData[i] = sd
			case "bool":
				sd := seriesData[i].([]bool)
				val := safeData[i].(*sql.NullBool)
				sd = append(sd, val.Bool)
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
		pgType := pgOIDToType[int(c.DataTypeOID)]
		switch pgType {
		case "varchar", "text":
			dFseries[i] = series.New(seriesData[i], series.String, c.Name)
		case "numeric", "float4", "float8":
			dFseries[i] = series.New(seriesData[i], series.Float, c.Name)
		case "int2", "int4":
			dFseries[i] = series.New(seriesData[i], series.Int, c.Name)
		case "bool":
			dFseries[i] = series.New(seriesData[i], series.String, c.Name)
		default:
			log.Warn().Str("type", pgType).Str("field", c.Name).Msg("type not implemented")
			dFseries[i] = series.New(seriesData[i], series.String, c.Name)
		}
	}

	df := dataframe.New(dFseries...)
	// log.Debug().Msg(df.String())
	return &df, nil
}

// PersistDataFrame implements PGDriver.
func (d *PostgresDBEngine) PersistDataFrame(tx interface{}, name string, df *dataframe.DataFrame) error {
	log.Debug().Str("name", name).Msg("Persisting DataFrame")
	query := fmt.Sprintf("create temp table %s (\n", name)
	colTypes := df.Types()
	colNames := df.Names()
	columnsPartExpression := make([]string, len(colNames))
	for colIdx, colName := range colNames {
		colType := colTypes[colIdx]
		if colType == "string" {
			colType = "text"
		}
		columnsPartExpression[colIdx] = fmt.Sprintf("	%s %s", colName, colType)
	}

	query += strings.Join(columnsPartExpression, ",\n")
	query += "\n);\n"
	nRows, _ := df.Dims()
	for rowIdx := 0; rowIdx < nRows; rowIdx++ {
		vals := make([]string, len(colTypes))
		for colIdx, colType := range colTypes {
			switch colType {
			case series.String:
				val := df.Elem(rowIdx, colIdx).String()
				val = strings.ReplaceAll(val, "'", "''")
				val = strings.ReplaceAll(val, "\"", "\"")
				vals[colIdx] = fmt.Sprintf("'%s'", val)
			case series.Float:
				val := df.Elem(rowIdx, colIdx).Float()
				vals[colIdx] = fmt.Sprintf("%f", val)
			case series.Int:
				val, err := df.Elem(rowIdx, colIdx).Int()
				if err != nil {
					log.Error().Caller().Stack().Err(err).Msg("val, err := df.Elem(rowIdx, colIdx).Int()")
					return err
				}
				vals[colIdx] = fmt.Sprintf("%d", val)
			case series.Bool:
				val, err := df.Elem(rowIdx, colIdx).Bool()
				if err != nil {
					log.Error().Caller().Stack().Err(err).Msg("val, err := df.Elem(rowIdx, colIdx).Bool()")
					return err
				}
				vals[colIdx] = fmt.Sprintf("%t, ", val)
			default:
				return fmt.Errorf("type %s not implemented", colType)
			}
		}
		query += fmt.Sprintf("insert into %s(%s) values(%s);\n", name, strings.Join(colNames, ", "), strings.Join(vals, ", "))
	}
	log.Debug().Msg(query)
	_, err := tx.(pgx.Tx).Exec(context.Background(), query)
	return err
}
