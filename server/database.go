package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var pool *pgxpool.Pool

func connect() {
	_pool, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err)
	}
	pool = _pool
}

func timeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(time.Second*10))
}

func makeWhereInVarsForArray(size int, offset int) string {
	str := ""
	for i := 0; i < size; i++ {
		str += fmt.Sprintf("$%d", i+offset+1)
		if i != size-1 {
			str += ","
		}
	}
	return str
}

func logNewReading(reading *Reading) error {
	ctx, cancel := timeoutContext()
	defer cancel()
	err := pool.QueryRow(ctx, "INSERT INTO readings (timestamp, uptime, meat_temp, smoke_temp) VALUES ($1,$2,$3,$4) RETURNING \"id\"",
		reading.Timestamp,
		reading.Uptime,
		reading.MeatTemp,
		reading.SmokeTemp,
	).Scan(&reading.Id)
	return err
}

func createNewCook(cook *Cook) error {
	ctx, cancel := timeoutContext()
	defer cancel()
	err := pool.QueryRow(ctx, "INSERT INTO cooks (description, created) VALUES ($1, $2) RETURNING \"id\"", cook.Description, cook.Created).Scan(&cook.Id)
	if err != nil {
		return err
	}
	return nil
}

func updateCook(cook *Cook) error {
	ctx, cancel := timeoutContext()
	defer cancel()
	_, err := pool.Exec(ctx, "UPDATE cooks SET description = $1 WHERE id = $2", cook.Description, cook.Id)
	if err != nil {
		return err
	}
	return nil
}

func addReadingsToCook(ids []int, cookId int) error {
	ctx, cancel := timeoutContext()
	defer cancel()
	vars := makeWhereInVarsForArray(len(ids), 1)
	params := make([]interface{}, 1+len(ids))
	params[0] = cookId
	for i, id := range ids {
		params[i+1] = id
	}
	_, err := pool.Exec(ctx, "UPDATE readings SET cook_id = $1 WHERE id IN ("+vars+")", params...)
	return err
}

func removeReadingsFromCook(ids []int) error {
	ctx, cancel := timeoutContext()
	defer cancel()
	vars := makeWhereInVarsForArray(len(ids), 0)
	params := make([]interface{}, len(ids))
	for i, id := range ids {
		params[i] = id
	}
	_, err := pool.Exec(ctx, "UPDATE readings SET cook_id = NULL WHERE id IN ("+vars+")", params...)
	return err
}

func getOrphanReadingsInRange(limit int, offset int) ([]Reading, error) {
	ctx, cancel := timeoutContext()
	defer cancel()
	rows, err := pool.Query(ctx, "SELECT id, timestamp, uptime, meat_temp, smoke_temp FROM readings WHERE cook_id IS NULL ORDER BY timestamp DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return make([]Reading, 0), err
	}
	defer rows.Close()
	return processReadingsRow(rows), nil
}

func getCooksInRange(limit int, offset int) ([]Cook, error) {
	ctx, cancel := timeoutContext()
	defer cancel()
	rows, err := pool.Query(ctx, "SELECT id, created, description FROM cooks ORDER BY timestamp DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return make([]Cook, 0), err
	}
	defer rows.Close()
	cooks := make([]Cook, 0)
	for rows.Next() {
		cook := Cook{}
		rows.Scan(
			&cook.Id,
			&cook.Created,
			&cook.Description,
		)
		cooks = append(cooks, cook)
	}
	return cooks, nil
}

func getCookById(id int) (Cook, error) {
	ctx, cancel := timeoutContext()
	defer cancel()
	cook := Cook{
		Id: id,
	}
	err := pool.QueryRow(ctx, "SELECT description, created FROM cooks WHERE id = $1", id).Scan(&cook.Description, &cook.Created)
	if err != nil {
		return Cook{}, err
	}
	rows, err := pool.Query(ctx, "SELECT id, timestamp, uptime, meat_temp, smoke_temp FROM readings WHERE cook_id = $1", id)
	if err != nil {
		return Cook{}, err
	}
	defer rows.Close()
	cook.Readings = processReadingsRow(rows)
	return cook, nil
}

func processReadingsRow(rows pgx.Rows) []Reading {
	readings := make([]Reading, 0)
	for rows.Next() {
		reading := Reading{}
		rows.Scan(
			&reading.Id,
			&reading.Timestamp,
			&reading.Uptime,
			&reading.MeatTemp,
			&reading.SmokeTemp,
		)
		readings = append(readings, reading)
	}
	return readings
}
