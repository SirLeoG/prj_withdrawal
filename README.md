# prj_withdrawal — Cajero en Go (CLI + App Web)

Proyecto del curso (DSL, Unidad III): desarrollo **iterativo e incremental** con **BDD** (Gherkin) y **TDD**.

| Iteración | Contenido |
|---|---|
| 1 | CLI interactivo con **Retiro** (punto de partida) |
| 2 | US-01b **Abono** en efectivo |
| 3 | US-04 **integridad de transacción** + log en **JSON** (éxito / fallo / error) |
| 4 | **Refactorización** a estructura `cmd/` + `internal/` y **App Web** (página + API JSON) |

## Ejecutar (requiere Go 1.24+)
```bash
go run ./cmd/cajero                      # CLI interactivo
go run ./cmd/cajero -modo web            # App Web en http://localhost:8080
go run ./cmd/cajero -modo web -titular Hamza -addr localhost:9000
go test -race ./...                      # pruebas
```
Cuentas iniciales: `internal/datos/cuentas.json` (incrustado con `go:embed`).
Log: `transacciones.json`, una línea JSON por transacción.

## API JSON
| Método y ruta | Descripción | Códigos |
|---|---|---|
| `GET /api/saldo` | Titular y saldo | 200 |
| `GET /api/transacciones` | Log completo | 200 |
| `POST /api/retiro` `{"monto":"200.00"}` | Retiro | 200 éxito · 422 fallo (regla de negocio) · 400 error (entrada ilegible) · 500 log no escrito |
| `POST /api/abono` `{"monto":"50.00"}` | Abono | igual que retiro |

```bash
curl -X POST localhost:8080/api/abono -d '{"monto":"100"}'
```

## Estructura
```
cmd/cajero/          punto de entrada (flags: -modo cli|web)
internal/banco/      dominio: Cuenta, Cajero (transacciones, log, rollback)
internal/cli/        interfaz de consola
internal/web/        interfaz web (net/http + html/template)
internal/datos/      datos iniciales JSON incrustados
features/            especificaciones Gherkin (.feature)
docs/                capturas e informe
```

## Reglas de negocio
- RN-01: no se retira más que el saldo. RN-02: no hay saldos negativos.
- RN-03: atomicidad; si la transacción no se puede registrar, la operación se revierte.
- Los montos se manejan en **centavos** (`int64`) para evitar errores de punto flotante.

Licencia MIT.
