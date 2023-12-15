package pkg

import (
	"context"
	"log"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type Record struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Value     string    `db:"value"`
	Value2    int64     `db:"value_2"`
	Value3    time.Time `db:"value_3"`
	CreatedAt time.Time `db:"created_at"`
}

type Container struct {
	conn *sqlx.DB
}

func NewContainer(conn *sqlx.DB) *Container {
	return &Container{
		conn: conn,
	}
}

func (c *Container) ReseedHandler(ctx *fiber.Ctx) error {
	if err := c.reseed(ctx.Context()); err != nil {
		// do something
	}

	return nil
}

func (c *Container) reseed(ctx context.Context) error {
	txx, err := c.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := txx.ExecContext(ctx, "DELETE FROM records"); err != nil {
		if err := txx.Rollback(); err != nil {
			log.Printf("failed to rollback: %v\n", err)
		}

		return err
	}

	newRows := make([]Record, 1000)
	for idx := range newRows {
		if err := faker.FakeData(&newRows[idx]); err != nil {
			if err := txx.Rollback(); err != nil {
				log.Printf("failed to rollback: %v\n", err)
			}

			return err
		}
	}

	for _, row := range newRows {
		if _, err := txx.NamedExecContext(ctx, "INSERT INTO records (name, value, value_2, value_3) VALUES (:name, :value, :value_2, :value_3)", row); err != nil {
			if err := txx.Rollback(); err != nil {
				log.Printf("failed to rollback: %v\n", err)
			}

			return err
		}
	}

	if err := txx.Commit(); err != nil {
		return err
	}

	return nil
}
