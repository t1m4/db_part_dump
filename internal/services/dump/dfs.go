package dump

import (
	"github.com/t1m4/db_part_dump/config"
	"github.com/t1m4/db_part_dump/internal/schemas"
)

func dfsRecursive(resultTablePks []*schemas.Table, visited map[string]bool, table *schemas.Table) []*schemas.Table {
	visited[table.Name] = true
	for fkTableName, fkTable := range table.Fks {
		if _, ok := visited[fkTableName]; ok {
			continue
		}
		resultTablePks = dfsRecursive(resultTablePks, visited, fkTable)
	}
	resultTablePks = append(resultTablePks, table)
	return resultTablePks
}

// Sort tablePksByTable by order of apperence
func dfsSort(tablePksByTable tablePksByTableT, configTables []config.Table) []*schemas.Table {
	resultTablePks := make([]*schemas.Table, 0)
	visited := make(map[string]bool, len(tablePksByTable))

	// Create queue this starting elements
	startingTables := make([]*schemas.Table, 0)
	configTablesSet := make(map[string]bool, len(configTables))
	for _, table := range configTables {
		startingTables = append(startingTables, tablePksByTable[table.Name])
		configTablesSet[table.Name] = true
	}
	for tableName, table := range tablePksByTable {
		if _, ok := configTablesSet[tableName]; ok {
			continue
		}
		startingTables = append(startingTables, table)
	}
	for _, table := range startingTables {
		if _, ok := visited[table.Name]; ok {
			continue
		}
		resultTablePks = dfsRecursive(resultTablePks, visited, table)

	}
	return resultTablePks

}
