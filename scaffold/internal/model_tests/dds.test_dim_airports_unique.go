package modeltests

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_TEST_DIM_AIRPORTS_UNIQUE = `


-- Test for duplicate airport keys
select 
    airport_key, 
    count(*) as duplicate_count 
from dds.dim_airports 
group by airport_key 
having count(*) > 1
`

const COUNT_TEST_SQL_DDS_TEST_DIM_AIRPORTS_UNIQUE = `
select count(*) as test_count from 
(


-- Test for duplicate airport keys
select 
    airport_key, 
    count(*) as duplicate_count 
from dds.dim_airports 
group by airport_key 
having count(*) > 1
) having test_count > 0 limit 1
`


var ddsTestDimAirportsUniqueTestDescriptor = &models.SQLModelTestDescriptor{
	Name: 				"dds.test_dim_airports_unique",
	RawSQL: 			RAW_SQL_DDS_TEST_DIM_AIRPORTS_UNIQUE,
	CountTestSQL: 		COUNT_TEST_SQL_DDS_TEST_DIM_AIRPORTS_UNIQUE,
	TestProfile: 		&configs.TestProfile {
		Name: 				"dds.test_dim_airports_unique",
		Stage: 				"dds",
		Description: 		`IyMg8J+UjSBBaXJwb3J0IERpbWVuc2lvbiBVbmlxdWVuZXNzIFRlc3QKCioqVGVzdCBUeXBlKio6IERhdGEgUXVhbGl0eSAtIFByaW1hcnkgS2V5IENvbnN0cmFpbnQKCioqVmFsaWRhdGlvbiBSdWxlKio6CmBgYHNxbApDT1VOVChESVNUSU5DVCBhaXJwb3J0X2tleSkgPSBDT1VOVCgqKQpgYGAKCioqRmFpbHVyZSBJbXBhY3QqKjoKLSDinYwgRHVwbGljYXRlIGFpcnBvcnRzIGluIGRpbWVuc2lvbgotIOKdjCBJbmNvcnJlY3QgZmxpZ2h0IGNvdW50cyBieSBhaXJwb3J0Ci0g4p2MIFdyb25nIGh1YiBwZXJmb3JtYW5jZSBtZXRyaWNzCi0g4p2MIERpc3RvcnRlZCBuZXR3b3JrIGFuYWx5c2lzCgoqKlJvb3QgQ2F1c2VzIG9mIEZhaWx1cmUqKjoKMS4gSGFzaCBjb2xsaXNpb24gaW4ga2V5IGdlbmVyYXRpb24KMi4gRHVwbGljYXRlIGFpcnBvcnQgY29kZXMgaW4gc291cmNlCjMuIEVUTCBwcm9jZXNzIGVycm9yCgoqKlBhc3MgQ3JpdGVyaWEqKjogWmVybyByb3dzIHJldHVybmVkIChubyBkdXBsaWNhdGVzIGZvdW5kKQ==`,
		Connection: 		"default",
	},
}

var ddsTestDimAirportsUniqueSimpleTestCase = processing.InitSQLModelTesting(ddsTestDimAirportsUniqueTestDescriptor)