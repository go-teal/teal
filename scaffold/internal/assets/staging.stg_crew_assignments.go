

package assets

import (
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_STAGING_STG_CREW_ASSIGNMENTS = `
select
    assignment_id,
    flight_id,
    employee_id,
    role_on_flight,
    assignment_date
from read_csv('store/crew_assignments.csv',
    delim = ',',
    header = true,
    columns = {
        'assignment_id': 'INT',
        'flight_id': 'INT',
        'employee_id': 'INT',
        'role_on_flight': 'VARCHAR',
        'assignment_date': 'DATE'
    }
)
`


const SQL_STAGING_STG_CREW_ASSIGNMENTS_CREATE_TABLE = `
create table staging.stg_crew_assignments
as (select
    assignment_id,
    flight_id,
    employee_id,
    role_on_flight,
    assignment_date
from read_csv('store/crew_assignments.csv',
    delim = ',',
    header = true,
    columns = {
        'assignment_id': 'INT',
        'flight_id': 'INT',
        'employee_id': 'INT',
        'role_on_flight': 'VARCHAR',
        'assignment_date': 'DATE'
    }
));

`
const SQL_STAGING_STG_CREW_ASSIGNMENTS_INSERT = `
insert into staging.stg_crew_assignments ({{ ModelFields }}) (select
    assignment_id,
    flight_id,
    employee_id,
    role_on_flight,
    assignment_date
from read_csv('store/crew_assignments.csv',
    delim = ',',
    header = true,
    columns = {
        'assignment_id': 'INT',
        'flight_id': 'INT',
        'employee_id': 'INT',
        'role_on_flight': 'VARCHAR',
        'assignment_date': 'DATE'
    }
))
`
const SQL_STAGING_STG_CREW_ASSIGNMENTS_DROP_TABLE = `
drop table staging.stg_crew_assignments
`
const SQL_STAGING_STG_CREW_ASSIGNMENTS_TRUNCATE = `
delete from staging.stg_crew_assignments where true;
truncate table staging.stg_crew_assignments;
`




var stagingStgCrewAssignmentsModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"staging.stg_crew_assignments",
	RawSQL: 			RAW_SQL_STAGING_STG_CREW_ASSIGNMENTS,

	CreateTableSQL: 	SQL_STAGING_STG_CREW_ASSIGNMENTS_CREATE_TABLE,
	InsertSQL: 			SQL_STAGING_STG_CREW_ASSIGNMENTS_INSERT,
	DropTableSQL: 		SQL_STAGING_STG_CREW_ASSIGNMENTS_DROP_TABLE,
	TruncateTableSQL: 	SQL_STAGING_STG_CREW_ASSIGNMENTS_TRUNCATE,


	Upstreams: []string {

	},
	Downstreams: []string {

		"dds.fact_crew_assignments",

	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"stg_crew_assignments",
		Description: 		`IyMgQ3JldyBBc3NpZ25tZW50IFN0YWdpbmcgTGF5ZXIKCioqUHVycG9zZSoqOiBNYXBzIGVtcGxveWVlcyB0byBmbGlnaHRzIHdpdGggcm9sZSBzcGVjaWZpY2F0aW9ucwoKKipLZXkgRmVhdHVyZXMqKjoKLSBFc3RhYmxpc2hlcyBtYW55LXRvLW1hbnkgcmVsYXRpb25zaGlwIGJldHdlZW4gZmxpZ2h0cyBhbmQgY3JldwotIFRyYWNrcyByb2xlIGFzc2lnbm1lbnRzIChwaWxvdCwgY28tcGlsb3QsIGZsaWdodCBhdHRlbmRhbnQpCi0gUmVjb3JkcyBhc3NpZ25tZW50IGRhdGVzIGZvciBzY2hlZHVsaW5nIGFuYWx5c2lzCgoqKkRhdGEgTW9kZWwqKjoKLSBMaW5rcyB0bzogYGVtcGxveWVlX2lkYCwgYGZsaWdodF9pZGAKLSBQcmltYXJ5IGtleTogYGFzc2lnbm1lbnRfaWRgCgoqKkJ1c2luZXNzIFZhbHVlKio6Ci0gRW5hYmxlcyBjcmV3IHV0aWxpemF0aW9uIGFuYWx5c2lzCi0gU3VwcG9ydHMgcmVndWxhdG9yeSBjb21wbGlhbmNlIHRyYWNraW5nCi0gRm91bmRhdGlvbiBmb3Igd29ya2xvYWQgZGlzdHJpYnV0aW9uIG1ldHJpY3M=`,
		Stage: 				"staging",
		Connection: 		"default",
		Materialization: 	"table",
		IsDataFramed: 		false,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {

		},
	},
}

var stagingStgCrewAssignmentsAsset processing.Asset = processing.InitSQLModelAsset(stagingStgCrewAssignmentsModelDescriptor)
