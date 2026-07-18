# Issue #004 — `teal ui` спавнит API-процесс без `-tags teal_ui` (API-половина никогда не собирается)

**Severity:** High — команда `teal ui` из коробки поднимает только статический дашборд;
API-процесс проекта падает на компиляции, весь функционал UI (DAG, данные ассетов,
тесты, логи) мёртв.

**Discovered:** 2026-07-17 при отладке `partner-analytics` (AN-01, minikube-БД).

**Status:** ✅ пофикшено локально (ветка `fix/ui-spawn-build-tag`, коммит `050e137`);
апстрим — TODO (см. Next steps).

---

## Symptom

`teal ui --port 8080` печатает:

```
Starting UI Dashboard on port 8081 (static assets server)
✨ UI available at: http://localhost:8081
Starting UI process: cmd/<project>-ui/<project>-ui.go (port=8080, log_level=debug)
UI process started (PID=...)

package command-line-arguments
	imports github.com/go-teal/teal/pkg/ui: build constraints exclude all Go files in .../teal@v1.1.1/pkg/ui
```

Дашборд (порт `--port + 1`) отдаёт HTML, но все запросы к API (`--port`) падают —
API-процесс не скомпилировался. Так себя ведут и v1.1.1, и **v1.2.0** (проверено по
`asset_observer.go` из module cache).

## Root cause

`pkg/ui` и генерируемый scaffold'ом `cmd/<project>-ui/<project>-ui.go` спрятаны за
build-тегом `teal_ui` (чтобы gin не попадал в прод-бинарь). Но CLI в
`internal/domain/services/asset_observer.go` запускает процесс так:

```go
cmd := exec.Command("go", "run", uiPath,
    "--port", fmt.Sprintf("%d", ao.port),
    "--log-level", ao.logLevel)
```

— без `-tags teal_ui`. CLI не может собрать код, который сам же сгенерировал.

## Fix

Одна строка (коммит `050e137` в этом клоне):

```go
cmd := exec.Command("go", "run", "-tags", "teal_ui", uiPath, ...)
```

Локальный CLI пересобран: `go install ./cmd/teal` (перекрывает апстримный в
`$GOBIN`). Проверено end-to-end на `partner-analytics`: `make ui` поднимает и
дашборд (8081), и API (8080), ошибок сборки нет.

## Next steps

- [ ] PR в апстрим `go-teal/teal` (одна строка + при желании тест на команду спавна).
- [ ] После релиза апстрима — вернуть CLI на released-версию, локальный патч снять.
- [ ] Примечание в `partner-analytics/Makefile` (`make ui`) обновить, когда патч
      перестанет быть нужен.

## Workaround (без патча CLI)

Запускать половинки руками: API — `go run -tags teal_ui ./cmd/<project>-ui --port 8080`
(в `partner-analytics` это `make run`), дашборд — `teal ui --port 8080` рядом (его
собственный спавн API падает безвредно, порт держит ручной процесс).
