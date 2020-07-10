# di-velocity

## TODO
- Refactor the while thing a few times
- k8s configs
- refine count => score transformation?
- tf for prod db
- ETL script to populate data
- http ingress
- FE

## Code structure

Framework heavily inspired by https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html

### App
Main entrypoint (ingress and egress) for the app. main.go calls this.

Contains:
- `service` struct which initializes and holds references to runtime dependencies
- `handler` functions which are called when messages are received from the broker. These handle external input/out validation and conversion
- `middleware` functions that can wrap handlers and provide additional functionality
- `mappers` convert domain objects to transport formats like protobuf

### Repository
Handles getting data in and out of persistent store(s).

Repository functions accept store connections and optional input. They return domain objects.

### Domain
Holds all domain logic.

These should be pure functions and have no knowledge of data stores, transport mechanisms, etc.
