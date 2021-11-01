package streaming

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
	"wb-test-task/internal/db"

	stan "github.com/nats-io/stan.go"
)

type Subscriber struct {
	sub      stan.Subscription
	dbObject *db.DB
	sc       *stan.Conn
	name     string
}

func NewSubscriber(db *db.DB, conn *stan.Conn) *Subscriber {
	return &Subscriber{
		name:     "Subscriber",
		dbObject: db,
		sc:       conn,
	}
}

func (s *Subscriber) Subscribe() {
	// Simple Async Subscriber
	var err error

	ackWait, err := strconv.Atoi(os.Getenv("NATS_ACK_WAIT_SECONDS"))
	if err != nil {
		log.Printf("%s: received a message!\n", s.name)
		return
	}

	s.sub, err = (*s.sc).Subscribe(
		os.Getenv("NATS_SUBJECT"),
		func(m *stan.Msg) {
			log.Printf("%s: received a message!\n", s.name)
			if s.messageHandler(m.Data) {
				err := m.Ack() // в случае успешного сохранения msg уведомляем NATS.
				if err != nil {
					log.Printf("%s ack() err: %s", s.name, err)
				}
			}
		},
		stan.AckWait(time.Duration(ackWait)*time.Second), // Интервал тайм-аута - AckWait (30 сек default) - ожидание уведомления NATS о чтении сообщения
		//stan.DeliverAllAvailable(),                       // DeliverAllAvailable доставит все доступные сообщения
		stan.DurableName(os.Getenv("NATS_DURABLE_NAME")), // долговечные подписки позволяют клиентам назначить постоянное имя подписке
		// Это приводит к тому, что сервер потоковой передачи NATS отслеживает последнее подтвержденное сообщение для этого clientID + постоянное имя,
		// так что клиенту будут доставлены только сообщения с момента последнего подтвержденного сообщения.
		stan.SetManualAckMode(), // ручной режим подтверждения приема сообщения для подписки
		stan.MaxInflight(10))    // указывает максимальное количество ожидающих подтверждения (сообщений, которые были доставлены, но не подтверждены),
	// которые NATS Streaming разрешит для данной подписки. При достижении этого предела NATS Streaming приостанавливает доставку сообщений в эту
	// подписку до тех пор, пока количество неподтвержденных сообщений не упадет ниже указанного предела
	if err != nil {
		log.Printf("%s: error: %v\n", s.name, err)
	}
	log.Printf("%s: subscribed to subject %s\n", s.name, os.Getenv("NATS_SUBJECT"))
}

func (s *Subscriber) messageHandler(data []byte) bool {
	recievedOrder := db.Order{}
	err := json.Unmarshal(data, &recievedOrder)
	if err != nil {
		log.Printf("%s: messageHandler() error, %v\n", s.name, err)
		// ошибка формата присланных данных. Пропускаем, сообщив серверу, что сообщение получили
		return true
	}
	log.Printf("%s: unmarshal Order to struct: %v\n", s.name, recievedOrder)

	_, err = s.dbObject.AddOrder(recievedOrder)
	if err != nil {
		log.Printf("%s: unable to add order: %v\n", s.name, err)
		return false
	}
	return true
}

func (s *Subscriber) Unsubscribe() {
	if s.sub != nil {
		s.sub.Unsubscribe()
	}
}
