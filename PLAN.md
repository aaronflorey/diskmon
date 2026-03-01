AGENTS.md

Project: diskmon

Purpose:
diskmon is a lightweight Go daemon and CLI tool that monitors hard drives using S.M.A.R.T, stores telemetry in DuckDB, and exposes a Web UI (Vue) and API for monitoring drive health, metrics, and history.

The system must be reliable, extensible, and readable. Future expansion will include additional telemetry sources and alerting.

⸻

PRODUCT_REQUIREMENTS.md

Goals

diskmon must:
	1.	Monitor all configured drives using S.M.A.R.T
	2.	Ingest and store all metrics in DuckDB
	3.	Provide drive health classification (green / yellow / red)
	4.	Provide a Web UI showing:
	•	Overview dashboard (cards)
	•	Detailed drive view
	•	Historical metrics
	5.	Run as:
	•	CLI tool
	•	daemon/service
	6.	Be configurable via:
	•	env vars
	•	yaml config file
	•	CLI args
	7.	Be designed for future extensibility

⸻

Non-Goals (for initial version)
	•	No authentication
	•	No distributed monitoring
	•	No alerting integrations
	•	No clustering

These will be added later.

⸻

ARCHITECTURE.md

High-level architecture

diskmon
 ├── CLI
 ├── Daemon
 │    ├── SMART collector
 │    ├── DuckDB storage
 │    ├── Health evaluator
 │    ├── API server
 │    └── Web UI
 └── Shared libraries


⸻

Technology stack

Backend

Language: Go

Required libraries:
	•	SMART collection: smartctl via exec (NOT CGO bindings)
	•	Database: DuckDB (github.com/marcboeker/go-duckdb)
	•	CLI: cobra
	•	config: viper
	•	web framework: chi (preferred) or echo/fiber
	•	logging: slog

Reasoning: minimal dependencies, high readability, production proven

⸻

Frontend

Framework: Vue 3
Build tool: Vite
Styling: TailwindCSS
Charts: lightweight (chart.js or apexcharts)

Frontend must be embedded into Go binary.

⸻

MODULE_STRUCTURE.md

Required Go module structure:

/cmd/diskmon/
    root.go
    daemon.go
    scan.go
    version.go

/internal/

/config/
    config.go

/smart/
    collector.go
    parser.go
    models.go

/storage/
    duckdb.go
    schema.sql
    ingest.go
    queries.go

/health/
    evaluator.go
    rules.go

/api/
    server.go
    routes.go
    handlers.go
    models.go

/web/
    embed.go

/util/
    logging.go
    time.go

Frontend:

/webui/
    /src/
        /components/
        /views/
        /stores/
        /api/


⸻

CLI_REQUIREMENTS.md

Tool name:

diskmon

Commands:

diskmon daemon
diskmon scan
diskmon version
diskmon config validate


⸻

CLI flags

Must support:

--config path.yaml
--db path.duckdb
--interval duration
--web-listen address
--drives /dev/sda,/dev/nvme0n1
--log-level


⸻

Config sources priority

Highest → lowest:
	1.	CLI args
	2.	ENV
	3.	YAML
	4.	Defaults

⸻

Example YAML:

database: diskmon.duckdb

collector:
  interval: 60s
  drives:
    - /dev/sda
    - /dev/nvme0n1

web:
  listen: 0.0.0.0:8976


⸻

STORAGE_SPEC.md

Database: DuckDB

File-based.

⸻

Required tables

drives

id
device
model
serial
wwn
first_seen_at
last_seen_at


⸻

smart_samples

id
drive_id
collected_at
temperature
power_on_hours
reallocated_sectors
pending_sectors
uncorrectable_sectors
wear_level
raw_json


⸻

smart_attributes

sample_id
attribute_id
name
value
worst
threshold
raw


⸻

drive_health

drive_id
sample_id
status
score


⸻

⸻

SMART_COLLECTION_SPEC.md

Collector must:
	•	use smartctl
	•	run:

smartctl -a -j /dev/sda

Must parse JSON output.

Must support:
	•	SATA
	•	NVMe

⸻

Collector output model:

DriveInfo
SmartSample
SmartAttribute


⸻

Collection loop:

every interval:
    collect all drives
    store sample
    evaluate health


⸻

HEALTH_EVALUATION_SPEC.md

Health states:

GREEN
YELLOW
RED
UNKNOWN


⸻

Initial rules:

RED if:
	•	failing_now == true
	•	reallocated_sectors > threshold
	•	uncorrectable > 0
	•	critical_warning present (NVMe)

YELLOW if:
	•	temperature > warning threshold
	•	pending sectors > 0
	•	wear level degraded

GREEN otherwise

⸻

Health evaluation must be rule-based and extensible.

⸻

API_SPEC.md

Base path:

/api/v1/

Endpoints:

⸻

GET /drives

Returns:

[
  {
    id
    device
    model
    serial
    health
    temperature
    last_seen
  }
]


⸻

GET /drives/:id

Returns:

drive info
latest stats
health


⸻

GET /drives/:id/history

Returns:

time series


⸻

GET /drives/:id/attributes

Returns:

all smart attributes
with health classification


⸻

WEBUI_SPEC.md

Dashboard must show:

Drive cards.

Each card shows:
	•	Device name
	•	Model
	•	Temperature
	•	Health status color
	•	Power on hours
	•	Serial

Card color:

Green
Yellow
Red


⸻

Clicking card opens detail page:

Must show:

Sections:

Summary
Health
SMART attributes
History charts

Each attribute must include:

name
value
raw value
threshold
status colour
explanation


⸻

FRONTEND_STRUCTURE.md

Vue structure:

/views/
    Dashboard.vue
    DriveDetail.vue

/components/
    DriveCard.vue
    AttributeTable.vue
    HealthBadge.vue
    TemperatureBadge.vue
    HistoryChart.vue


⸻

DAEMON_SPEC.md

Daemon must:

Start:

collector
api server
webui server

Must shutdown cleanly.

Must handle:

SIGINT
SIGTERM


⸻

WEB_SERVER_SPEC.md

Web server must serve:

/api/*
/ -> webui

Frontend must be embedded using go:embed.

⸻

CONFIG_SPEC.md

Config system must use:

viper

Sources:

yaml
env
flags

Env naming:

DISKMON_DATABASE
DISKMON_WEB_LISTEN
DISKMON_INTERVAL


⸻

LOGGING_SPEC.md

Use slog.

Levels:

DEBUG
INFO
WARN
ERROR


⸻

CODING_REQUIREMENTS.md

The agent MUST:

Follow:
	•	small files
	•	clean module boundaries
	•	readable code over clever code
	•	no global state
	•	dependency injection where appropriate

Avoid:
	•	monolithic files
	•	overly abstract code

⸻

FUTURE_EXPANSION_POINTS.md

Must be designed to support future:
	•	alerting
	•	distributed agents
	•	remote collection
	•	filesystem monitoring
	•	RAID monitoring
	•	Prometheus exporter
	•	notifications
	•	authentication

⸻

BUILD_REQUIREMENTS.md

Must support:

go build
go run

Frontend:

npm install
npm run build

Build pipeline must embed frontend automatically.

⸻

ACCEPTANCE_CRITERIA.md

Agent implementation is complete when:
	•	daemon runs
	•	drives detected
	•	SMART data stored
	•	health evaluated
	•	API functional
	•	Web UI shows drives
	•	details page works
	•	history stored and displayed

⸻

PRIORITY_IMPLEMENTATION_ORDER.md

Agent must implement in this order:
	1.	config
	2.	duckdb storage
	3.	smart collector
	4.	health evaluator
	5.	daemon loop
	6.	api server
	7.	web ui embed
	8.	frontend dashboard
	9.	frontend detail page