package pkg

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/a-h/templ"
	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
	"github.com/stneto1/htmx-webcomponents/views"
)

type Container struct {
	conn *sqlx.DB
}

func NewContainer(conn *sqlx.DB) *Container {
	return &Container{
		conn: conn,
	}
}

func render(component templ.Component, c *fiber.Ctx) error {
	c.Response().Header.SetContentType("text/html")
	return component.Render(c.Context(), c.Response().BodyWriter())
}

func (c *Container) IndexHandler(ctx *fiber.Ctx) error {
	rows, err := c.getRecords(ctx.Context())
	if err != nil {
		log.Printf("failed to get records: %v\n", err)

		rows = &[]Record{}
	}

	root := views.RootLayout("Page Title", mapRecordsIntoView(rows))

	return render(root, ctx)
}

func (c *Container) ReseedHandler(ctx *fiber.Ctx) error {
	if err := c.reseed(ctx.Context()); err != nil {
		// do something
		log.Printf("failed to reseed: %v\n", err)
	}

	controls := views.Controls()

	return render(controls, ctx)
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

var availableColumns = []string{
	"id",
	"name",
	"value",
	"value_2",
	"value_3",
	"created_at",
}

func (c *Container) getRecords(ctx context.Context) (*[]Record, error) {
	qty := 10
	order := "id"
	direction := "DESC"

	if !slices.Contains(availableColumns, "id") {
		return nil, fmt.Errorf("invalid column")
	}

	rows := make([]Record, qty)
	builder := sqlbuilder.SQLite.
		NewSelectBuilder().
		Select("*").
		From("records").
		Limit(qty)

	if direction == "DESC" {
		builder = builder.Desc().OrderBy(order)
	} else {
		builder = builder.Asc().OrderBy(order)
	}

	sql, args := builder.Build()

	if err := c.conn.SelectContext(ctx, &rows, sql, args...); err != nil {
		log.Printf("failed to select records: %v\n", err)
		return nil, err
	}

	return &rows, nil
}

func mapRecordIntoView(r Record) views.ViewRecord {
	return views.ViewRecord{
		ID:        fmt.Sprintf("%d", r.ID),
		Name:      r.Name,
		Value:     r.Value,
		Value2:    fmt.Sprintf("%d", r.Value2),
		Value3:    r.Value3.Format(time.RFC3339),
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
	}
}

func mapRecordsIntoView(rs *[]Record) []views.ViewRecord {
	vrs := make([]views.ViewRecord, len(*rs))

	for idx, r := range *rs {
		vrs[idx] = mapRecordIntoView(r)
	}

	return vrs
}
