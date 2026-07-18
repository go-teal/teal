# Issue #005 — int4/int8 колонки Postgres теряются в DataFrame (везде null)

**Severity:** High — любые integer/bigint колонки (`amocrm_lead_id`, счётчики,
внешние ID) отображаются как `null` в data-preview дашборда `teal ui` и в любых
cross-DB dataframe-потоках.

**Discovered:** 2026-07-18 при исследовании данных `partner-analytics`
(raw.stg_leads: `amocrm_lead_id` «пуст» при заполненной колонке в БД).

**Status:** ✅ RESOLVED — фикс в `fix/pg-dataframe-int-slices` (PR #21), релиз **v1.2.3** (v1.2.2 успел выйти параллельно с фиксом column-order — v1.2.3 содержит оба).

---

## Symptom

`POST /api/dag/asset/:name/select` + `GET /api/dag/asset/:name/data` возвращают
`"amocrm_lead_id": null` для всех строк, хотя `SELECT` в psql показывает значения.
Данные в БД целы; ломается только конвертация в DataFrame.

## Root cause

`pkg/drivers/postgres_dataframe.go` (ToDataFrame) накапливал колонки как:

- `int4` → `[]int32`
- `int8` → `[]int64`

а `series.New` нашего gota-форка поддерживает только `[]int` (кейсов `[]int32` /
`[]int64` нет) — неизвестный тип слайса деградирует в NA-элементы → `null` в JSON.
`int2` работал случайно (уже собирался в `[]int`).

## Fix

Все integer-типы (`int2`/`int4`/`int8`) накапливаются в `[]int` с конверсией
из соответствующих `sql.NullInt*`. `int` 64-битный на всех release-таргетах
(linux/darwin amd64/arm64) — конверсия из int64 без потерь.

Проверено против живой БД: `SELECT amocrm_lead_id ...` → DataFrame показывает
реальные значения `<int>`.

## Notes

- DuckDB-драйвер использует собственный маппинг — проверить аналогичный паттерн
  отдельно, если DuckDB вернётся в работу.
- Правильный фикс «по-хорошему» — поддержка `[]int32`/`[]int64` в gota-форке;
  текущий фикс закрывает проблему на стороне teal без релиза gota.
