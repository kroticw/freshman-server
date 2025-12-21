package cmd

import (
	"context"

	"github.com/kroticw/freshman-server/infrastructure/sql"
	"github.com/kroticw/freshman-server/infrastructure/transport/http"
	"github.com/kroticw/freshman-server/internal/music"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "запуск сервера",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(_ *cobra.Command, _ []string) {
	var driver music.Storage
	switch cfg.SourceStorage.Type {
	case "filesystem":
		if filesystemDriver == nil {
			logger.Fatalln("filesystem driver selected but driver is not initialized")
		}
		driver = filesystemDriver
	case "s3":
		if s3Driver == nil {
			logger.Fatalln("s3 driver selected but driver is not initialized")
		}
		driver = s3Driver
	default:
		logger.Fatalf("unknown storage driver %s", cfg.SourceStorage.Type)
	}
	musicRepo := sql.NewMusicRepo(dbConn)
	musSvc := music.NewMusicService(driver, musicRepo, logger)
	router := http.SetupRouter(context.Background(), musSvc, logger)
	if cfg.Web.Enable {
		err := router.Run(cfg.Web.Listen)
		if err != nil {
			logger.Fatal(err)
		}
	}
}
