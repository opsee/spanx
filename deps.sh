#!/bin/bash

docker pull sameersbn/postgresql:9.4-3
docker run --name postgresql -d -e PSQL_TRUST_LOCALNET=true -e DB_USER=postgres -e DB_PASS= -e DB_NAME=spanx_test sameersbn/postgresql:9.4-3
echo "started postgresql"
