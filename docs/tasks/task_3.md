## Task 3: Fix configs.go Generation ✅ COMPLETED

### Issues Fixed
1. ✅ **Removed blank lines** in generated configs.go
   - Added pongo2 whitespace control (`{%-` and `-%}`) in template
   - File: `internal/domain/generators/templates/configs.go.tmpl`

2. ✅ **Alphabetical sorting** for consistent asset ordering
   - Assets in `ProjectAssets` map are now sorted alphabetically by ModelName
   - Model names within each DAG priority group are sorted alphabetically
   - File: `internal/domain/generators/gen_asset_config.go`

### Changes Made
**File 1:** `internal/domain/generators/templates/configs.go.tmpl`
- Lines 7, 9: Added `{%-` whitespace control to `ProjectAssets` loop
- Lines 13, 15, 17, 19: Added `{%-` whitespace control to `DAG` loops

**File 2:** `internal/domain/generators/gen_asset_config.go`
- Added `sort` package import
- Lines 59-64: Sort assets alphabetically by ModelName before rendering
- Lines 66-73: Sort model names within each DAG priority group

### Result
- Generated `configs.go` now has no extra blank lines
- Assets and DAG entries are consistently ordered alphabetically on every generation
- No more random ordering changes between `teal gen` runs