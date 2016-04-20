#!/bin/bash
set -e

echo "loading schema for tests..."
echo "drop database if exists spanx_test; create database spanx_test" | psql -U postgres -h postgres
migrate -url $POSTGRES_CONN -path ./migrations up

go run generate.go
