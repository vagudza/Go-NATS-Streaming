package api

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
	"wb-test-task/internal/db"

	"github.com/go-chi/chi/v5"
)

type ordkey string

const orderKey ordkey = "order"

type Api struct {
	rtr                *chi.Mux
	csh                *db.Cache
	name               string
	srv                *http.Server
	httpServerExitDone *sync.WaitGroup
}

func NewApi(csh *db.Cache) *Api {
	api := Api{}
	api.Init(csh)
	return &api
}

// Инициализация сервера
func (a *Api) Init(csh *db.Cache) {
	a.csh = csh
	a.name = "API"
	a.rtr = chi.NewRouter()
	a.rtr.Get("/", a.WellcomeHandler)

	// RESTy routes https://github.com/go-chi/chi
	a.rtr.Route("/orders", func(r chi.Router) {
		r.Route("/{orderID}", func(r chi.Router) {
			r.Use(a.orderCtx)
			r.Get("/", a.GetOrder) // GET /orders/123
		})
	})

	a.httpServerExitDone = &sync.WaitGroup{}
	a.httpServerExitDone.Add(1)
	a.StartServer()
}

// Корректное завершение работы сервера
func (a *Api) Finish() {
	log.Printf("%v: Выключение сервера...\n", a.name)

	// now close the server gracefully ("shutdown")
	if err := a.srv.Shutdown(context.Background()); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}

	// wait for goroutine started in startHttpServer() to stop
	a.httpServerExitDone.Wait()
	log.Printf("%v: Сервер успешно выключен!\n", a.name)
}

// Запуск сервера в отдельном потоке (для корректного завершения работы программы: очистка кеша из БД, отключение от подписки)
func (a *Api) StartServer() {
	a.srv = &http.Server{
		Addr:    ":3333",
		Handler: a.rtr,
	}

	go func() {
		defer a.httpServerExitDone.Done() // let main know we are done cleaning up

		log.Printf("%v: сервер будет запущен по адресу http://localhost:3333\n", a.name)
		// always returns error. ErrServerClosed on graceful close
		if err := a.srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexpected error. port in use?
			log.Printf("ListenAndServe() error: %v", err)
			return
		}
	}()
}

// Мидлвара, сохраняющая в контекст Order
func (a *Api) orderCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderIDstr := chi.URLParam(r, "orderID")
		orderID, err := strconv.ParseInt(orderIDstr, 10, 64)
		if err != nil {
			log.Printf("%v: ошибка конвертации %s в число: %v\n", a.name, orderIDstr, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		log.Printf("%v: запрос OrderOut из кеша/бд, OrderID: %v\n", a.name, orderIDstr)
		orderOut, err := a.csh.GetOrderOutById(orderID)
		if err != nil {
			log.Printf("%v: ошибка получения OrderOut из базы данных: %v\n", a.name, err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound) // 404
			return
		}
		ctx := context.WithValue(r.Context(), orderKey, orderOut)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Обработчик главной страницы http://localhost:3333
func (a *Api) WellcomeHandler(w http.ResponseWriter, r *http.Request) {
	// Установка ответа для браузера, что страница загрузилась
	t, err := template.ParseFiles("ui/templates/order.html")
	if err != nil {
		log.Printf("%v: getOrder(): ошибка парсинга шаблона html: %s\n", a.name, err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = t.ExecuteTemplate(w, "order.html", nil)
	if err != nil {
		log.Printf("%v: WellcomeHandler(): ошибка выполнения шаблона html: %s\n", a.name, err)
		return
	}
}

// Хендлер запроса Order
func (a *Api) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orderOut, ok := ctx.Value(orderKey).(*db.OrderOut)
	if !ok {
		log.Printf("%v: getOrder(): ошибка приведения интерфейса к типу *OrderOut\n", a.name)
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity) // 422
		return
	}

	// Установка ответа для браузера, что страница загрузилась
	t, err := template.ParseFiles("ui/templates/order.html")
	if err != nil {
		log.Printf("%v: getOrder(): ошибка парсинга шаблона html: %s\n", a.name, err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	t.ExecuteTemplate(w, "order.html", orderOut)
	if err != nil {
		log.Printf("%v: GetOrder(): ошибка выполнения шаблона html: %s\n", a.name, err)
		return
	}
}
