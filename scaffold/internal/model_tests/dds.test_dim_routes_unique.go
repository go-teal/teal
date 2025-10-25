package modeltests

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_TEST_DIM_ROUTES_UNIQUE = `


-- Test for duplicate route keys
select 
    route_key, 
    count(*) as duplicate_count 
from dds.dim_routes 
group by route_key 
having count(*) > 1
`

const COUNT_TEST_SQL_DDS_TEST_DIM_ROUTES_UNIQUE = `
select count(*) as test_count from 
(


-- Test for duplicate route keys
select 
    route_key, 
    count(*) as duplicate_count 
from dds.dim_routes 
group by route_key 
having count(*) > 1
) having test_count > 0 limit 1
`


var ddsTestDimRoutesUniqueTestDescriptor = &models.SQLModelTestDescriptor{
	Name: 				"dds.test_dim_routes_unique",
	RawSQL: 			RAW_SQL_DDS_TEST_DIM_ROUTES_UNIQUE,
	CountTestSQL: 		COUNT_TEST_SQL_DDS_TEST_DIM_ROUTES_UNIQUE,
	TestProfile: 		&configs.TestProfile {
		Name: 				"dds.test_dim_routes_unique",
		Stage: 				"dds",
		Description: 		`IyMg8J+UjSBSb3V0ZSBEaW1lbnNpb24gVW5pcXVlbmVzcyBUZXN0CgoqKlRlc3QgVHlwZSoqOiBOZXR3b3JrIEludGVncml0eSBWYWxpZGF0aW9uCgoqKkJ1c2luZXNzIFJ1bGUqKjogT25lIHJvdXRlIHBlciBvcmlnaW4tZGVzdGluYXRpb24gcGFpcgoKKipLZXkgR2VuZXJhdGlvbioqOgpgYGAKcm91dGVfa2V5ID0gU0hBMjU2KG9yaWdpbiB8fCAnLScgfHwgZGVzdGluYXRpb24pCmBgYAoKKipEdXBsaWNhdGUgSW1wYWN0IEFuYWx5c2lzKio6CnwgQXJlYSB8IEltcGFjdCB8IFNldmVyaXR5IHwKfC0tLS0tLXwtLS0tLS0tLXwtLS0tLS0tLS0tfAp8IFBlcmZvcm1hbmNlIE1ldHJpY3MgfCBTcGxpdCBhY3Jvc3MgZHVwbGljYXRlcyB8IEhJR0ggfAp8IENhcGFjaXR5IFBsYW5uaW5nIHwgVW5kZXJlc3RpbWF0ZWQgdXNhZ2UgfCBISUdIIHwKfCBSZXZlbnVlIEFuYWx5c2lzIHwgSW5jb3JyZWN0IHJvdXRlIHByb2ZpdGFiaWxpdHkgfCBDUklUSUNBTCB8CnwgTmV0d29yayBPcHRpbWl6YXRpb24gfCBXcm9uZyByb3V0aW5nIGRlY2lzaW9ucyB8IE1FRElVTSB8CgoqKkNvbW1vbiBGYWlsdXJlIFBhdHRlcm5zKio6Ci0gU2FtZSByb3V0ZSB3aXRoIGRpZmZlcmVudCByb3V0ZV9pZHMKLSBCaWRpcmVjdGlvbmFsIHJvdXRlcyBjb3VudGVkIHNlcGFyYXRlbHkKLSBEYXRhIGludGVncmF0aW9uIGVycm9ycwoKKipSZXNvbHV0aW9uKio6IERlZHVwIGF0IHNvdXJjZSBvciBpbiBzdGFnaW5n`,
		Connection: 		"default",
	},
}

var ddsTestDimRoutesUniqueSimpleTestCase = processing.InitSQLModelTesting(ddsTestDimRoutesUniqueTestDescriptor)