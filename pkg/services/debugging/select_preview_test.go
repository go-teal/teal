package debugging

import "testing"

func TestIsSingleReadOnlySelect(t *testing.T) {
	cases := []struct {
		name string
		sql  string
		want bool
	}{
		{"plain select", "SELECT * FROM raw.stg_leads", true},
		{"trailing semicolon", "select id from t;", true},
		{"with cte", "WITH x AS (SELECT 1) SELECT * FROM x", true},
		{"leading line comment", "-- snapshot\nSELECT 1", true},
		{"leading block comment", "/* profile */ SELECT 1", true},
		{"scd2 script", "UPDATE t SET a=1;\nINSERT INTO t VALUES (1);", false},
		{"multi select", "SELECT 1;\nSELECT 2;", false},
		{"ddl", "CREATE INDEX i ON t(a)", false},
		{"semicolon in literal degrades safely", "SELECT ';' AS c FROM t WHERE x = 1; SELECT 2", false},
	}
	for _, c := range cases {
		if got := isSingleReadOnlySelect(c.sql); got != c.want {
			t.Errorf("%s: isSingleReadOnlySelect(%q) = %v, want %v", c.name, c.sql, got, c.want)
		}
	}
}
