# Kiln - Tezos Delegation Service

## Description

Simple Tezos delegation operations scraper that stores them in a PostgreSQL database and exposes data on a REST API endpoint.

## Technical Setup

### Requirements

The following must be properly set-up on the machine running the component
- [Go 1.22](https://go.dev/doc/install)
- [Docker Engine](https://docs.docker.com/engine/) with [Docker Compose plugin](https://docs.docker.com/compose/install/linux/) (not the [docker-compose tool](https://docs.docker.com/compose/install/standalone/))

### Tooling

The `$GOPATH/bin` directory must be in the `PATH`.

This project uses [Tern](github.com/jackc/tern) for database migrations.

Install everything

```bash
make install
```

## Run the component

### Environment
- `API_ADDR` is the address listened to by the REST API server
- `DB_HOST`, `DB_DATABASE`, `DB_USER`, `DB_PASSWORD` are the database's host name, database name, user and password
- `TZKT_BASE_URL` is the TzKT API URL to scrap from (trailing slash is **mandatory**)
- `SCRAP_SINCE` is the starting date and time of scraping in RFC3339 format (e.g. `2024-06-26T19:14:33Z`). When set, the component will not fetch the most recent block's timestamp from storage and use this value instead.

### First run

These steps must be followed if the project has never been run on the current machine.

```bash
# default settings for docker compose and tern
# adapt to your needs in repository/tern.conf and build/docker-compose.yaml
export API_ADDR="localhost:8080"
export DB_HOST="localhost:5432"
export DB_DATABASE="kiln-tezos"
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export TZKT_BASE_URL="https://api.tzkt.io/"

# setup technical stack
make docker-up
make migrate-up

# shutdown / cleanup
make docker-down
```

### Run

Run the delegation scraper and the service

```bash
make run
# or
go run main.go
```

Then play

```bash
curl http://localhost:8080/xtz/delegations?year=2024
```

See [OpenAPI - Swagger](swagger.yaml) for more details

## Testing

Tests are a mix of unit tests and standalone integration tests. No initial environment is needed.

```bash
make test
```

Note: `repository` package is not tested as it would needed complex low-level mocking with libraries like `pgmock`. Integration tests will make this easier.

## Design choices

**REST API**

The REST API is split into three parts: "server" for low-level HTTP matters, "handler" for HTTP request handling and "controller" for addressing any business logic.
THis is to apply a certain level of "separation of concerns" at the early stage of the project, instead of having one single file gathering all three.
It might imply a greater number of code files, but as the project might grow, it would keep things clear. In a big project, it these parts might become
separate packages.

No library/framework has been used to build this REST API to keep things simple, as required. Gin would be a great fit otherwise.

**Executable**

For simplicity at various levels, the REST API server and the TzKT scraper run in different Go routines but in the same -unique- executable.
It might be useful to have one executable each for the sake of robustness, since an issue on one side might drag the other side down with it:
memory leaks, panics, orchestration effects ...
Resource provisioning and scaling ability would also be improved and fine-tuned.

The scraper is a daemon and not a one-shot executable to permit a wider range of scraping intervals. Having a Kubernetes CRON job would not be suitable for short intervals
and not always reliable in some infrastructure transient contexts.

The entire component stop with an error message on any unmanageable issue, whoever failed between the API or the scraper.

**Scraping**

Cyclic scraping relies on a `time.Ticker` to ensure the regular aggregation of data (CRON-like behaviour). Tickers have the down side of not triggering as soon as they are created,
meaning that the first scraping cycle will always start after waiting for one interval after starting the executable. It can be easily worked around but will make code more complex to read.

Listening to messages of the TzKT WebSocket API appears to greatly improve the conception of the scraper.

**Storage**

PostgreSQL was a good fit for this purpose since it is a performant database that provides efficient and easy to use features, for example the `ON CONSTRAINT ...` statement or
the ability of doing [functional indexes](https://www.postgresql.org/docs/current/indexes-expressional.html) for fast by-year filtering.

Using the `repository` package is safe from concurrency.

## Possible optimizations & improvements

- add REST API **versioning**
- use TzKT WebSocket API for real-time updates
- expose a gRPC endpoint for inter-service efficient calls
- `Dockerfile` and Helm packaging for deployments
- functional index on block timestamp year:
`CREATE INDEX idx_delegation_block_timestamp_year ON delegation ((EXTRACT(YEAR FROM block_timestamp AT TIME ZONE 'UTC')));`
- contextual logging for better log management
- leveled logging for debugging
- environment variable to set the maximum number of returned delegations from TzKT API to increase the data aggregation rate when needed (catching-up)
- add `page`/`offset` and `size`/`limit` query parameters to the REST API for finer queries
- add rate-limiting/throttling on REST APIs
- use of [Gin](https://github.com/gin-gonic/gin) for simpler request management and middleware support, if more endpoints are needed
- add Prometheus metrics for monitoring and alerting
- security: pass database password in a more secure way

## Author

- Alexandre Gille (alexandre.teva.gille@gmail.com)
