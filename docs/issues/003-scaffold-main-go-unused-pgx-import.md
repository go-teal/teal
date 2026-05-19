# Issue #003 — scaffold генерирует unused `"github.com/jackc/pgx/v5"` import

**Severity:** High — `make build` падает на свежесгенерённом проекте.

**Discovered:** 2026-05-19 при сборке `partner-analytics`.

---

## Symptom

Сгенерированный `cmd/<project>/<project>.go` начинается так:

```go
package main

import (


	"github.com/jackc/pgx/v5"

	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	modeltests "..."
	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/dags"
	"..."
)
```

`pgx/v5` импортирован, но **нигде в файле не используется** (даже если PostgreSQL
выбран как driver в `config.yaml`). Go отказывается компилировать:

```
cmd/<project>/<project>.go:6:2: "github.com/jackc/pgx/v5" imported as pgx and not used
```

Тот же баг в UI-debug main'е `cmd/<project>-ui/<project>-ui.go`.

---

## Root cause

В шаблоне генератора `cmd/main.go.tmpl` (или аналог) импорт `pgx` присутствует
**безусловно** — вероятно был добавлен «на будущее» (для дефолтных типов или
side-effect register). Но фактически использования в коде шаблона нет.

---

## Steps to reproduce

```bash
mkdir test && cd test && teal init && teal gen
# Опционально: fix issue #002 (pin go-teal version в go.mod)

go build ./cmd/test/test.go
# fails:
# cmd/test/test.go:6:2: "github.com/jackc/pgx/v5" imported as pgx and not used
```

---

## Fixes

### Option A: удалить импорт из шаблона

Если pgx не используется — просто удалить строку. Если в будущем понадобится
(например, для driver-registration) — добавить тогда же.

### Option B: blank-import

Изменить шаблон на:

```go
_ "github.com/jackc/pgx/v5"
```

— гарантирует side-effect registration без compile-error. Подходит если
`pgx` нужен для каких-нибудь init()-функций.

### Option C: условный импорт в шаблоне

Подставлять `pgx`-импорт только если в config.yaml есть postgres-connection.
Pongo2 поддерживает conditionals.

**Рекомендуется:** **Option A** — удалить целиком, если не нужно. Если есть
скрытая необходимость в side-effect — **Option B**.

---

## Workaround (текущий)

В `partner-analytics` используем `scripts/regen.sh`:

```bash
sed -i 's|^\t"github.com/jackc/pgx/v5"$|\t_ "github.com/jackc/pgx/v5"|' \
    cmd/<proj>/<proj>.go cmd/<proj>-ui/<proj>-ui.go
```
