package database

import (
	"context"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
)

type DB struct {
	DSN            string
	CollectionName string
	DatabaseName   string
	Conn           *pgx.Conn
}

func New(dsn, collectionName, databaseName string) *DB {
	return &DB{
		DSN:            dsn,
		CollectionName: collectionName,
		DatabaseName:   databaseName,
	}
}

func (db *DB) MakeMigration() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")

	m, err := migrate.New(
		migrationsDir,
		mergeUrlParams(db.DSN, db.DatabaseName, "disable"),
	)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		return err
	}
	return err
}

func (db *DB) Connect(ctx context.Context) error {
	dbConn, err := pgx.Connect(ctx, db.DSN)
	if err != nil {
		return err
	}
	db.Conn = dbConn
	return nil
}

func (db *DB) Close(ctx context.Context) error {
	return db.Conn.Close(ctx)
}

func mergeUrlParams(dsn string, tableName string, sslmode string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		log.Fatal(err)
	}

	u.Path = "/" + tableName
	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()

	return u.String()
}
