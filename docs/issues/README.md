# Issues

Внутренний bug tracker — баги найденные при использовании teal'а в реальных
проектах. После фикса в коде — отметить здесь резолюшен (commit hash + версия).

## Open

| # | Severity | Title | Discovered | Project |
|---|---|---|---|---|
| 001 | Critical | [`pgx.Conn` panic при параллельном выполнении ассетов](./001-pgx-conn-panic-on-parallel-asset-execution.md) | 2026-05-19 | partner-analytics |
| 002 | High | [scaffold генерирует невалидный `require ... dev` в go.mod](./002-scaffold-go-mod-invalid-version.md) | 2026-05-19 | partner-analytics |
| 003 | High | [scaffold генерирует unused `pgx/v5` import](./003-scaffold-main-go-unused-pgx-import.md) | 2026-05-19 | partner-analytics |
| 004 | High | [`teal ui` спавнит API-процесс без `-tags teal_ui`](./004-ui-spawn-missing-build-tag.md) | 2026-07-17 | partner-analytics |
| 005 | High | [int4/int8 колонки теряются в DataFrame — везде null](./005-pg-dataframe-int-slices-become-null.md) | 2026-07-18 | partner-analytics |
| 006 | High | [«Select» ассета исполняет SQL модели — 42601 на SCD2, риск мутаций](./006-asset-select-executes-model-script.md) | 2026-07-18 | partner-analytics |
| 007 | High | [`teal gen` затирает пользовательский Makefile](./007-gen-makefile-clobbers-user-edits.md) | 2026-07-18 | partner-analytics |
| 008 | High | [`CREATE SCHEMA` без `IF NOT EXISTS` — гонка при первом параллельном создании схемы стейджа](./008-create-schema-race-on-parallel-first-creation.md) | 2026-07-19 | partner-analytics |

## Resolved

_(empty)_
