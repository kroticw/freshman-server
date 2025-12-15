package sql

import (
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	log "github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var migrations embed.FS

func MigrateUp(dsn string, logger *log.Logger) {
	d, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		logger.WithError(err).Fatalln("Не удалось подключиться к БД")
	}
	m.Log = &migrateLogger{logger: logger}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Infoln("Нет новых миграций для применения")
			return
		}
		logger.WithError(err).Fatalln("Не удалось применить миграции")
	}
	logger.Infoln("Миграции успешно применены")
}

func MigrateDown(dsn string, count int, logger *log.Logger) {
	d, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		log.Fatal(err)
	}
	m.Log = &migrateLogger{logger: logger}

	err = m.Steps(-count)
	if err != nil {
		logger.WithError(err).Fatalln("Не удалось откатить миграции")
	}
	logger.Infoln("Откат миграций был успешно произведен")
}

type migrateLogger struct {
	logger *log.Logger
}

// Printf is like fmt.Printf
func (l *migrateLogger) Printf(format string, v ...interface{}) {
	l.logger.WithField("category", "migration").Printf(format, v...)
}

// Verbose should return true when verbose logging output is wanted
func (l *migrateLogger) Verbose() bool {
	return l.logger.Level > log.InfoLevel
}
