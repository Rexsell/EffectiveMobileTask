package main

import (
	"EffectiveMobileTask/internal/database"
	"EffectiveMobileTask/internal/server"
	"context"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	SourceDatabaseDsn := os.Getenv("DB_DSN")
	SourceDatabaseName := os.Getenv("DB_NAME")
	SourceCollectionName := os.Getenv("DB_TABLE")

	dbPort := os.Getenv("SERVER_PORT")

	db := database.New(SourceDatabaseDsn, SourceCollectionName, SourceDatabaseName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		log.Fatal("connection failed: ", err)
		return
	}
	defer func() {
		if err := db.Close(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	if err := db.MakeMigration(); err != nil {
		log.Fatal("migration failed: ", err)
		return
	}
	srv := server.New(dbPort, db)
	log.Printf("Server started at port %s", dbPort)
	if err := srv.StartServer(); err != nil {
		log.Fatal(err)
		return
	}

}
