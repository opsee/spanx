#!/bin/bash
set -e

echo "loading schema for tests..."
echo "drop database if exists spanx_test; create database spanx_test" | psql -U postgres -h postgresql
migrate -url $POSTGRES_CONN -path ./migrations up
