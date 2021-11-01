package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Инициализация пула соединений
func (db *DB) Init() {
	db.name = "Postgres"
	var err error
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s/%s", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

	// создаем конфиг
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		log.Fatalf("%v: Init() error: %s\n", db.name, err)
	}

	// настраиваем конфиг
	/*
		maxConn, err := strconv.Atoi(os.Getenv("DB_POOL_MAXCONN"))
		if err != nil {
			log.Fatalf("%v: Init() error: %s\n", db.name, err)
		}

		maxConnLifeTime, err := strconv.Atoi(os.Getenv("DB_POOL_MAXCONN_LIFETIME"))
		if err != nil {
			log.Fatalf("%v: Init() error: %s\n", db.name, err)
		}
	*/
	//config.MaxConns = int32(maxConn)
	//config.MaxConnLifetime = time.Duration(maxConnLifeTime) * time.Second

	db.pool, err = pgxpool.ConnectConfig(context.Background(), config)
	//db.pool, err = pgxpool.Connect(context.Background(), dbUrl)
	if err != nil {
		//db.pool.Close()
		log.Fatalf("%v: unable to connect to database: %v\n", db.name, err)
	}
	log.Printf("%v: connected to database\n", db.name)
}
