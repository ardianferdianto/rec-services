package console

import (
	"context"
	"errors"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/config"
	"github.com/ardianferdianto/reconciliation-service/internal/infrastructure/sqlstore"
	"github.com/ardianferdianto/reconciliation-service/pkg/db"
	"github.com/ardianferdianto/reconciliation-service/pkg/logger"
	"gopkg.in/ukautz/clif.v1"
	"log/slog"
)

type MigrateConsole struct {
	Migrator *sqlstore.Migrator
}

func NewMigrateConsole(dbConfig db.Config) *MigrateConsole {
	ctx := context.Background()
	migrator, err := sqlstore.NewMigrator(dbConfig)
	if err != nil {
		slog.ErrorContext(ctx, "Error while initiating migrator", logger.ErrAttr(err))
		return nil
	}
	return &MigrateConsole{
		Migrator: migrator,
	}
}

func (c *MigrateConsole) MigrateCreate() *clif.Command {
	return clif.NewCommand("migrate:create", "Create a new migration file.", func(o *clif.Command, in clif.Input, out clif.Output) error {
		name := o.Argument("name").String()

		if err := c.Migrator.Create(name); err != nil {
			return err
		}

		return nil
	}).NewArgument("name", "Migration name", "", true, false)
}

func (c *MigrateConsole) MigrateRun(ctx context.Context) *clif.Command {
	return clif.NewCommand("migrate:run", "Run the database migrations.", func(o *clif.Command, in clif.Input, out clif.Output) error {
		if isNotInProduction() {
			return c.Migrator.MigrateAll(ctx)
		}

		if confirmInProduction(in) {
			return c.Migrator.MigrateAll(ctx)
		}

		out.Printf("Migrate aborted.\n")

		return nil
	})
}

func (c *MigrateConsole) MigrateRollback() *clif.Command {
	return clif.NewCommand("migrate:rollback", "Rollback the last database migration.", func(o *clif.Command, in clif.Input, out clif.Output) error {
		step := o.Option("step").Int()

		if step <= 0 {
			return errors.New("step can't be zero or negative")
		}

		if isNotInProduction() {
			return c.Migrator.Rollback(step)
		}

		if confirmInProduction(in) {
			return c.Migrator.Rollback(step)
		}

		out.Printf("Rollback aborted.\n")

		return nil
	}).NewOption("step", "s", "The number of migrations to be reverted", "1", false, false)
}

func confirmInProduction(in clif.Input) bool {
	fmt.Println("**************************************")
	fmt.Println("*     Application In Production!     *")
	fmt.Println("**************************************")

	return in.Confirm("Do you really wish to run this command?")
}

func isNotInProduction() bool {
	return config.Get().App.ENV != "prod"
}
