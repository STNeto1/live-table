package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/go-faker/faker/v4"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
	"github.com/stneto1/htmx-webcomponents/views"
	"golang.org/x/exp/constraints"
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
	rows, pages, err := c.getRecords(ctx.Context(), TableState{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		log.Printf("failed to get records: %v\n", err)

		rows = &[]Record{}
	}

	root := views.RootLayout("Page Title", mapRecordsIntoView(rows), "1", numberToStr(pages), "10")

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

func (c *Container) RecordsWsHandler(ws *websocket.Conn) {
	defer func() {
		unregister <- ws
		ws.Close()
	}()

	register <- ws

	for {
		messageType, msg, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}

			break
		}

		if messageType == websocket.TextMessage {
			var payload EventBody
			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Println("error unmarshalling payload:", err)
				continue
			}

			if payload.Event == "reseed" {
				if err := c.reseed(context.Background()); err != nil {
					log.Printf("failed to reseed: %v\n", err)
					continue
				}

				c.updateConnectionTableView(ws, states[ws])
			}

			if payload.Event == "change_page" {
				parsed, err := strconv.ParseInt(payload.ToPage, 10, 64)
				if err != nil {
					log.Println("error parsing page:", err)
					continue
				}

				states[ws].Page = int(parsed)

				c.updateConnectionTableView(ws, states[ws])

			}

			if payload.Event == "change_page_size" {
				parsed, err := strconv.ParseInt(payload.PageSize, 10, 64)
				if err != nil {
					log.Println("error parsing page size:", err)
					continue
				}

				states[ws].PageSize = int(parsed)

				c.updateConnectionTableView(ws, states[ws])
			}

		} else {
			log.Println("websocket message received of type", messageType)
		}
	}
}

func (c *Container) updateConnectionTableView(ws *websocket.Conn, state *TableState) {
	rows, pages, err := c.getRecords(context.Background(), *state)
	if err != nil {
		log.Printf("failed to get records: %v\n", err)
		return
	}

	component := views.RecordTable(mapRecordsIntoView(rows), numberToStr(state.Page), numberToStr(pages), numberToStr(state.PageSize))

	htmlWriter := &bytes.Buffer{}
	if err := component.Render(context.Background(), htmlWriter); err != nil {
		log.Println("error rendering component:", err)
		return
	}

	broadcast <- string(htmlWriter.Bytes())
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

func (c *Container) getRecords(ctx context.Context, state TableState) (*[]Record, int, error) {
	order := "id"
	direction := "DESC"
	offset := (state.Page - 1) * state.PageSize

	if !slices.Contains(availableColumns, "id") {
		order = "id"
	}

	rows := make([]Record, state.PageSize)
	builder := sqlbuilder.SQLite.
		NewSelectBuilder().
		Select("*").
		From("records").
		Limit(int(state.PageSize)).
		Offset(offset)

	if direction == "DESC" {
		builder = builder.Desc().OrderBy(order)
	} else {
		builder = builder.Asc().OrderBy(order)
	}

	sql, args := builder.Build()

	if err := c.conn.SelectContext(ctx, &rows, sql, args...); err != nil {
		log.Printf("failed to select records: %v\n", err)
		return nil, 0, err
	}

	var total int
	if err := c.conn.GetContext(ctx, &total, "SELECT COUNT(*) FROM records"); err != nil {
		log.Printf("failed to get total: %v\n", err)
		return nil, 0, err
	}

	return &rows, total / state.PageSize, nil
}

func mapRecordIntoView(r Record) views.ViewRecord {
	return views.ViewRecord{
		ID:        numberToStr(r.ID),
		Name:      r.Name,
		Value:     r.Value,
		Value2:    numberToStr(r.Value2),
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

type Number interface {
	constraints.Integer
}

// source -> https://stackoverflow.com/a/70370013
func numberToStr[T Number](n T) string {
	return fmt.Sprintf("%d", n)
}
