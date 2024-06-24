# Kiln - Tezos Delegation Service

# Description

Scrapes Tezos delegations and stores them in a dedicated database

# Technical Setup

## Requirements

- Go 1.22
- PostgreSQL 15 database

## Tooling

(Tern)[github.com/jackc/tern] - Database migrations

``̀bash
go get github.com/jackc/tern
``̀

(sqlc)[] - Database schema-as-code

``̀bash
go get github.com/jackc/tern
``̀

## Run the component

``̀bash
>$ make run
``̀

### Author

- Alexandre Gille (alexandre.teva.gille@gmail.com)