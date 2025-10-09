
package assets

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_DIM_EMPLOYEES = `


select
    sha256(employee_id::varchar) as employee_key,
    employee_id,
    first_name,
    last_name,
    first_name || ' ' || last_name as full_name,
    date_of_birth,
    date_part('year', age(current_date, date_of_birth)) as age,
    position,
    salary,
    hire_date,
    date_part('year', age(current_date, hire_date)) as years_of_service,
    base_airport,
    sha256(base_airport::varchar) as base_airport_key,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from staging.stg_employees
`
const SQL_DDS_DIM_EMPLOYEES_CREATE_TABLE = `
create table dds.dim_employees 
as (

select
    sha256(employee_id::varchar) as employee_key,
    employee_id,
    first_name,
    last_name,
    first_name || ' ' || last_name as full_name,
    date_of_birth,
    date_part('year', age(current_date, date_of_birth)) as age,
    position,
    salary,
    hire_date,
    date_part('year', age(current_date, hire_date)) as years_of_service,
    base_airport,
    sha256(base_airport::varchar) as base_airport_key,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from staging.stg_employees);
create unique index dim_employees_pkey on dds.dim_employees (employee_key);
create unique index dim_employees_employee_id_idx_idx on dds.dim_employees (employee_id);

`
const SQL_DDS_DIM_EMPLOYEES_INSERT = `
insert into dds.dim_employees ({{ ModelFields }}) (

select
    sha256(employee_id::varchar) as employee_key,
    employee_id,
    first_name,
    last_name,
    first_name || ' ' || last_name as full_name,
    date_of_birth,
    date_part('year', age(current_date, date_of_birth)) as age,
    position,
    salary,
    hire_date,
    date_part('year', age(current_date, hire_date)) as years_of_service,
    base_airport,
    sha256(base_airport::varchar) as base_airport_key,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from staging.stg_employees)
`
const SQL_DDS_DIM_EMPLOYEES_DROP_TABLE = `
drop table dds.dim_employees
`
const SQL_DDS_DIM_EMPLOYEES_TRUNCATE = `
delete from dds.dim_employees where true;
truncate table dds.dim_employees;
`

var ddsDimEmployeesModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"dds.dim_employees",
	RawSQL: 			RAW_SQL_DDS_DIM_EMPLOYEES,
	CreateTableSQL: 	SQL_DDS_DIM_EMPLOYEES_CREATE_TABLE,
	InsertSQL: 			SQL_DDS_DIM_EMPLOYEES_INSERT,
	DropTableSQL: 		SQL_DDS_DIM_EMPLOYEES_DROP_TABLE,
	TruncateTableSQL: 	SQL_DDS_DIM_EMPLOYEES_TRUNCATE,	
	Upstreams: []string {
		"staging.stg_employees",
	},
	Downstreams: []string {
		"dds.fact_crew_assignments",
		"mart.mart_airport_statistics",
		"mart.mart_crew_utilization",
	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"dim_employees",
		Stage: 				"dds",
		Description: 		`IyMgRW1wbG95ZWUgRGltZW5zaW9uIChTQ0QgVHlwZSAxKQoKKipQdXJwb3NlKio6IENvbXByZWhlbnNpdmUgd29ya2ZvcmNlIGRpbWVuc2lvbiB3aXRoIGRlcml2ZWQgYXR0cmlidXRlcwoKKipTdXJyb2dhdGUgS2V5IERlc2lnbioqOgotIFByaW1hcnk6IFNIQTI1NihlbXBsb3llZV9pZCkKLSBGb3JlaWduOiBTSEEyNTYoYmFzZV9haXJwb3J0KSDihpIgZGltX2FpcnBvcnRzCgoqKkNhbGN1bGF0ZWQgRmllbGRzKio6Ci0gYGFnZWA6IEN1cnJlbnQgYWdlIGZvciBjb21wbGlhbmNlCi0gYHllYXJzX29mX3NlcnZpY2VgOiBTZW5pb3JpdHkgY2FsY3VsYXRpb24KLSBgZnVsbF9uYW1lYDogQ29uY2F0ZW5hdGVkIGRpc3BsYXkgbmFtZQoKKipXb3JrZm9yY2UgQW5hbHl0aWNzKio6Ci0g8J+RpSBQb3NpdGlvbiBkaXN0cmlidXRpb24gKHBpbG90LCBjby1waWxvdCwgYXR0ZW5kYW50KQotIPCfk4ogU2VuaW9yaXR5IGFuYWx5c2lzIGZvciBhc3NpZ25tZW50cwotIPCfkrAgQ29tcGVuc2F0aW9uIGJlbmNobWFya2luZwotIPCfj6AgQmFzZSBhaXJwb3J0IGNyZXcgZGlzdHJpYnV0aW9uCgoqKlJlZ3VsYXRvcnkgQ29tcGxpYW5jZSoqOgotIEFnZSByZXN0cmljdGlvbnMgbW9uaXRvcmluZwotIFNlcnZpY2UgeWVhciByZXF1aXJlbWVudHMKLSBCYXNlIGFzc2lnbm1lbnQgcnVsZXMKCioqUXVhbGl0eSBUZXN0cyoqOiBgdGVzdF9kaW1fZW1wbG95ZWVzX3VuaXF1ZWAK`,
		Connection: 		"default",
		Materialization: 	"table",
		IsDataFramed: 		false,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {
			{
				Name: 			"dds.test_dim_employees_unique",		
			},
		},
	},
}

var ddsDimEmployeesAsset processing.Asset = processing.InitSQLModelAsset(ddsDimEmployeesModelDescriptor)