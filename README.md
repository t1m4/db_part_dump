# Dump part dump
dump tables data based on fk(Foreign keys) with given fields and values

## Features
- PostgreSQL dump format 
- Incoming fks. Fetch reversed relationships for all tables
- Handle cycles removing and restoring constraints

## Algorithm
Collect all pks and filters for all related table starting from initial table and create graph of table with fks. 

Dump result graph of table to the file


### Usage 
- dump data using config 
```
go run cmd/main.go -c config.yaml
```
- restore data using pg_dump 
```
pg_dump -d db_part_dump --schema-only > schema_only.sql
psql -d db_part_dump < backups/test.sql
```
 
### Config params 
- `schema_name` - name of schema name for PostgreSQL
- `tables` - array of tables to start dump
- `direction` - choices are outgoing/incoming. outgoing only fks that have in tables. incoming include tables that referencing current table.
- `include_incoming_tables` - including table in outgoing mode to use as incoming tables


