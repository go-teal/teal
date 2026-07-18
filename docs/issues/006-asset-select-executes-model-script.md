# Issue #006 — «Select» ассета исполняет SQL модели: 42601 на мульти-стейтментных и риск мутаций

**Severity:** High — предпросмотр данных падает на SCD2/custom-моделях
(`ERROR: cannot insert multiple commands into a prepared statement, SQLSTATE 42601`),
а будь протокол попроще — кнопка предпросмотра ИСПОЛНЯЛА бы UPDATE/INSERT модели.

**Discovered:** 2026-07-18 при исследовании `partner-analytics` (Select на
`core.dim_partner` — SCD2-скрипт из 8 стейтментов).

**Status:** ✅ RESOLVED — fix в `fix/asset-select-from-relation`, релиз v1.2.4.

## Root cause

`ExecuteAssetSelect` рендерил `RawSQL` модели и запускал его через
`ToDataFrame` как есть. Для table/view-моделей это одиночный SELECT и работает;
для SCD2/custom-моделей — многокомандный скрипт: pgx (extended protocol) его
запрещает, а семантика «предпросмотр = выполнить мутации» неверна в принципе.

## Fix

Если отрендеренный SQL — не одиночный read-only SELECT/WITH (проверка
`isSingleReadOnlySelect`: комментарии пропускаются, внутренние `;` запрещены),
предпросмотр читает **материализованное отношение**: `SELECT * FROM <schema>.<asset>`.
Точка с запятой в строковом литерале даёт ложное срабатывание — деградация
безопасная (тоже уходим на relation-preview). Юнит-тесты приложены.
