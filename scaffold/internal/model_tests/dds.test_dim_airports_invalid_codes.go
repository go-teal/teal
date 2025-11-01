package modeltests

import (
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_TEST_DIM_AIRPORTS_INVALID_CODES = `
-- Test for airports with invalid IATA codes
-- This test will find airports where code is not exactly 3 characters or is NULL/empty
SELECT
    airport_key,
    airport_name,
    airport_code,
    city,
    country,
    CASE
        WHEN airport_code IS NULL OR airport_code = '' THEN 'NULL or empty code'
        WHEN LENGTH(airport_code) < 3 THEN 'Too short (' || LENGTH(airport_code) || ' chars)'
        WHEN LENGTH(airport_code) > 3 THEN 'Too long (' || LENGTH(airport_code) || ' chars)'
    END as violation_type,
    LENGTH(airport_code) as actual_length
FROM dds.dim_airports
WHERE
    airport_code IS NULL
    OR airport_code = ''
    OR LENGTH(airport_code) != 3
`

const COUNT_TEST_SQL_DDS_TEST_DIM_AIRPORTS_INVALID_CODES = `
select count(*) as test_count from
(
-- Test for airports with invalid IATA codes
-- This test will find airports where code is not exactly 3 characters or is NULL/empty
SELECT
    airport_key,
    airport_name,
    airport_code,
    city,
    country,
    CASE
        WHEN airport_code IS NULL OR airport_code = '' THEN 'NULL or empty code'
        WHEN LENGTH(airport_code) < 3 THEN 'Too short (' || LENGTH(airport_code) || ' chars)'
        WHEN LENGTH(airport_code) > 3 THEN 'Too long (' || LENGTH(airport_code) || ' chars)'
    END as violation_type,
    LENGTH(airport_code) as actual_length
FROM dds.dim_airports
WHERE
    airport_code IS NULL
    OR airport_code = ''
    OR LENGTH(airport_code) != 3
) having test_count > 0 limit 1
`


var ddsTestDimAirportsInvalidCodesTestDescriptor = &models.SQLModelTestDescriptor{
	Name: 				"dds.test_dim_airports_invalid_codes",
	RawSQL: 			RAW_SQL_DDS_TEST_DIM_AIRPORTS_INVALID_CODES,
	CountTestSQL: 		COUNT_TEST_SQL_DDS_TEST_DIM_AIRPORTS_INVALID_CODES,
	TestProfile: 		&configs.TestProfile {
		Name: 				"dds.test_dim_airports_invalid_codes",
		Description: 		`IyMg4pqg77iPIEFpcnBvcnQgSW52YWxpZCBDb2RlIFRlc3QKCioqVGVzdCBUeXBlKio6IERhdGEgUXVhbGl0eSAtIEZvcm1hdCBWYWxpZGF0aW9uCgoqKkJ1c2luZXNzIFJ1bGUqKjogQWxsIGFpcnBvcnRzIG11c3QgaGF2ZSB2YWxpZCAzLWNoYXJhY3RlciBJQVRBIGNvZGVzCgoqKlZhbGlkYXRpb24gTG9naWMqKjoKYGBgc3FsCkxFTkdUSChhaXJwb3J0X2NvZGUpICE9IDMgT1IgYWlycG9ydF9jb2RlIElTIE5VTEwgT1IgYWlycG9ydF9jb2RlID0gJycKYGBgCgoqKkZhaWx1cmUgSW1wYWN0IEFuYWx5c2lzKio6CnwgQXJlYSB8IEltcGFjdCB8IFNldmVyaXR5IHwKfC0tLS0tLXwtLS0tLS0tLXwtLS0tLS0tLS0tfAp8IEZsaWdodCBCb29raW5nIHwgSW52YWxpZCBhaXJwb3J0IHNlbGVjdGlvbiB8IENSSVRJQ0FMIHwKfCBSb3V0ZSBQbGFubmluZyB8IEluY29ycmVjdCByb3V0aW5nIHwgSElHSCB8CnwgSW5kdXN0cnkgU3RhbmRhcmRzIHwgTm9uLWNvbXBsaWFuY2Ugd2l0aCBJQVRBIHwgQ1JJVElDQUwgfAp8IEludGVncmF0aW9uIHwgQVBJIGZhaWx1cmVzIHdpdGggcGFydG5lcnMgfCBISUdIIHwKCioqQ29tbW9uIEZhaWx1cmUgUGF0dGVybnMqKjoKLSBFbXB0eSBvciBOVUxMIGFpcnBvcnQgY29kZXMKLSAyLWNoYXJhY3RlciBkb21lc3RpYyBjb2RlcyAoZS5nLiwgJ0FCJykKLSA0LWNoYXJhY3RlciBJQ0FPIGNvZGVzIG1pc3Rha2VubHkgdXNlZCAoZS5nLiwgJ0FCQ0QnKQotIFNwZWNpYWwgY2hhcmFjdGVycyBvciBzcGFjZXMKCioqRXhhbXBsZSBWaW9sYXRpb25zKio6CmBgYApBaXJwb3J0OiAnQUInIC0gVG9vIHNob3J0ICgyIGNoYXJhY3RlcnMpCkFpcnBvcnQ6ICdBQkNEJyAtIFRvbyBsb25nICg0IGNoYXJhY3RlcnMpCkFpcnBvcnQ6IE5VTEwvZW1wdHkgLSBNaXNzaW5nIGNvZGUKYGBgCgoqKlJlc29sdXRpb24qKjogVmFsaWRhdGUgYXQgZGF0YSBpbmdlc3Rpb24sIGFwcGx5IGRlZmF1bHQgZm9yIHByaXZhdGUgYWlyZmllbGRz`,
		Stage: 				"dds",
		Connection: 		"default",
	},
}

var ddsTestDimAirportsInvalidCodesSimpleTestCase = processing.InitSQLModelTesting(ddsTestDimAirportsInvalidCodesTestDescriptor)
