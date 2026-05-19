# Issue #002 — scaffold генерирует невалидный `require ... dev` в go.mod

**Severity:** High — out-of-box `make build` падает на свежесгенерённом проекте.

**Discovered:** 2026-05-19 при сборке `partner-analytics`.

---

## Symptom

После `teal init` + `teal gen` сгенерированный `go.mod` содержит:

```go
module github.com/your_user/your_project

go 1.25.1

require github.com/go-teal/teal dev

// replace github.com/go-teal/teal => ./..
```

`go build` или `go mod tidy` падает:

```
go: errors parsing go.mod:
go.mod:5: require github.com/go-teal/teal: version "dev" invalid: unknown revision dev
```

`dev` — не валидная semver-версия и не существующая git-ветка/тег в публичном
`github.com/go-teal/teal` репозитории.

---

## Root cause

В шаблоне scaffold (`scaffold/...` или генератор в `cmd/teal/gen`) hardcoded строка
`require github.com/go-teal/teal dev`. Замысел, видимо, был — указать на локальный
working copy через закомментированный `replace`-блок ниже. Но out-of-box у пользователя
работающего teal source нет, и `dev` — это просто placeholder.

---

## Steps to reproduce

```bash
mkdir test && cd test && teal init && teal gen
cat go.mod | grep "require github.com/go-teal/teal"
# Видим: require github.com/go-teal/teal dev

go mod tidy
# fails with "version \"dev\" invalid"
```

---

## Fixes

### Option A (minimal): pin актуальной релиза в шаблоне

```go
require github.com/go-teal/teal v1.0.5  // ← текущая релизная версия
```

Минус: при каждом релизе teal'а надо бампить шаблон. Можно через `${VERSION}` substitution
в генераторе подставлять значение из `pkg/version.Version` (если есть).

### Option B: убрать require из шаблона целиком

`go mod tidy` подцепит teal как dep из импортов в сгенерированных `internal/assets/*.go`.
Самый чистый вариант.

### Option C: оставить commented placeholder

```go
// require github.com/go-teal/teal v1.0.5  // pin актуальной версии
```

— но тогда `tidy` всё равно добавит, и `dev`-bug исчезнет.

**Рекомендуется:** **Option B** — убрать строку, дать `tidy` подцепить.

---

## Workaround (текущий)

В `partner-analytics` используем `scripts/regen.sh` который после `teal gen` делает:

```bash
sed -i 's|^require github.com/go-teal/teal dev$|require github.com/go-teal/teal v1.0.5|' go.mod
go mod tidy
```

См. `scripts/regen.sh` в `dubai-one-click/partner-analytics`.
