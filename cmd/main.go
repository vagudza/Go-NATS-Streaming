package main

import (
	"fmt"
	"wb-test-task/api"
	"wb-test-task/cmd/config"
	"wb-test-task/internal/db"
	"wb-test-task/internal/streaming"
)

func main() {
	// Инициализация конфигурации проекта
	config.ConfigSetup()

	// Инициализация подключения к БД
	var dbObject db.DB = db.DB{}
	dbObject.Init()

	// Инициализация кеша
	var csh db.Cache = db.Cache{}
	csh.Init(&dbObject)
	defer csh.Finish()

	// Инициализация NATS Streaming: Publisher + Subscriber
	var sh streaming.StreamingHandler = streaming.StreamingHandler{}
	sh.Init(&dbObject)
	defer sh.Finish()

	// Запуск сервера для выдачи OrderOut по адресу http://localhost:3333/orders/123
	var myApi api.Api = api.Api{}
	myApi.Init(&csh)
	defer myApi.Finish()

	fmt.Println("====Нажмите Enter для выхода из программы===")
	fmt.Scanln()
}
