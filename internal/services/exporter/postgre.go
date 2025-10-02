package exporter

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/t1m4/db_part_dump/config"
	"github.com/t1m4/db_part_dump/internal/repositories"
	"github.com/t1m4/db_part_dump/internal/schemas"
)

type Exporter interface {
	ExportToFile(ctx context.Context, tablePks []*schemas.Table) error
}

type PostgresqlExporter struct {
	c    *config.Config
	repo repositories.RepositoriesI
}

func New(c *config.Config, repo *repositories.Repositories) Exporter {
	return &PostgresqlExporter{c: c, repo: repo}
}
func (d *PostgresqlExporter) createFile(filename string) (*os.File, error) {
	if filename == "" {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("backup_%s.sql", timestamp)
	}
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Export to file
func (d *PostgresqlExporter) ExportToFile(ctx context.Context, tablePks []*schemas.Table) error {
	file, err := d.createFile(d.c.Settings.Output)
	defer file.Close()
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)

	for _, tablePk := range tablePks {
		slog.Debug("")
		slog.Debug("ExportSQL", tablePk.Name, tablePk.Filters)
		err := d.repo.GetRows(ctx, d.c.Settings.SchemaName, tablePk, writer)
		if err != nil {
			return err
		}
	}
	slog.Info(fmt.Sprintf("Export to %s finished", file.Name()))
	return nil
}
