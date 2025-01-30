package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/pkg/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"os"
	"time"
)

const migrationDir = "internal/infrastructure/sqlstore/migrations"

type Migrator struct {
	migrator *migrate.Migrate
}

func openDBWrapper(dbConfig db.Config) (database.Driver, error) {
	masterDsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)
	sqlDB, err := sql.Open(dbConfig.Driver, masterDsn)
	if err != nil {
		fmt.Printf("error while opening sql connection: %s\n", err.Error())
		return nil, err
	}

	err = sqlDB.Ping()
	if err != nil {
		fmt.Printf("error while pinging sql connection: %s\n", err.Error())
		return nil, err
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		fmt.Printf("error while initiating postgres driver: %s\n", err.Error())
		return nil, err
	}

	return driver, nil
}

func NewMigrator(dbConfig db.Config) (*Migrator, error) {
	driver, err := openDBWrapper(dbConfig)
	if err != nil {
		return nil, err
	}

	migrator, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", migrationDir), "postgres", driver)
	if err != nil {
		return nil, err
	}

	return &Migrator{
		migrator: migrator,
	}, err
}

// Migrate executes all migrations exists
func (m *Migrator) Migrate(ctx context.Context) error {
	if err := m.migrator.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}

		fmt.Println("No migration performed, current schema is the latest version.")
	}
	return nil
}

// MigrateAll Run migration
func (m *Migrator) MigrateAll(ctx context.Context) error {
	err := m.Migrate(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Migrate run successfully.")

	return nil
}

// Rollback migration(s) that start from the last till N step backward.
func (m *Migrator) Rollback(step int) error {
	err := m.migrator.Steps(step * -1)
	if err != nil {
		fmt.Println("Failed to rollback.")
		return err
	}

	fmt.Println("Rollback successfully.")

	return nil
}

func (m *Migrator) Create(name string) (err error) {
	now := time.Now()
	unixTime := now.Unix()

	fileNameUp := fmt.Sprintf("%s/%d_%s.up.sql", migrationDir, unixTime, name)
	fileNameDown := fmt.Sprintf("%s/%d_%s.down.sql", migrationDir, unixTime, name)

	if _, err := os.Create(fileNameUp); err != nil {
		return err
	}

	if _, err := os.Create(fileNameDown); err != nil {
		_ = os.Remove(fileNameUp)
		return err
	}

	fmt.Printf("New migration files created successfully.")

	return nil
}
