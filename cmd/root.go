package cmd

import (
	"context"
	"os"

	"github.com/exaring/otelpgx"
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

type configuration struct {
	Web struct {
		Enable bool   `mapstructure:"enable"`
		Listen string `mapstructure:"listen"`
	}
	Logger struct {
		LogLevel int `mapstructure:"logLevel"`
	} `mapstructure:"logger"`
	Database struct {
		URL string `json:"url" mapstructure:"url"`
	} `mapstructure:"database"`
	OpenTelemetry *struct {
		Enabled   bool    `mapstructure:"enabled"`
		Endpoint  string  `mapstructure:"endpoint"`
		TraceRate float64 `mapstructure:"traceRate"`
	} `mapstructure:"openTelemetry"`
}

var logger *logrus.Logger
var cfg configuration
var dbConn *pgxpool.Pool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "freshman-server.git",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger, initConfig, initDatabase)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigType("yaml")
		viper.SetConfigName("config.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.WithField("filename", viper.ConfigFileUsed()).Infoln("Выбранный конфиг-файл")
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		logger.WithError(err).Fatalln("Ошибка парсинга конфига")
	}

	if cfg.Logger.LogLevel > 0 {
		logger.SetLevel(logrus.Level(cfg.Logger.LogLevel))
	}
}

func initLogger() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyMsg: "log",
		},
	})
}

func initDatabase() {
	var err error
	config, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		logger.WithError(err).Fatalln("Некорректная конфигурация pgx")
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		return nil
	}
	config.ConnConfig.Tracer = otelpgx.NewTracer()
	dbConn, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.WithError(err).Fatalln("Не удалось подключиться к БД")
	}
	logger.Infoln("Подключение к БД успешно")
}
