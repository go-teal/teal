# Issue #008 — `CREATE SCHEMA` без `IF NOT EXISTS`: гонка при первом параллельном создании схемы стейджа

**Severity:** High — на **первом** прогоне против свежей БД (или при первом появлении нового
стейджа/схемы) один из параллельных ассетов падает с `duplicate key ... pg_namespace`
(SQLSTATE 23505), теряется он и все его зависимые ассеты. Данные стейджа получаются
**частичными**. Саморазрешается со следующего прогона (схема уже существует), но первый
прогон недостоверен.

**Discovered:** 2026-07-19 при раскатке rewards-v2 пайплайна `partner-analytics` (Elstate
Partner Platform) на dev — первый прогон новых стейджей `dds_rewards` / `analytical_updates`.

**Reported by:** Claude Code (Opus 4.8) при выводе rewards-v2 аналитики на dev DO.

**Affected versions:** v1.2.12 (и, судя по коду, все версии с `CheckSchemaExists`+`CREATE SCHEMA`
в `sql_asset.go` — механизм не менялся). Родственно [#001](./001-pgx-conn-panic-on-parallel-asset-execution.md)
(та же параллельность ассетов).

**Status:** ✅ RESOLVED — со второго захода.

1. v1.2.13 — минимальный фикс (вариант 4.1 ниже): `CREATE SCHEMA IF NOT EXISTS` в
   `sql_asset.go:92`. **Оказался недостаточен:** «узкое окно неатомарности», которое тогда
   приняли теоретическим, воспроизводится на практике — 23505 на `pg_namespace_nspname_index`
   продолжал ловиться на первом прогоне. `IF NOT EXISTS` в PostgreSQL проверяет каталог вне
   блокировки, поэтому две сессии проходят проверку и обе вставляют в `pg_namespace`.
2. Текущий фикс — вариант 4.3 («ловить дубликат») плюс явная сериализация DDL, вынесенные
   в драйвер (`DBDriver.CreateSchema`, снимает `TODO: Move this to the driver`):
   - `PostgresDBEngine.schemaMutex` — сериализует горутины одного процесса;
   - `pg_advisory_xact_lock(<fnv64(schema)>)` — сериализует разные процессы teal на одной БД;
   - DDL внутри savepoint, `42P06`/`23505` трактуются как success — проигранная гонка не
     аборчит внешнюю транзакцию ассета;
   - `DuckDBEngine` — отдельный `schemaMutex` (не `Mutex` из `ConcurrencyLock()`: он уже
     захвачен на всё время `Execute`, а мьютексы в Go не реентерабельны — см. `f7acafe`).

---

## 1. Symptom

Стейдж `dds_rewards` содержит несколько независимых ассетов (`dim_scd_user_level`,
`dim_current_user_level`, …). На **первом** прогоне, когда схемы `dds_rewards` ещё нет, два
ассета стартуют почти одновременно, оба проверяют существование схемы, оба видят «нет», оба
идут создавать. Победитель коммитит `CREATE SCHEMA`, проигравший падает:

```
17:49:46.207099  dim_scd_user_level     "Schema dds_rewards does not exist"
17:49:46.207301  dim_current_user_level "Schema dds_rewards does not exist"   ← 0.2 мс спустя, до коммита победителя
17:49:46.237513  dim_scd_user_level     "Asset complete"                      ← создал схему, закоммитил
17:49:46.260134  ERROR sql="CREATE SCHEMA dds_rewards;"
                 error="ERROR: duplicate key value violates unique constraint
                        \"pg_namespace_nspname_index\" (SQLSTATE 23505)"       ← postgres.go:125
17:49:46.260168  dim_current_user_level "Failed to create schema"             ← sql_asset.go:95
17:49:46.262114  dim_current_user_level "Asset Error"                         ← channel_dag.go:237
17:49:46.262129  update_partner_levels  "Task has been ignored"              ← зависит от dim_current, пропущен
```

Итог прогона: `dim_current_user_level` + весь его downstream (`update_partner_levels`,
пишущий `public.partners.level_id`) — не выполнены. Остальные ассеты прошли. То же затем
повторяется на первом создании схемы `analytical_updates`.

---

## 2. Root cause

Неатомарный **check-then-act** в `pkg/processing/sql_asset.go` (v1.2.12):

```go
// sql_asset.go:82
isSchemaExists := dbConnection.CheckSchemaExists(tx, s.descriptor.Name)

if !isSchemaExists {                                    // :84
    splitted := strings.Split(s.descriptor.Name, ".")
    // TODO: Move this to the driver                     // :91 (комментарий уже есть)
    err = dbConnection.Exec(tx, fmt.Sprintf("CREATE SCHEMA %s;", splitted[0]))  // :92 — БЕЗ IF NOT EXISTS
    ...
}
```

`CheckSchemaExists` (`pkg/drivers/postgres.go:77`) читает `information_schema.schemata`. Каждый
ассет исполняется в **своей** транзакции/goroutine (`ChannelDag` спавнит goroutine per asset,
см. #001). Параллельность ограничена размером **pgxpool** (у нас `ANALYTICS_DB_MAX_CONNS`,
проброшенный в `PoolMaxConns`), а **не** `cores` (`cores` в channel_dag влияет только на
размер буферов каналов). При pool ≥ 2 два ассета одного стейджа:

1. оба `CheckSchemaExists` → `false` (ни один ещё не закоммитил `CREATE SCHEMA`);
2. оба `CREATE SCHEMA <stage>;`;
3. первый коммит выигрывает, второй → 23505 на `pg_namespace_nspname_index`.

Схема относится ко всему стейджу, у неё общее имя → это единственный DDL с «пересекающимся»
именем между параллельными ассетами (таблицы/вью у каждого ассета уникальны, поэтому там
`CREATE TABLE IF NOT EXISTS` гонки не даёт).

---

## 3. Почему саморазрешается

После первого прогона победитель оставляет схему закоммиченной. На следующем прогоне
`CheckSchemaExists` → `true` у всех ассетов → `CREATE SCHEMA` не вызывается → гонки нет. То
есть баг проявляется ровно один раз на каждый новый стейдж (на свежей БД — на первом прогоне
по одному разу на каждую схему). Именно поэтому dev «починился» повторным прогоном, а на
prod-катовере первый прогон снова частичный.

---

## 4. Suggested fix

**Минимум (идемпотентный DDL):** `sql_asset.go:92` →

```go
err = dbConnection.Exec(tx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", splitted[0]))
```

Убирает 23505 в подавляющем большинстве случаев. Нюанс: `CREATE SCHEMA IF NOT EXISTS` в
PostgreSQL под высокой конкуренцией всё ещё имеет узкое окно (проверка каталога неатомарна),
поэтому строго-надёжный вариант — один из:

**Надёжно (создание схем один раз до DAG):** перед спавном goroutine'ов ассетов пройтись по
уникальным стейджам проекта и `CREATE SCHEMA IF NOT EXISTS <stage>` последовательно в одной
транзакции (owner — DAG init, а не отдельный ассет). Тогда ассеты вообще не занимаются DDL
схем. Это чище и снимает `TODO: Move this to the driver` из sql_asset.go:91.

**Либо** ловить `23505`/`duplicate_schema` в ветке создания и трактовать как success (schema
уже создана конкурентом).

---

## 5. Workaround (без правки teal)

- Пред-создать схемы стейджей вручную/миграцией до первого прогона (`CREATE SCHEMA IF NOT
  EXISTS <stage>`), тогда `CheckSchemaExists` = true с самого начала.
- Либо сериализовать доступ: pool = 1 (`ANALYTICS_DB_MAX_CONNS=1`) — один ассет за раз,
  первый создаёт схему, остальные видят её существующей. Минус — теряется параллелизм DAG.

На partner-analytics обошли пред-созданием схем на dev вручную; для prod ждём фикса в teal.
