# Backend Agent Guidelines (Anahix Server)

This document contains specialized rules for agents and developers modifying the Golang backend (`server/`).

---

## 1. Core Architecture & Layer Conventions
- **Clean Architecture Pattern**:
  - `server/cmd/`: CLI commands with Cobra. Auto-migrations and seed data run in `run.go`.
  - `server/internal/handler/`: Gin HTTP handlers. ONLY decode requests, call logic layer, and wrap responses using `result.HttpResult(c, data, err)` or `result.ParamErrorResult(c, err)`. NEVER write database queries or business logic in handlers.
  - `server/internal/logic/`: Core business logic and database queries (`gorm.DB`). Organized into `public/`, `admin/`, `fivesim/`, and `vault/`.
  - `server/internal/model/`: GORM models (`Product`, `InventoryItem`, `Order`, `Payment`, `SystemSetting`, `AdminUser`, `VirtualNumberOrder`).
  - `server/internal/svc/`: `ServiceContext` containing `DB`, `Config`, and `Redis`.
  - `server/internal/types/`: DTOs, request payloads, and API response structs.
  - `server/apis/`: Goctl `.api` DSL specifications (ArchNet standard).

---

## 2. Database-First Configuration Rules
- **Payment Gateways**: Do NOT put payment gateways or credentials in `config.yaml`. Payment options are dynamically queried from the `payment` table. The customer balance gateway (`id: -1`, `platform: "balance"`) is the default.
- **System Settings**: SMS gateways and 5SIM configuration are stored in the `system` table (`category: "sms"`, `category: "fivesim"`), NOT in `config.yaml`.
- **Zero-Config Development (SQLite Fallback)**:
  - If PostgreSQL is unreachable or not running, the server automatically falls back to local SQLite (`server/anahix.db`).
  - Models and initial seed data are auto-migrated on startup (`server/cmd/run.go`).

---

## 3. 5SIM Virtual Number Integration Rules
- **No Hardcoding**: Services, countries, and operator pricing must be fetched from live 5SIM endpoints (`/guest/products/any/any` and `/guest/prices?product={service}`).
- **Currency & Margin (Dual-Priority)**: 5SIM base costs are in USD; customer Toman price is calculated using the effective rate determined by `fivesim_rate_priority` (`manual` vs `swapwallet`) and `fivesim_margin_percent` from the `system` table and rounded to 500 Toman.
  - Priority 1: Manual rate in `fivesim_usd_rate`.
  - Priority 2: Live USDT/IRT from SwapWallet (`server/pkg/exchangerate/swapwallet.go`) with auto-fallback to manual rate on upstream failure.

---

## 4. Verification & Build Commands
```bash
# Build
go build ./...

# Run unit & integration tests
go test -v ./...

# Run local development server (port 8080)
go run main.go run

# Kill background server process on port 8080
make kill

# Restart server (frees port 8080 first, then runs)
make restart

# Generate Swagger docs from .api DSL
make swagger-gen
```
