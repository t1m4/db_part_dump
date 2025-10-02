package repositories

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/t1m4/db_part_dump/config"
	"github.com/t1m4/db_part_dump/internal/constants"
	"github.com/t1m4/db_part_dump/internal/db"
	"github.com/t1m4/db_part_dump/internal/schemas"
)

type RepositoriesI interface {
	GetPKColumnName(ctx context.Context, schemaName string, tableName string) (string, error)
	GetPkIdRows(ctx context.Context, schemaName string, table config.Table, pkColumnName string) ([]map[string]any, error)
	GetFKs(
		ctx context.Context,
		direction string,
		schemaName string,
		tableName string,
		isIncludeIncoming bool,
	) ([]db.Fk, error)
	GetFkIdRows(ctx context.Context, schemaName string, table config.Table, fks []db.Fk) ([]map[string]any, error)
	GetRows(
		ctx context.Context,
		schemaName string,
		pkTable *schemas.Table,
		writer *bufio.Writer,
	) error
}

type Repositories struct {
	db *sql.DB
}

func New(db *sql.DB) *Repositories {
	return &Repositories{db: db}
}

func (r *Repositories) GetPKColumnName(ctx context.Context, schemaName string, tableName string) (string, error) {
	query := fmt.Sprintf(GetTablePkColumnName, tableName, schemaName)
	slog.Debug("SQL", "GetTablePkColumnName", query)
	var columnName string
	err := r.db.QueryRowContext(ctx, query).Scan(&columnName)
	if err != nil {
		return "", err
	}
	return columnName, nil
}

func (r *Repositories) GetPkIdRows(ctx context.Context, schemaName string, table config.Table, pkColumnName string) ([]map[string]any, error) {
	condition := buildFilterCondition(table)
	query := fmt.Sprintf(Select, pkColumnName, buildTableNameWithSchema(schemaName, table.Name))
	query += condition
	slog.Debug("SQL", "GetPkIds", query)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pkIdRows, err := getManyRows(rows, []string{pkColumnName})
	if err != nil {
		return nil, err
	}
	return pkIdRows, nil
}

func (r *Repositories) GetFKs(
	ctx context.Context,
	direction string,
	schemaName string,
	tableName string,
	isIncludeIncoming bool,
) ([]db.Fk, error) {
	var query string
	if direction == constants.OUTGOING && !isIncludeIncoming {
		query = fmt.Sprintf(GetTableOutgoingFks, schemaName, tableName)
	} else {
		query = fmt.Sprintf(GetTableAllFks, schemaName, tableName, schemaName, tableName)
	}
	slog.Debug("SQL", "GetTableFks", query)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fks := make([]db.Fk, 0)
	for rows.Next() {
		fk := db.Fk{}
		if err := rows.Scan(&fk.ColumnName, &fk.ForeignTableSchema, &fk.ForeignTableName, &fk.ForeignColumnName, &fk.Direction); err != nil {
			return nil, err
		}
		fks = append(fks, fk)
	}
	return fks, nil
}

func (r *Repositories) GetFkIdRows(ctx context.Context, schemaName string, table config.Table, fks []db.Fk) ([]map[string]any, error) {
	fkColumnNames := getFkColumnNames(fks)
	fkColumnNamesString := strings.Join(fkColumnNames, ", ")
	condition := buildFilterCondition(table)
	query := fmt.Sprintf(Select, fkColumnNamesString, buildTableNameWithSchema(schemaName, table.Name))
	query += condition
	slog.Debug("SQL", "GetFkIds", query)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fkIdRows, err := getManyRows(rows, fkColumnNames)
	if err != nil {
		return nil, err
	}
	return fkIdRows, nil
}

func (r *Repositories) GetRows(
	ctx context.Context,
	schemaName string,
	pkTable *schemas.Table,
	writer *bufio.Writer,
) error {
	// TODO add ordering
	tableName := buildTableNameWithSchema(schemaName, pkTable.Name)
	query := fmt.Sprintf(Select, "*", tableName)
	condition := buildPkCondition(pkTable)
	query += condition
	slog.Debug("SQL", "GetRows", query)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	for i, columnName := range columns {
		columns[i] = strconv.Quote(columnName)
	}
	if err != nil {
		return err
	}

	fileColumns := strings.Join(columns, ", ")
	slog.Debug("FILE", "Columns", fileColumns)
	var copyStatement strings.Builder
	copyStatement.WriteString(fmt.Sprintf("COPY %s (%s) FROM stdin;\n", tableName, fileColumns))
	comments := fmt.Sprintf("-- Data for Name: %s; Type: TABLE DATA;\n", tableName)
	writer.WriteString(comments)
	writer.WriteString(fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER ALL;\n", tableName))
	writer.WriteString(copyStatement.String())
	count := 0
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err = rows.Scan(valuePtrs...); err != nil {
			return err
		}
		var rowBuf strings.Builder
		for i := range columns {
			value := values[i]
			rowBuf.WriteString(fmt.Sprintf("%s\t", AnyToPsqlString(value)))
		}
		rowString := rowBuf.String()
		rowString = rowString[:len(rowString)-1] + "\n"
		writer.WriteString(rowString)
		if count%1000 == 0 {
			err = writer.Flush()
			if err != nil {
				return err
			}
			count = 0
		}
		count++
	}
	writer.WriteString("\\.\n")
	writer.WriteString(fmt.Sprintf("ALTER TABLE %s ENABLE TRIGGER ALL;\n\n\n", tableName))
	if count != 0 {
		err = writer.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}
