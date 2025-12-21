package cmd

import (
	"context"
	"os"

	"github.com/exaring/otelpgx"
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kroticw/freshman-server/infrastructure/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
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
	SourceStorage StorageDriverConfig `json:"sourceStorage" yaml:"sourceStorage" mapstructure:"sourceStorage"`
}

type StorageDriverConfig struct {
	Type       string `json:"type" yaml:"type" mapstructure:"type"`
	FileSystem *struct {
		BaseDir string `json:"baseDir" yaml:"baseDir" mapstructure:"baseDir"`
	} `json:"fileSystem" yaml:"fileSystem" mapstructure:"fileSystem"`
	S3 *struct {
		Endpoint        string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`
		Region          string `json:"region" yaml:"region" mapstructure:"region"`
		AccessKeyID     string `json:"accessKeyId" yaml:"accessKeyId" mapstructure:"accessKeyId"`
		SecretAccessKey string `json:"secretAccessKey" yaml:"secretAccessKey" mapstructure:"secretAccessKey"`
		ForcePathStyle  bool   `json:"forcePathStyle" yaml:"forcePathStyle" mapstructure:"forcePathStyle"`
		Bucket          string `json:"bucket" yaml:"bucket" mapstructure:"bucket"`
		BasePath        string `json:"basePath" yaml:"basePath" mapstructure:"basePath"`
	} `json:"s3" yaml:"s3" mapstructure:"s3"`
	// MaxSize актуален только для кэширующего хранилища (в байтах)
	MaxSize            *uint64  `json:"maxSize" yaml:"maxSize" mapstructure:"maxSize"`
	RetentionThreshold *float64 `json:"retentionThreshold" yaml:"retentionThreshold" mapstructure:"retentionThreshold"`
}

var (
	logger           *logrus.Logger
	cfg              configuration
	dbConn           *pgxpool.Pool
	filesystemDriver *storage.FilesystemDriver = nil
	s3Driver         *storage.S3Driver         = nil
)

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
	cobra.OnInitialize(initLogger, initConfig, initDatabase, initStorageDriver)
	rootCmd.PersistentFlags().StringVar(&cfgFile,
		"config", "", "config file (default is ./config.yaml)")
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

func initStorageDriver() {
	switch cfg.SourceStorage.Type {
	case "filesystem":
		if cfg.SourceStorage.FileSystem == nil {
			logger.Fatalln("filesystem driver selected but sourceStorage.fileSystem is not configured")
		}
		if cfg.SourceStorage.FileSystem.BaseDir == "" {
			logger.Fatalln("filesystem driver selected but sourceStorage.fileSystem.baseDir is empty")
		}
		filesystemDriver = storage.NewFilesystemDriver(cfg.SourceStorage.FileSystem.BaseDir, logger)
	case "s3":
		if cfg.SourceStorage.S3 == nil {
			logger.Fatalln("s3 driver selected but sourceStorage.s3 is not configured")
		}
		if cfg.SourceStorage.S3.Endpoint == "" ||
			cfg.SourceStorage.S3.Region == "" ||
			cfg.SourceStorage.S3.AccessKeyID == "" ||
			cfg.SourceStorage.S3.SecretAccessKey == "" ||
			cfg.SourceStorage.S3.Bucket == "" {
			logger.Fatalln("s3 driver selected but required fields are empty " +
				"(need endpoint, region, accessKeyId, secretAccessKey, bucket)")
		}
		s3Driver = storage.NewS3(
			cfg.SourceStorage.S3.Region,
			cfg.SourceStorage.S3.Endpoint,
			cfg.SourceStorage.S3.AccessKeyID,
			cfg.SourceStorage.S3.SecretAccessKey,
			cfg.SourceStorage.S3.ForcePathStyle,
			cfg.SourceStorage.S3.Bucket,
			cfg.SourceStorage.S3.BasePath,
			otel.GetTracerProvider().Tracer("S3"),
			logger,
		)
	default:
		if cfg.SourceStorage.Type == "" {
			logger.Fatalln("storage driver type is not configured " +
				"(set sourceStorage.type to filesystem|s3)")
		}
		logger.Fatalf("unknown storage driver %s", cfg.SourceStorage.Type)
	}
}
