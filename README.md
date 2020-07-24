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

Inspired by [Netflix' Hexagonal Architecture](https://netflixtechblog.com/ready-for-changes-with-hexagonal-architecture-b315ec967749)

The main difference is that the nats ingress _is_ the interaction layer. If we had a use-case for exposing the internals directly via http or cli, we could include an ingress-agnostic interaction layer.
