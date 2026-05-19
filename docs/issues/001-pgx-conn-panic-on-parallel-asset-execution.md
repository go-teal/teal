# Issue #001 — PostgreSQL driver: `pgx.Conn` panic при параллельном выполнении ассетов

**Severity:** Critical — без обхода teal-пайплайн с PostgreSQL и **2+ ассетами** падает с panic'ом независимо от `cores` config.

**Discovered:** 2026-05-19 при сборке `partner-analytics` (Elstate Partner Platform) — 24 модели в DAG, 8 staging + 5 dim + 5 fact + 14 mart.

**Reported by:** Claude Code (PoC по dimensional modeling для admin dashboard TASK-AD-027).

**Affected versions:** v1.0.5 (вероятно все версии с того момента как добавили `ChannelDag`).

---

## 1. Symptom

При запуске сгенерированного бинаря против PostgreSQL подключения:

```
panic: BUG: slow write timer already active

goroutine 11 [running]:
github.com/jackc/pgx/v5/pgconn.(*PgConn).enterPotentialWriteReadDeadlock(...)
	pgconn.go:2048
github.com/jackc/pgx/v5/pgconn.(*PgConn).flushWithPotentialWriteReadDeadlock(0xc0002b6008)
	pgconn.go:2067 +0xa6
github.com/jackc/pgx/v5/pgconn.(*PgConn).Close(0xc0002b6008, ...)
	pgconn.go:715 +0x185
github.com/jackc/pgx/v5.(*Conn).die(0xc0001bec80)
	conn.go:445 +0x65
github.com/jackc/pgx/v5.(*Conn).BeginTx(0xc0001bec80, ...)
	tx.go:105 +0x110
github.com/jackc/pgx/v5.(*Conn).Begin(...)
	tx.go:95
github.com/go-teal/teal/pkg/drivers.(*PostgresDBEngine).Begin(...)
	pkg/drivers/postgres.go:85 +0x46
github.com/go-teal/teal/pkg/processing.(*SQLModelAsset).Execute(...)
	pkg/processing/sql_asset.go:70 +0x25b
github.com/go-teal/teal/pkg/dags.(*DagRoutine).run(...)
	pkg/dags/channel_dag.go:234 +0x438
created by github.com/go-teal/teal/pkg/dags.(*ChannelDag).Run
	pkg/dags/channel_dag.go:103 +0x6b
```

Voiced by `pgx`: «BUG: slow write timer already active». Это **намеренный panic из pgx** —
сигнал что `*pgx.Conn` используется из нескольких goroutine'ов одновременно. `pgx.Conn` контрактно
**не thread-safe**, для concurrency нужен `pgxpool.Pool`.

---

## 2. Root cause

Два независимых факта вместе создают race:

### 2.1 `ChannelDag` спавнит goroutine **per asset**

`pkg/dags/channel_dag.go:97-105`:

```go
func (dag *ChannelDag) Run() *sync.WaitGroup {
    var wg sync.WaitGroup
    for _, taskGroup := range dag.dagGrpah {
        for _, task := range taskGroup {
            wg.Add(1)
            go dag.dagRoutineMap[task].run(&wg)   // <- goroutine per asset
        }
    }
    return &wg
}
```

При DAG'е из, скажем, 24 ассетов — 24 одновременно живых goroutine'ы. `cores` config
**не ограничивает количество goroutine'ов** (проверено эмпирически: `cores: 1` → тот же panic).
Видимо, `cores` это semafор где-то внутри `DagRoutine.run`, но `Begin()` происходит до семафора.

### 2.2 PostgreSQL driver хранит **один** `*pgx.Conn`

`pkg/drivers/postgres.go:14-17`:

```go
type PostgresDBEngine struct {
    dbConnection *configs.DBConnectionConfig
    db           *pgx.Conn   // <- единственный экземпляр на весь pipeline
}
```

Все asset-goroutine'ы зовут `dbConnection.Begin()` (`postgres.go:84-86`), который делает
`d.db.Begin(ctx)` на shared `*pgx.Conn`. Это и есть race.

### 2.3 `ConcurrencyLock/Unlock` для Postgres **— NO-OP**

`pkg/drivers/postgres.go:159-165`:

```go
func (d *PostgresDBEngine) ConcurrencyLock() {

}

func (d *PostgresDBEngine) ConcurrencyUnlock() {

}
```

Хотя сами hooks правильно вызываются из `pkg/processing/sql_asset.go:55-56`:

```go
dbConnection.ConcurrencyLock()
defer dbConnection.ConcurrencyUnlock()
```

— реализация для Postgres пустая.

---

## 3. Steps to reproduce

```bash
# 1. Создать teal-проект с подключением к PostgreSQL
mkdir test-pg-panic && cd test-pg-panic && teal init

# 2. config.yaml — заменить duckdb на postgres
cat > config.yaml <<EOF
version: '1.0.0'
module: example.com/test
connections:
  - name: default
    type: postgres
    config:
      host: <your-pg-host>
      port: 5432
      user: <user>
      password: <password>
      database: <db>
      db_sslnmode: require
EOF

# 3. Минимум 2 независимые модели (одновременно становятся runnable):
mkdir -p assets/models/raw
cat > assets/models/raw/stg_a.sql <<EOF
{{ define "profile.yaml" }}
    materialization: 'table'
{{ end }}
SELECT 1 as x
EOF

cat > assets/models/raw/stg_b.sql <<EOF
{{ define "profile.yaml" }}
    materialization: 'table'
{{ end }}
SELECT 2 as x
EOF

# 4. Build and run
teal gen
# Patch unused pgx import (issue #003), pin version (issue #002), then:
make build
./bin/test-pg-panic --log-output=raw --log-level=info --with-tests=false

# Expected: panic: BUG: slow write timer already active
```

Воспроизводится надёжно на 2+ ассетах. На 1 ассете — race не происходит, всё работает.

---

## 4. Workarounds (от плохого к лучшему)

### 4.1 «Запустить с одним ассетом» — неприемлемо.
Real-world pipeline это десятки моделей.

### 4.2 **Mutex в `ConcurrencyLock/Unlock`** — рабочий минимальный патч

Файл `pkg/drivers/postgres.go`:

```go
type PostgresDBEngine struct {
    dbConnection  *configs.DBConnectionConfig
    db            *pgx.Conn
    concurrencyMu sync.Mutex   // ← добавить
}

// …

func (d *PostgresDBEngine) ConcurrencyLock() {
    d.concurrencyMu.Lock()
}

func (d *PostgresDBEngine) ConcurrencyUnlock() {
    d.concurrencyMu.Unlock()
}
```

**Эффект:** asset Execute() сериализуется. Pipeline проходит корректно.
**Цена:** теряется параллелизм per-asset (`cores: 4` де-факто становится `cores: 1`).
На наших объёмах (DAG ≤30 ассетов, full run <3 sec) — не заметно.

**Это решение я применил локально** в `partner-analytics` через `replace github.com/go-teal/teal => ../teal` в go.mod.

### 4.3 **`pgxpool.Pool` вместо `pgx.Conn`** — правильный fix

Заменить `db *pgx.Conn` на `db *pgxpool.Pool`. Каждый `Begin()` тогда берёт connection
из пула — true concurrent execution. Изменения:

- `Connect()`: `pgxpool.New(ctx, dsn)` вместо `pgx.Connect(ctx, dsn)`.
- `Begin()`: `tx, err := d.db.Begin(ctx)` — Pool API такой же как у Conn.
- `Close()`: `d.db.Close()` (без ctx).
- ConcurrencyLock/Unlock оставить no-op.

Pool size конфигурируется через DSN параметр `pool_max_conns=N`. Дефолт pgxpool — 4.
Совместимо с DO Managed PG Basic (max_connections=25) если pipeline один, и
с теми же ограничениями для concurrent pipeline'ов.

**Рекомендуется** именно этот fix — это правильный pgx-pattern для concurrent работы.

### 4.4 Использовать `MAT_CUSTOM` для всего — обход на стороне моделей

Если все модели сделать `materialization: 'custom'` — `customQuery()` тоже зовёт
`Begin()`, так что race остаётся. **Не помогает.**

---

## 5. Связанные issue'ы

- **#002:** scaffold генерирует невалидный `require github.com/go-teal/teal dev` в go.mod
- **#003:** scaffold генерирует unused `"github.com/jackc/pgx/v5"` импорт в `cmd/<proj>/main.go`

Все три проблемы вместе блокируют out-of-box работу teal-проекта с PostgreSQL. После
их фикса проект собирается и работает корректно.
