package streaming

import (
	"encoding/json"
	"log"
	"os"
	"wb-test-task/internal/db"

	stan "github.com/nats-io/stan.go"
)

type Subscriber struct {
	sub      stan.Subscription
	dbObject *db.DB
	sc       stan.Conn
	name     string
}

func (s *Subscriber) Subscribe() {
	// Simple Async Subscriber
	sub, err := s.sc.Subscribe(
		os.Getenv("NATS_SUBJECT"),
		func(m *stan.Msg) {
			log.Printf("%s: received a message!\n", s.name)
			s.messageHandle(m.Data)
		},
		stan.DurableName(os.Getenv("NATS_CLIENT_ID")))
	if err != nil {
		log.Printf("%s: error: %v\n", s.name, err)
	}
	log.Printf("%s: subscribed to subject %s\n", s.name, os.Getenv("NATS_SUBJECT"))
	s.sub = sub
}

func (s *Subscriber) messageHandle(data []byte) {
	recievedOrder := db.Order{}
	err := json.Unmarshal(data, &recievedOrder)
	if err != nil {
		log.Printf("%s: messageHandle() error, %v\n", s.name, err)
		return
	}
	log.Printf("%s: unmarshal Order to struct: %v\n", s.name, recievedOrder)

	_, err = s.dbObject.AddOrder(recievedOrder)
	if err != nil {
		log.Printf("%s: unable to add order: %v\n", s.name, err)
		return
	}
}

func (s *Subscriber) Unsubscribe() {
	s.sub.Unsubscribe()
}
