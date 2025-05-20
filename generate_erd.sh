#!/bin/bash

echo "[+] Applying schema.sql to local DB..."
psql -U postgres -d Orbit -f schema.sql

echo "[+] Generating ERD with SchemaSpy..."
docker run --rm -v "$PWD/erd:/output" schemaspy/schemaspy \
  -t pgsql \
  -host host.docker.internal \
  -port 5432 \
  -db Orbit \
  -u postgres \
  -p password \
  -s public

echo "[+] Done. View: erd/index.html"