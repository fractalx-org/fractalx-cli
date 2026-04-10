# fractalx-cli

**FractalX CLI** — a command-line tool that generates a Spring Boot modular monolith pre-annotated with [FractalX](https://github.com/fractalx/FractalX) decomposition markers. When you're ready to scale, run `mvn fractalx:decompose` to split it into production-ready microservices.

Think of it as `start.spring.io`, but purpose-built for the decomposition-first workflow.

---

## Installation

### Homebrew (macOS / Linux)

```bash
brew install fractalx/tap/fractalx-cli
```

### curl (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/fractalx/fractalx-cli/main/install.sh | sh
```

### Go install

```bash
go install github.com/fractalx/fractalx-cli@latest
```

### Manual download

Download the binary for your platform from [GitHub Releases](https://github.com/fractalx/fractalx-cli/releases), extract, and place it on your `$PATH`.

---

## Quick start

```bash
fractalx
```

This launches an interactive 6-step wizard. At the end it downloads a `.zip` with a ready-to-compile Spring Boot project.

---

## Usage

```
fractalx [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--from <file>` | — | Load a `fractalx.yaml` spec and skip the wizard |
| `--output <dir>` | `.` | Directory where the ZIP (or project folder) is written |
| `--no-zip` | false | Write files directly to disk instead of a ZIP archive |
| `--help` | — | Show help |

### Examples

```bash
# Interactive wizard → my-platform.zip in current directory
fractalx

# Re-generate from an existing spec, write directly to disk
fractalx --from fractalx.yaml --no-zip

# Write ZIP to a specific folder
fractalx --output ~/projects
```

---

## The wizard (6 steps)

### Step 1 — Project

Basic Maven coordinates and runtime versions.

```
Group ID         com.example
Artifact ID      my-platform
Version          1.0.0-SNAPSHOT
Description      FractalX modular monolith
Spring Boot      3.3.0
Java             21
```

### Step 2 — Services

Define one or more bounded-context services. Each service gets its own package, module marker class, database config, and (optionally) JPA/MongoDB entities.

```
Service name     order-service
Port             8081
Database         postgresql

  Entity: Order
    createdAt    LocalDateTime
    totalAmount  BigDecimal
    status       String
```

Supported databases: `h2`, `postgresql`, `mysql`, `mongodb`, `redis`

### Step 3 — Dependencies

Choose which services call which. The wizard detects circular dependencies and refuses to generate if any are found.

```
'payment-service' depends on:  [ ] order-service
```

### Step 4 — Sagas (optional)

Configure distributed sagas that will be annotated with `@DistributedSaga`.

```
Saga ID              place-order-saga
Owner service        order-service
Compensation method  cancelOrder

Steps:
  1. payment-service  →  charge()
  2. inventory-service → reserve()
```

### Step 5 — Infrastructure

Toggle infrastructure components to include in the generated project.

| Component | What it generates |
|-----------|-------------------|
| API Gateway | `fractalx-gateway` config (port 9999) |
| Admin Dashboard | `fractalx-admin` config (port 9090) |
| Service Registry | Eureka-compatible registry config |
| Docker Compose | `docker-compose.dev.yml` with DB + observability containers |
| GitHub Actions CI | `.github/workflows/ci.yml` |
| Kubernetes manifests | `k8s/{service}-deployment.yml` per service |
| Observability | Jaeger + OpenTelemetry config |
| Saga Orchestrator | `fractalx-saga-orchestrator` config |

### Step 6 — Security

Pick an authentication strategy.

| Option | What's added |
|--------|-------------|
| `none` | No security config |
| `jwt` | Spring Security + `jjwt` dependency + bearer config |
| `oauth2` | Spring OAuth2 Resource Server + JWKS URI config |
| `apikey` | API key filter config |

---

## Generated project structure

```
my-platform/
├── pom.xml
├── fractalx.yaml                        ← round-trip spec (re-usable with --from)
├── README.md
├── docker-compose.dev.yml               (if Docker selected)
├── .github/workflows/ci.yml             (if CI selected)
├── k8s/
│   └── order-service-deployment.yml     (if Kubernetes selected)
└── src/
    ├── main/
    │   ├── java/com/example/myplatform/
    │   │   ├── MyPlatformApplication.java
    │   │   └── order/
    │   │       ├── OrderModule.java      ← @DecomposableModule marker
    │   │       ├── Order.java            ← JPA entity
    │   │       ├── OrderRepository.java
    │   │       ├── OrderController.java
    │   │       └── OrderService.java
    │   └── resources/
    │       ├── application.yml
    │       ├── application-dev.yml
    │       ├── fractalx-config.yml
    │       └── db/migration/V1__init.sql
    └── test/
        ├── java/com/example/myplatform/
        │   └── MyPlatformApplicationTests.java
        └── resources/
            └── application.yml          ← H2 in-memory for tests
```

---

## Non-interactive mode (`--from`)

The generated `fractalx.yaml` is a complete round-trip spec. You can edit it and regenerate at any time:

```bash
# Edit fractalx.yaml to add a new service, then regenerate
fractalx --from fractalx.yaml --no-zip --output ./regenerated
```

`fractalx.yaml` format:

```yaml
project:
  groupId: com.example
  artifactId: my-platform
  version: "1.0.0-SNAPSHOT"
  javaVersion: "21"
  springBootVersion: 3.3.0
  description: "FractalX modular monolith"

services:
  - name: order-service
    port: 8081
    database: postgresql
    entities:
      - name: Order
        fields:
          - createdAt: LocalDateTime
          - totalAmount: BigDecimal

sagas:
  - id: place-order-saga
    owner: order-service
    compensationMethod: cancelOrder
    timeoutMs: 30000
    steps:
      - service: payment-service
        method: charge

infrastructure:
  gateway: true
  admin: true
  serviceRegistry: true
  docker: true
  kubernetes: false
  ci: github-actions
  observability: true
  sagaOrchestrator: true

security:
  type: jwt
```

---

## After generation

```bash
# Unzip and enter the project
unzip my-platform.zip && cd my-platform

# Start infrastructure (if Docker was selected)
docker compose -f docker-compose.dev.yml up -d

# Run the monolith
mvn spring-boot:run -Dspring-boot.run.profiles=dev

# When ready to decompose into microservices
mvn fractalx:decompose
```

---

## How decomposition works

Each service boundary is marked with `@DecomposableModule`:

```java
@DecomposableModule(
    serviceName = "order-service",
    port = 8081,
    ownedSchemas = {"order_db"},
    independentDeployment = true
)
public class OrderModule {}
```

Running `mvn fractalx:decompose` reads these markers, validates the dependency graph, and generates fully independent Spring Boot microservices — each with its own `pom.xml`, database config, Flyway migrations, Docker setup, and observability config.

---

## Validation

Before generating, `fractalx` checks:

- **Circular dependencies** — detected via DFS; generation is blocked if any cycle exists
- **Port uniqueness** — two services cannot share a port
- **Saga owners** — saga owner and step services must exist in the service list

---

## Contributing

Issues and pull requests welcome at [github.com/fractalx/fractalx-cli](https://github.com/fractalx/fractalx-cli).

## License

Copyright 2024 FractalX

Licensed under the [Apache License, Version 2.0](LICENSE) (the "License"); you may not use this software except in compliance with the License. You may obtain a copy of the License at:

```
http://www.apache.org/licenses/LICENSE-2.0
```

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the [LICENSE](LICENSE) file for the specific language governing permissions and limitations under the License.
