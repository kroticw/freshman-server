package cmd

import (
	"strconv"

	"github.com/kroticw/freshman-server.git/infrastructure/sql"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Применение или откат миграций",
	Long:  `Команда для применения или отката миграций БД`,
	Run:   runMigrateCmd,
}

var migrateCmdUp bool
var migrateCmdDown int

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().BoolVar(&migrateCmdUp, "up", true, "Накатить миграции. Это действие по умолчанию")
	migrateCmd.Flags().IntVar(&migrateCmdDown, "down", 0, "Откат миграций. Значение - количество миграций для отката")
}

func runMigrateCmd(_ *cobra.Command, _ []string) {
	if migrateCmdUp {
		runMigrateUp()
		return
	}

	if migrateCmdDown > 0 {
		runMigrateDown(migrateCmdDown)
		return
	}
	logger.Fatalln("Действие не передано")
}

func runMigrateUp() {
	logger.Infoln("Применение новых миграций...")
	sql.MigrateUp(cfg.Database.URL, logger)
}

func runMigrateDown(count int) {
	logger.Infoln("Откат " + strconv.Itoa(count) + " миграций...")
	sql.MigrateDown(cfg.Database.URL, count, logger)
}
