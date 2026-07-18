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

## Resolved

_(empty)_
