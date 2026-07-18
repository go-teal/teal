# Issue #007 — `teal gen` затирает пользовательский Makefile при каждой регенерации

**Severity:** High — ручные правки Makefile теряются на любом `teal gen`
(а `make run` вызывает `teal gen` внутри), молча ломая кастомные таргеты.

**Discovered:** 2026-07-18 в `partner-analytics`: добавленные таргеты
`ui` / `run` / `check-minikube-db` (env-обвязка под minikube postgresql,
влиты в develop) исчезли после `teal gen` — `make ui` перестал существовать.

**Status:** ✅ RESOLVED — fix в `fix/gen-makefile-skip-existing` (PR #24), релиз v1.2.7.

## Root cause

`internal/domain/generators/gen_makefile.go::RenderToFile` безусловно
делал `os.Create` и перезаписывал `Makefile` шаблоном на каждом gen.

Все остальные скаффолд-генераторы, которые проект кастомизирует —
`go.mod` (`gen_go_mod.go`), `main.go` (`gen_main.go`), `main-ui.go`,
`Dockerfile` (`gen_dockerfile.go`) — реализуют skip-if-exists: `os.Stat`
на целевой путь, при существовании возвращают `(nil, true)` → `[SKIPPED]`.
`Makefile` был единственным исключением, хотя это ровно такой же
одноразовый скаффолд, который дальше правит разработчик.

## Fix

Добавлен тот же guard в `GenMakefile.RenderToFile`:

```go
if _, err := os.Stat(g.GetFullPath()); err == nil {
    return nil, true // существующий Makefile не трогаем
}
```

Регенерация происходит только если файла нет (как и просил кейс). Чтобы
получить свежий шаблон, Makefile нужно удалить перед gen — идентично
поведению go.mod / main.go / Dockerfile.

Тесты: `internal/domain/generators/gen_makefile_test.go`
(skip при наличии файла + генерация при отсутствии).

## Impact на partner-analytics

Восстановлен Makefile из git (таргеты `ui`/`run`/`check-minikube-db`).
После апдейта teal CLI регенерация его больше не тронет.
