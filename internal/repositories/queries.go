package repositories

var GetTableOutgoingFks string = `
SELECT 
    att.attname AS column_name,
    ref_nsp.nspname AS foreign_table_schema, 
    ref_tbl.relname AS foreign_table_name,
    ref_att.attname AS foreign_column_name,
    'outgoing' AS direction
FROM pg_constraint con
JOIN pg_class tbl ON con.conrelid = tbl.oid
JOIN pg_namespace nsp ON tbl.relnamespace = nsp.oid AND nsp.nspname = '%s'
JOIN pg_attribute att ON att.attrelid = tbl.oid 
                     AND att.attnum = ANY(con.conkey)
                     AND NOT att.attisdropped
JOIN pg_class ref_tbl ON con.confrelid = ref_tbl.oid
JOIN pg_namespace ref_nsp ON ref_tbl.relnamespace = ref_nsp.oid
JOIN pg_attribute ref_att ON ref_att.attrelid = ref_tbl.oid 
                         AND ref_att.attnum = ANY(con.confkey)
                         AND NOT ref_att.attisdropped
WHERE con.contype = 'f'
  AND tbl.relname = '%s'
`

var GetTableAllFks string = `
SELECT 
    att.attname AS column_name,
    ref_nsp.nspname AS foreign_table_schema, 
    ref_tbl.relname AS foreign_table_name,
    ref_att.attname AS foreign_column_name,
    'outgoing' AS direction
FROM pg_constraint con
JOIN pg_class tbl ON con.conrelid = tbl.oid
JOIN pg_namespace nsp ON tbl.relnamespace = nsp.oid AND nsp.nspname = '%s'
JOIN pg_attribute att ON att.attrelid = tbl.oid 
                     AND att.attnum = ANY(con.conkey)
                     AND NOT att.attisdropped
JOIN pg_class ref_tbl ON con.confrelid = ref_tbl.oid
JOIN pg_namespace ref_nsp ON ref_tbl.relnamespace = ref_nsp.oid
JOIN pg_attribute ref_att ON ref_att.attrelid = ref_tbl.oid 
                         AND ref_att.attnum = ANY(con.confkey)
                         AND NOT ref_att.attisdropped
WHERE con.contype = 'f'
  AND tbl.relname = '%s'

UNION ALL

SELECT 
    ref_att.attname AS column_name,
    nsp.nspname AS foreign_table_schema,
    tbl.relname AS foreign_table_name,
    att.attname AS foreign_column_name,
    'incoming' AS direction
FROM pg_constraint con
JOIN pg_class ref_tbl ON con.confrelid = ref_tbl.oid
JOIN pg_namespace ref_nsp ON ref_tbl.relnamespace = ref_nsp.oid AND ref_nsp.nspname = '%s'
JOIN pg_attribute ref_att ON ref_att.attrelid = ref_tbl.oid 
                         AND ref_att.attnum = ANY(con.confkey)
                         AND NOT ref_att.attisdropped
JOIN pg_class tbl ON con.conrelid = tbl.oid
JOIN pg_namespace nsp ON tbl.relnamespace = nsp.oid
JOIN pg_attribute att ON att.attrelid = tbl.oid 
                     AND att.attnum = ANY(con.conkey)
                     AND NOT att.attisdropped
WHERE con.contype = 'f'
  AND ref_tbl.relname = '%s';
`

var GetTablePkColumnName string = `
SELECT 
    kcu.column_name
FROM 
    information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu 
        ON tc.constraint_name = kcu.constraint_name
WHERE 
    tc.constraint_type = 'PRIMARY KEY'
    AND tc.table_name = '%s'
    AND tc.table_schema = '%s'
`

var Select = "SELECT %s FROM %s"
