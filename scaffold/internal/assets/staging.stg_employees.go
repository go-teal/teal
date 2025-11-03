

package assets

import (
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_STAGING_STG_EMPLOYEES = `
select
    employee_id,
    first_name,
    last_name,
    date_of_birth,
    position,
    salary,
    hire_date,
    base_airport
from read_csv('store/employees.csv',
    delim = ',',
    header = true,
    columns = {
        'employee_id': 'INT',
        'first_name': 'VARCHAR',
        'last_name': 'VARCHAR',
        'date_of_birth': 'DATE',
        'position': 'VARCHAR',
        'salary': 'DOUBLE',
        'hire_date': 'DATE',
        'base_airport': 'VARCHAR'
    }
)
`


const SQL_STAGING_STG_EMPLOYEES_CREATE_TABLE = `
create table staging.stg_employees
as (select
    employee_id,
    first_name,
    last_name,
    date_of_birth,
    position,
    salary,
    hire_date,
    base_airport
from read_csv('store/employees.csv',
    delim = ',',
    header = true,
    columns = {
        'employee_id': 'INT',
        'first_name': 'VARCHAR',
        'last_name': 'VARCHAR',
        'date_of_birth': 'DATE',
        'position': 'VARCHAR',
        'salary': 'DOUBLE',
        'hire_date': 'DATE',
        'base_airport': 'VARCHAR'
    }
));

`
const SQL_STAGING_STG_EMPLOYEES_INSERT = `
insert into staging.stg_employees ({{ ModelFields }}) (select
    employee_id,
    first_name,
    last_name,
    date_of_birth,
    position,
    salary,
    hire_date,
    base_airport
from read_csv('store/employees.csv',
    delim = ',',
    header = true,
    columns = {
        'employee_id': 'INT',
        'first_name': 'VARCHAR',
        'last_name': 'VARCHAR',
        'date_of_birth': 'DATE',
        'position': 'VARCHAR',
        'salary': 'DOUBLE',
        'hire_date': 'DATE',
        'base_airport': 'VARCHAR'
    }
))
`
const SQL_STAGING_STG_EMPLOYEES_DROP_TABLE = `
drop table staging.stg_employees
`
const SQL_STAGING_STG_EMPLOYEES_TRUNCATE = `
delete from staging.stg_employees where true;
truncate table staging.stg_employees;
`




var stagingStgEmployeesModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"staging.stg_employees",
	RawSQL: 			RAW_SQL_STAGING_STG_EMPLOYEES,

	CreateTableSQL: 	SQL_STAGING_STG_EMPLOYEES_CREATE_TABLE,
	InsertSQL: 			SQL_STAGING_STG_EMPLOYEES_INSERT,
	DropTableSQL: 		SQL_STAGING_STG_EMPLOYEES_DROP_TABLE,
	TruncateTableSQL: 	SQL_STAGING_STG_EMPLOYEES_TRUNCATE,


	Upstreams: []string {

	},
	Downstreams: []string {

		"dds.dim_employees",

	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"stg_employees",
		Description: 		`IyMgRW1wbG95ZWUgTWFzdGVyIERhdGEgU3RhZ2luZwoKKipQdXJwb3NlKio6IENlbnRyYWwgcmVwb3NpdG9yeSBmb3Igd29ya2ZvcmNlIGluZm9ybWF0aW9uCgoqKkRhdGEgRWxlbWVudHMqKjoKLSBQZXJzb25hbCBkZXRhaWxzIChuYW1lLCBET0IpCi0gUG9zaXRpb24vcm9sZSBjbGFzc2lmaWNhdGlvbgotIENvbXBlbnNhdGlvbiAoc2FsYXJ5KQotIEJhc2UgYWlycG9ydCBhc3NpZ25tZW50Ci0gRW1wbG95bWVudCB0aW1lbGluZSAoaGlyZV9kYXRlKQoKKipEZXJpdmVkIE1ldHJpY3MgU3VwcG9ydCoqOgotIFllYXJzIG9mIHNlcnZpY2UgY2FsY3VsYXRpb24KLSBBZ2UgZm9yIHJlZ3VsYXRvcnkgY29tcGxpYW5jZQotIEJhc2UgZGlzdHJpYnV0aW9uIGFuYWx5c2lzCgoqKkNyaXRpY2FsIGZvcioqOgotIENyZXcgc2NoZWR1bGluZyBvcHRpbWl6YXRpb24KLSBTZW5pb3JpdHktYmFzZWQgYXNzaWdubWVudHMKLSBXb3JrZm9yY2UgY29zdCBhbmFseXNpcw==`,
		Stage: 				"staging",
		Connection: 		"default",
		Materialization: 	"table",
		IsDataFramed: 		false,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {

		},
	},
}

var stagingStgEmployeesAsset processing.Asset = processing.InitSQLModelAsset(stagingStgEmployeesModelDescriptor)
