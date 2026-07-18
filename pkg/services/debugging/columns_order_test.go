package debugging

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-teal/gota/dataframe"
	"github.com/go-teal/gota/series"
)

// Result rows are serialized as maps, so their JSON keys always come out
// alphabetically sorted. The columns field is the only carrier of the original
// SQL column order — this test protects that contract.
func TestSerializeResultPreservesColumnOrder(t *testing.T) {
	df := dataframe.New(
		series.New([]string{"z1", "z2"}, series.String, "zebra"),
		series.New([]int{1, 2}, series.Int, "id"),
		series.New([]string{"a1", "a2"}, series.String, "alpha"),
	)

	s := &DebuggingService{}
	result, total, columns := s.serializeResultWithPagination(&df, 0, 0)

	if total != 2 {
		t.Fatalf("expected 2 records, got %d", total)
	}
	expected := []string{"zebra", "id", "alpha"}
	if len(columns) != 3 || columns[0] != expected[0] || columns[1] != expected[1] || columns[2] != expected[2] {
		t.Fatalf("columns order lost: got %v, want %v", columns, expected)
	}

	dto := AssetExecuteResponseDTO{Result: result, Columns: columns, TotalRecords: total}
	raw, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"columns":["zebra","id","alpha"]`) {
		t.Fatalf("columns field missing or reordered in JSON: %s", raw)
	}
}
