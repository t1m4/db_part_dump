package repositories

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/t1m4/db_part_dump/config"
	"github.com/t1m4/db_part_dump/internal/db"
	"github.com/t1m4/db_part_dump/internal/schemas"
)

// Get unique column names
func getFkColumnNames(fks []db.Fk) []string {
	namesSet := make(map[string]bool, 0)
	fkColumnNames := make([]string, 0)
	for _, fk := range fks {
		namesSet[fk.ColumnName] = true
	}
	for fkName := range namesSet {
		fkColumnNames = append(fkColumnNames, fkName)
	}
	return fkColumnNames
}

// Build condition based on tabl filters
func buildFilterCondition(table config.Table) string {
	var condition strings.Builder
	if len(table.Filters) != 0 {
		condition.WriteString(" where ")
	}
	for _, filter := range table.Filters {
		condition.WriteString(fmt.Sprintf("%s in (%s)", filter.Name, filter.Value))
	}
	return condition.String()
}

func buildTableNameWithSchema(schemaName string, tableName string) string {
	if schemaName == "" {
		return tableName
	}
	return fmt.Sprintf("%s.%s", schemaName, tableName)
}

func getManyRows(rows *sql.Rows, columnNames []string) ([]map[string]any, error) {
	resultRows := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columnNames))
		valuePtrs := make([]any, len(columnNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		rowMap := make(map[string]any)
		for i, columnName := range columnNames {
			value := values[i]
			rowMap[columnName] = value
		}
		resultRows = append(resultRows, rowMap)
	}
	return resultRows, nil
}

func buildPkCondition(pkTable *schemas.Table) string {
	conditions := make([]string, 0)
	for name, pks := range pkTable.Filters {
		i := 0
		filterPks := make([]string, len(pks))
		for pk := range pks {
			filterPks[i] = pk
			i++
		}
		conditions = append(conditions, fmt.Sprintf("%s in (%s)", name, strings.Join(filterPks, ", ")))
	}
	return fmt.Sprintf(" WHERE %s", strings.Join(conditions, " OR "))

}

// TODO check is it complete
func AnyToPsqlString(value any) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "t"
		}
		return "f"
	case nil:
		return "\\N"
	case []byte:
		return string(v)
	case time.Time:
		return v.Format(time.RFC3339Nano)
	case string:
		escapedV := strconv.Quote(v)
		escapedV = escapedV[1 : len(escapedV)-1]
		return fmt.Sprintf("%v", escapedV)
	default:
		return fmt.Sprintf("%v", v)
	}
}
