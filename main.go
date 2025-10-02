package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/t1m4/db_part_dump/config"
	"github.com/t1m4/db_part_dump/internal/db"
	"github.com/t1m4/db_part_dump/internal/repositories"
	"github.com/t1m4/db_part_dump/internal/services/dump"

	"github.com/spf13/cobra"
)

var (
	configPath string
)

func main() {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	var rootCmd = &cobra.Command{
		Use:   "db_part_dump",
		Short: "PostgreSQL backup utility with dependency resolution",
		Run:   RunRoot,
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Config file path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func RunRoot(_ *cobra.Command, args []string) {
	c, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := db.NewDB(ctx, c)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	repo := repositories.New(db)
	service := dump.New(c, repo)
	err = service.StartDump(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		log.Fatal(err)
	}
}
