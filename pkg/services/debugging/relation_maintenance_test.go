package debugging

import (
	"testing"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/models"
)

func descriptorFor(mat configs.MatType) *models.SQLModelDescriptor {
	return &models.SQLModelDescriptor{
		Name:             "dds.some_model",
		DropTableSQL:     "drop table dds.some_model",
		DropViewSQL:      "drop view dds.some_model",
		TruncateTableSQL: "truncate table dds.some_model",
		ModelProfile:     &configs.ModelProfile{Materialization: mat},
	}
}

func TestResolveDropSQL(t *testing.T) {
	cases := []struct {
		mat     configs.MatType
		wantSQL string
		wantOK  bool
	}{
		{configs.MAT_TABLE, "drop table dds.some_model", true},
		{configs.MAT_INCREMENTAL, "drop table dds.some_model", true},
		{configs.MAT_VIEW, "drop view dds.some_model", true},
		{configs.MAT_CUSTOM, "", false},
		{configs.MAT_RAW, "", false},
	}
	for _, c := range cases {
		sql, ok := resolveDropSQL(descriptorFor(c.mat))
		if ok != c.wantOK || sql != c.wantSQL {
			t.Errorf("resolveDropSQL(%s) = (%q, %v), want (%q, %v)", c.mat, sql, ok, c.wantSQL, c.wantOK)
		}
	}
}

func TestResolveTruncateSQL(t *testing.T) {
	cases := []struct {
		mat     configs.MatType
		wantSQL string
		wantOK  bool
	}{
		{configs.MAT_TABLE, "truncate table dds.some_model", true},
		{configs.MAT_INCREMENTAL, "truncate table dds.some_model", true},
		{configs.MAT_VIEW, "", false},
		{configs.MAT_CUSTOM, "", false},
		{configs.MAT_RAW, "", false},
	}
	for _, c := range cases {
		sql, ok := resolveTruncateSQL(descriptorFor(c.mat))
		if ok != c.wantOK || sql != c.wantSQL {
			t.Errorf("resolveTruncateSQL(%s) = (%q, %v), want (%q, %v)", c.mat, sql, ok, c.wantSQL, c.wantOK)
		}
	}
}
