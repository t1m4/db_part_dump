package dump

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/t1m4/db_part_dump/config"
	"github.com/t1m4/db_part_dump/internal/constants"
	"github.com/t1m4/db_part_dump/internal/db"
	"github.com/t1m4/db_part_dump/internal/repositories"
	"github.com/t1m4/db_part_dump/internal/schemas"
	"github.com/t1m4/db_part_dump/internal/services/exporter"
)

type fksByTableT map[string][]db.Fk
type tablePksByTableT map[string]*schemas.Table

type DumpService struct {
	c        *config.Config
	repo     repositories.RepositoriesI
	exporter exporter.Exporter
}

func New(c *config.Config, repo *repositories.Repositories) *DumpService {
	return &DumpService{c: c, repo: repo, exporter: exporter.New(c, repo)}
}

// Starting point of service
func (d *DumpService) StartDump(ctx context.Context) error {
	tablePks, err := d.collectTableFkIds(ctx)
	if err != nil {
		return err
	}
	sortedTablePks := dfsSort(tablePks, d.c.Settings.Tables)
	err = d.exporter.ExportToFile(ctx, sortedTablePks)
	if err != nil {
		return err
	}
	return nil

}

// Collect all table pks using config tables
func (d *DumpService) collectTableFkIds(ctx context.Context) (tablePksByTableT, error) {
	fksByTable := make(fksByTableT)
	var err error
	var newTables []config.Table
	tablePksByTable := make(tablePksByTableT, len(d.c.Settings.Tables))
	includeIncomingTables := make(map[string]bool, len(d.c.Settings.IncludeIncomingTables))
	for _, tableName := range d.c.Settings.IncludeIncomingTables {
		includeIncomingTables[tableName] = true
	}
	tablesQueue, err := d.initTables(ctx, tablePksByTable)
	if err != nil {
		return nil, err
	}

	i := 0
	for len(tablesQueue) != 0 {
		slog.Debug("")
		table := tablesQueue[0]
		tablesQueue = tablesQueue[1:]
		slog.Debug(fmt.Sprintf("\nStarted %s", table.Name))
		fks, err := d.getFks(ctx, fksByTable, table.Name, includeIncomingTables[table.Name])
		if err != nil {
			return nil, err
		}
		if len(fks) == 0 {
			continue
		}
		newTables, err = d.getFksIds(ctx, tablePksByTable, fks, table)
		if err != nil {
			return nil, err
		}
		tablesQueue = append(tablesQueue, newTables...)
		i++
	}
	// d.debugTables(tablePksByTable)
	return tablePksByTable, nil
}

// Init config table pk ids
func (d *DumpService) initTables(ctx context.Context, tablePksByTable tablePksByTableT) ([]config.Table, error) {
	tablesQueue := make([]config.Table, len(d.c.Settings.Tables))
	copy(tablesQueue, d.c.Settings.Tables)
	for _, table := range tablesQueue {
		// Get pk column name
		pkColumnName, err := d.repo.GetPKColumnName(ctx, d.c.Settings.SchemaName, table.Name)
		if err != nil {
			return nil, err
		}
		// Get table ids using select pk ids
		pkIdRows, err := d.repo.GetPkIdRows(ctx, d.c.Settings.SchemaName, table, pkColumnName)
		if err != nil {
			return nil, err
		}

		slog.Debug("DATA", "pkIds", pkIdRows)
		currentPkIds := d.createIdsSet(pkIdRows, pkColumnName)
		tablePksByTable[table.Name] = &schemas.Table{
			Name:    table.Name,
			Filters: map[string]schemas.Pks{pkColumnName: currentPkIds},
			Fks:     make(map[string]*schemas.Table, 0),
		}
		slog.Debug("PkIds", table.Name, currentPkIds)
	}
	return tablesQueue, nil
}

// Get table fks by tableName
func (d *DumpService) getFks(
	ctx context.Context,
	fksByTable fksByTableT,
	tableName string,
	isIncludeIncoming bool,
) ([]db.Fk, error) {
	var err error
	fks, ok := fksByTable[tableName]
	if !ok {
		fks, err = d.repo.GetFKs(ctx, d.c.Settings.Direction, d.c.Settings.SchemaName, tableName, isIncludeIncoming)
		if err != nil {
			return nil, err
		}
		fksByTable[tableName] = fks
		slog.Debug("DATA", "fks", fks)
	}
	return fks, nil
}

// Collect fks ids and new tables by fks.
// If table already visited and there is not new pks then do not add to queue again
// Create fks relationships for dfs sorting
func (d *DumpService) getFksIds(
	ctx context.Context,
	tablePksByTable tablePksByTableT,
	fks []db.Fk,
	table config.Table,
) ([]config.Table, error) {
	fkIdRows, err := d.repo.GetFkIdRows(ctx, d.c.Settings.SchemaName, table, fks)
	if err != nil {
		return nil, err
	}
	resultTables := make([]config.Table, 0)
	for _, fk := range fks {
		currentFkIds := d.createIdsSet(fkIdRows, fk.ColumnName)
		if len(currentFkIds) == 0 {
			continue
		}
		slog.Debug("FkIds", fk.ForeignTableName, currentFkIds)
		isVisited := true
		if tablePks, ok := tablePksByTable[fk.ForeignTableName]; ok {
			if _, ok := tablePks.Filters[fk.ForeignColumnName]; !ok {
				tablePks.Filters[fk.ForeignColumnName] = make(map[string]bool, len(currentFkIds))
			}
			for fkId := range currentFkIds {
				if _, ok := tablePks.Filters[fk.ForeignColumnName][fkId]; !ok {
					isVisited = false
					tablePks.Filters[fk.ForeignColumnName][fkId] = true
				}

			}
		} else {
			isVisited = false
			tablePksByTable[fk.ForeignTableName] = &schemas.Table{
				Name:    fk.ForeignTableName,
				Filters: map[string]schemas.Pks{fk.ForeignColumnName: currentFkIds},
				Fks:     make(map[string]*schemas.Table, 0),
			}
		}
		tablePks := tablePksByTable[table.Name]
		if fk.Direction == constants.OUTGOING {
			tablePks.Fks[fk.ForeignTableName] = tablePksByTable[fk.ForeignTableName]
		}
		if !isVisited {
			newTable := config.Table{
				Name:    fk.ForeignTableName,
				Filters: []config.Filter{{Name: fk.ForeignColumnName, Value: buildStringFromSet(currentFkIds)}},
			}
			resultTables = append(resultTables, newTable)
		}
	}
	return resultTables, nil
}

// Create set of ids
func (d *DumpService) createIdsSet(idsRows []map[string]any, columnName string) map[string]bool {
	currentPkIds := make(map[string]bool)
	for _, idsRow := range idsRows {
		currentFkId := idsRow[columnName]
		if currentFkId != nil {
			fkIdString := AnyToString(currentFkId)
			currentPkIds[fkIdString] = true
		}
	}
	return currentPkIds
}

func (d *DumpService) debugTables(tablePksByTable tablePksByTableT) {
	for tableName, table := range tablePksByTable {
		// fmt.Println("TABLE", tableName, table.Filters)
		// for _, fk := range table.Fks {
		// 	fmt.Println("fk", tableName, fk.Name)
		// }
		count := 0
		for _, pks := range table.Filters {
			count += len(pks)
		}
		fmt.Println("TABLE", tableName, count)
	}
}
