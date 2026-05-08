# Twelve-Factor App Methodology

**Source**: [12factor.net](https://12factor.net/) by Adam Wiggins
**Purpose**: A methodology for building modern, scalable, maintainable software-as-a-service applications. Applicable to any language and any combination of backing services.

---

## When to Apply

Use this methodology when designing or reviewing any service that:
- Deploys to cloud infrastructure (Kubernetes, cloud VMs, PaaS)
- Needs to scale horizontally
- Requires continuous deployment
- Must minimize divergence between development and production
- Is part of a microservices or distributed system architecture

---

## I. Codebase

> One codebase tracked in revision control, many deploys

### What It Means

A twelve-factor app has exactly **one codebase** (a single repo or a set of repos sharing a root commit) per application. The relationship is always one-to-one:

- **Multiple codebases** = not an app, it's a distributed system. Each component in a distributed system is an app and can individually comply with twelve-factor.
- **Multiple apps sharing the same code** = violation. Factor shared code into libraries included through the dependency manager.

There are many **deploys** of the same codebase: production, staging, each developer's local environment. They may run different versions (commits), but they all share the same codebase.

### How to Apply

- One Git repository per service/application.
- Shared code lives in separate library repositories, pulled via dependency manager (`go get`, `npm install`, etc.).
- Every environment (dev, staging, prod) deploys from the same repository -- different tags/branches, same repo.

---

## II. Dependencies

> Explicitly declare and isolate dependencies

### What It Means

A twelve-factor app **never relies on implicit existence of system-wide packages**. It declares all dependencies completely and exactly via a dependency declaration manifest and uses a dependency isolation tool to ensure no implicit dependencies leak in from the surrounding system.

The full dependency specification is applied uniformly to both production and development.

Examples by language:
- **Go**: `go.mod` / `go.sum` (declaration), Go modules (isolation)
- **Python**: `pip` + `requirements.txt` (declaration), `virtualenv` (isolation)
- **Ruby**: `Gemfile` (declaration), `bundle exec` (isolation)
- **Node.js**: `package.json` + `package-lock.json` (declaration), `node_modules` (isolation)

### How to Apply

- All dependencies in `go.mod` (or equivalent). No implicit reliance on system tools.
- If the app requires a system tool (e.g., `ImageMagick`, `ffmpeg`), vendor it into the app or bundle it in the container image.
- A new developer should need only the language runtime and dependency manager to build and run the app.

---

## III. Config

> Store config in the environment

### What It Means

An app's **config** is everything likely to vary between deploys:
- Database credentials and connection strings
- Credentials for external services (S3, NATS, etc.)
- Per-deploy values (hostnames, feature flags, log levels)

Config does **not** include internal application structure (routes, DI wiring, module configuration) -- those rarely change between deploys and belong in code.

**Litmus test**: Could the codebase be open-sourced right now without exposing any credentials? If yes, config is properly separated.

The twelve-factor approach stores config in **environment variables**. Env vars are:
- Easy to change between deploys without touching code
- Language- and OS-agnostic
- Unlikely to be accidentally committed to the repo
- Granular and independently manageable (not grouped into "environments")

### How to Apply

- All secrets, connection strings, and per-deploy values come from environment variables.
- Use a config struct that reads from env vars at startup (e.g., `env` tags in Go, `os.Getenv`).
- Never hardcode credentials, URLs, or environment-specific values in source code.
- Avoid "environment groups" like `config/production.yaml` -- each variable is independently managed.
- Config files are acceptable for internal application wiring (route definitions, DI configuration) since they don't vary between deploys.

---

## IV. Backing Services

> Treat backing services as attached resources

### What It Means

A **backing service** is any service the app consumes over the network as part of its normal operation:
- **Datastores**: PostgreSQL, MySQL, Aerospike, Redis
- **Message/queue systems**: NATS, RabbitMQ, Kafka
- **SMTP services**: Postfix, Mailgun
- **Caching systems**: Memcached, Redis
- **Monitoring/logging**: Prometheus, Grafana, Elasticsearch
- **Third-party APIs**: S3, payment gateways, geolocation services

The code makes **no distinction between local and third-party services**. Both are attached resources, accessed via URL or credentials stored in config. Swapping a local PostgreSQL for Amazon RDS, or a local SMTP server for a cloud provider, should require **only a config change, zero code changes**.

Resources can be attached and detached at will. If a database is misbehaving, the administrator spins up a new one restored from backup, and attaches it -- all through config, no code deploy required.

### How to Apply

- All backing service connections are configured via environment variables (Factor III).
- Use interfaces/abstractions for storage, messaging, and external services so implementations are swappable.
- Never hardcode hostnames, ports, or credentials for any backing service.
- Each distinct backing service (e.g., two different databases) is a distinct attached resource with its own config entry.

---

## V. Build, Release, Run

> Strictly separate build and run stages

### What It Means

The deployment pipeline has exactly **three stages**:

| Stage | What Happens | Who Triggers |
|-------|-------------|-------------|
| **Build** | Converts code repo into an executable bundle. Fetches dependencies, compiles binaries and assets. Runs at a specific commit. | Developer (via CI/CD) |
| **Release** | Combines the build with deploy-specific **config**. The result is ready for immediate execution. | CI/CD pipeline |
| **Run** | Launches the app in the execution environment by starting processes against a selected release. | Automated (on deploy, restart, crash recovery) |

Strict separation means:
- You **cannot** make code changes at runtime -- changes must go through the build stage.
- Every release has a **unique ID** (timestamp like `2025-03-17T14:30:00` or incrementing number `v42`).
- Releases are **immutable** -- any change creates a new release.
- Rollback = deploying a previous release, not patching the current one.

### How to Apply

- CI/CD pipeline builds a Docker image (build stage), tags it with a release ID, and pushes it to a registry.
- Deployment combines the image with environment-specific config (release stage).
- The orchestrator (Kubernetes) runs the image (run stage).
- Never SSH into a container to edit code. Never modify a running release.
- Keep the run stage as simple as possible -- if it breaks at 3 AM, it should auto-recover without developer intervention.

---

## VI. Processes

> Execute the app as one or more stateless processes

### What It Means

Twelve-factor processes are **stateless and share-nothing**. Any data that needs to persist must be stored in a stateful backing service (database, cache, object store).

The memory space or filesystem of the process can be used as a brief, single-transaction cache (e.g., downloading a large file, processing it, storing the result in the database). But the app **never assumes** that anything cached in memory or on disk will be available on a future request -- with many processes of each type running, a future request will likely be served by a different process.

**Sticky sessions are a violation.** Session state data should be stored in a datastore with time-expiration, such as Redis or Memcached.

Asset compilation (e.g., bundling CSS/JS) happens during the **build stage**, not at runtime.

### How to Apply

- No in-process session state. Use Redis/Memcached for sessions.
- No local filesystem writes that other processes or future requests depend on. Use object storage (S3) or a database.
- Design every process to be killable and replaceable at any moment without data loss.
- Asset bundling and code generation happen at build time, not runtime.

---

## VII. Port Binding

> Export services via port binding

### What It Means

The twelve-factor app is **completely self-contained**. It does not rely on runtime injection of a webserver into the execution environment (no external Apache/Nginx as the app container). Instead, the app **exports HTTP (or any protocol) as a service by binding to a port** and listening for requests.

In development: `http://localhost:8080/`
In production: A routing layer (load balancer, ingress controller) handles public-facing hostname mapping and routes requests to the port-bound process.

This is implemented by adding a webserver library as a dependency:
- **Go**: `net/http` (stdlib), Fiber, Echo, Chi
- **Python**: Tornado, Gunicorn
- **Java**: Jetty, embedded Tomcat
- **Node.js**: Express, Fastify

This applies to **any protocol**, not just HTTP. A NATS client, a gRPC server, a Redis-protocol server -- all export their service via port binding.

One app's port-bound service can become another app's backing service (Factor IV) by providing its URL as a config resource.

### How to Apply

- The app starts its own HTTP/gRPC server. No external webserver container required.
- The port to bind to comes from config (environment variable, e.g., `PORT=8080`).
- In Kubernetes, the container listens on a port; the Service/Ingress routes traffic to it.

---

## VIII. Concurrency

> Scale out via the process model

### What It Means

In the twelve-factor app, **processes are a first-class citizen**. The app uses the Unix process model for running service daemons, assigning each type of work to a **process type**:
- **Web processes** handle HTTP requests
- **Worker processes** handle background jobs
- **Clock processes** handle scheduled tasks

The developer can architect the app to handle diverse workloads by assigning each type of work to a process type. Individual processes may still internally multiplex (Go goroutines, threads, async I/O), but the app **scales horizontally by running more processes**, not by making one process larger.

The array of process types and the number of each is called the **process formation**.

Twelve-factor processes should **never daemonize or write PID files**. Instead, rely on the OS process manager (systemd, Kubernetes, a cloud platform's process manager) to manage output streams, respond to crashed processes, and handle restarts and shutdowns.

### How to Apply

- Scale by increasing replicas of a process type (Kubernetes `replicas: N`), not by giving one instance more CPU/RAM.
- Different workloads run as different process types (separate Deployments in Kubernetes): `web`, `worker`, `scheduler`.
- The app does not daemonize itself, manage PID files, or trap signals for self-restarting. Let the orchestrator handle it.
- Internal concurrency (goroutines, threads) is fine but not a substitute for horizontal scaling.

---

## IX. Disposability

> Maximize robustness with fast startup and graceful shutdown

### What It Means

Processes are **disposable** -- they can be started or stopped at a moment's notice. This facilitates fast elastic scaling, rapid deployment of code/config changes, and robustness of production deploys.

**Fast startup**: Processes should minimize startup time -- ideally a few seconds from launch to ready-to-serve. Short startup time provides more agility for the release process and scaling up.

**Graceful shutdown**: On receiving `SIGTERM`, the process:
1. Stops accepting new requests (closes the listening port)
2. Finishes processing any current requests
3. Exits

For worker processes: return the current job to the work queue (e.g., NACK in RabbitMQ, automatic return in Beanstalkd). All jobs should be **reentrant** (safe to re-execute), typically achieved through transactions or idempotent operations.

**Crash resilience**: Processes should also handle sudden death (hardware failure, OOM kill). Use a robust queueing backend that returns jobs to the queue when a client disconnects or times out. The architecture embraces **crash-only design**.

### How to Apply

- Optimize startup: defer heavy initialization, use connection pooling, minimize bootstrap work.
- Implement graceful shutdown: handle `SIGTERM`, drain in-flight requests, close connections cleanly.
- Make all background jobs idempotent and reentrant -- safe to retry after a crash.
- Kubernetes readiness/liveness probes should reflect actual readiness, not just process existence.
- Use `preStop` hooks or `terminationGracePeriodSeconds` in Kubernetes to allow graceful drain.

---

## X. Dev/Prod Parity

> Keep development, staging, and production as similar as possible

### What It Means

Historically, there are three gaps between development and production:

| Gap | Traditional App | Twelve-Factor App |
|-----|----------------|-------------------|
| **Time gap** | Weeks/months between code and deploy | Hours or minutes |
| **Personnel gap** | Developers write, ops deploy | Developers who write code are closely involved in deploying and observing it in production |
| **Tools gap** | Different stacks in dev vs. prod (e.g., SQLite vs. PostgreSQL) | Same backing services everywhere |

The twelve-factor developer **resists the urge** to use different backing services between environments, even when adapter libraries theoretically smooth over differences. Tiny incompatibilities crop up, causing code that worked in development to fail in production.

Modern tools make running production-equivalent services locally trivial: Docker, Docker Compose, Kubernetes local clusters (minikube, kind).

### How to Apply

- Use the same database, cache, queue, and message broker in dev as in prod. Docker Compose for local development.
- Deploy continuously -- code merged today should reach production today (or within hours).
- Developers observe their code in production (logs, metrics, alerts).
- Never use SQLite locally and PostgreSQL in production, or in-memory cache locally and Redis in production.
- Infrastructure-as-code ensures staging and production are structurally identical.

---

## XI. Logs

> Treat logs as event streams

### What It Means

Logs provide visibility into the behavior of a running app. They are a **stream of aggregated, time-ordered events** collected from the output streams of all running processes and backing services. In their raw form, they are typically one event per line (though backtraces may span multiple lines).

A twelve-factor app **never concerns itself with routing or storage of its output stream**:
- It should not write to or manage logfiles.
- Each running process writes its event stream, **unbuffered, to `stdout`**.

During local development, the developer views the stream in the terminal foreground. In production, the execution environment captures each process's stream and routes it to one or more final destinations (log aggregation service, archival storage) for viewing and long-term analysis. These destinations are not visible to or configurable by the app.

Event streams can be routed to tools like Splunk, Elasticsearch, or Fluentd for:
- Finding specific past events
- Graphing trends (requests per minute)
- Active alerting (error rate exceeds threshold)

### How to Apply

- Log to `stdout` / `stderr`. Period. No log files, no log rotation, no log management in the app.
- Use structured logging (JSON lines) for machine-parseable output.
- In Kubernetes, container stdout is automatically captured. Use Fluentd/Fluent Bit or a log collector to route to Elasticsearch, Loki, etc.
- Logging configuration (level, format) can come from environment variables (Factor III).
- Never use `log.Fatal` (or equivalent) in request handlers -- it kills the process. Use error returns.

---

## XII. Admin Processes

> Run admin/management tasks as one-off processes

### What It Means

Developers frequently need to do one-off administrative or maintenance tasks:
- Running database migrations (e.g., `migrate up`)
- Running a console / REPL to inspect live data
- Running one-time scripts (data fixes, backfill jobs)

These tasks should run as **one-off processes** in an **identical environment** to the app's regular long-running processes:
- Against the same **release** (same code and config)
- Using the same **dependency isolation**
- Against the same **backing services**

Admin code **ships with application code** to avoid synchronization issues. The same codebase, the same image.

The methodology strongly favors languages that provide a REPL shell out of the box and make it easy to run one-off scripts.

### How to Apply

- Database migrations are part of the application codebase and run as a step in the deployment pipeline (init container in Kubernetes, or a CI/CD job using the same image).
- One-off admin tasks run via `kubectl exec` or a Kubernetes Job using the same container image and config.
- Never run admin tasks from a developer laptop against production data. Use the same image, same config, same environment.
- Admin scripts live in the same repository as the application code.

---

## Quick Reference

| # | Factor | Rule | Violation Example |
|---|--------|------|-------------------|
| I | Codebase | 1 repo = 1 app, many deploys | Sharing code by copy-paste between repos |
| II | Dependencies | Declare and isolate all deps | Relying on a globally installed system tool |
| III | Config | Environment variables | Hardcoded DB password in source code |
| IV | Backing Services | Attached resources via config | `if env == "prod" { useRDS() } else { useSQLite() }` |
| V | Build, Release, Run | Immutable releases, strict stages | SSH into prod to hotfix code |
| VI | Processes | Stateless, share-nothing | Sticky sessions, local file uploads |
| VII | Port Binding | Self-contained, export via port | Deploying a WAR file into Tomcat |
| VIII | Concurrency | Scale via process model | Scaling vertically to a bigger VM instead of adding replicas |
| IX | Disposability | Fast startup, graceful shutdown | 5-minute startup, no SIGTERM handling |
| X | Dev/Prod Parity | Same stack everywhere | SQLite in dev, PostgreSQL in prod |
| XI | Logs | stdout, no file management | Writing to `/var/log/app.log` and rotating manually |
| XII | Admin Processes | One-off, same environment | Running migrations from a dev laptop against prod DB |
