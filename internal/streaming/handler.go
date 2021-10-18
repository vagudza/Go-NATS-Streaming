package streaming

import (
	"log"
	"os"
	"time"
	"wb-test-task/internal/db"

	"github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

type StreamingHandler struct {
	conn  stan.Conn
	sub   Subscriber
	pub   Publisher
	name  string
	isErr bool
}

// Инициализация Subscriber и Publisher
func (sh *StreamingHandler) Init(dbObject *db.DB) {
	sh.name = "StreamingHandler"
	err := sh.Connect()

	if err == nil {
		sh.sub = Subscriber{dbObject: dbObject, sc: sh.conn, name: "Subscriber"}
		sh.sub.Subscribe()

		sh.pub = Publisher{sc: sh.conn, name: "Publisher"}
		sh.pub.Publish()
	} else {
		sh.isErr = true
	}
}

// Подключение к NATS
func (sh *StreamingHandler) Connect() error {
	conn, err := stan.Connect(
		os.Getenv("NATS_CLUSTER_ID"),
		os.Getenv("NATS_CLIENT_ID"),
		stan.NatsURL(os.Getenv("NATS_HOSTS")),
		stan.NatsOptions(
			nats.ReconnectWait(time.Second*4),
			nats.Timeout(time.Second*4),
		),
		//stan.Pings(5, 3), // Send PINGs every 5 seconds, and fail after 3 PINGs without any response.
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("%s: connection lost, reason: %v", sh.name, reason)
		}),
	)
	if err != nil {
		log.Printf("%s: can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", sh.name, err, os.Getenv("NATS_HOSTS"))
		return err
	}
	sh.conn = conn
	log.Printf("%s: connected to %s clusterID: [%s] clientID: [%s]\n", sh.name, os.Getenv("NATS_HOSTS"), os.Getenv("NATS_CLUSTER_ID"),
		os.Getenv("NATS_CLIENT_ID"))
	return nil
}

// Завершение работы с NATS
func (sh *StreamingHandler) Finish() {
	if !sh.isErr {
		log.Printf("%s: Finish...", sh.name)
		sh.sub.Unsubscribe()
		sh.conn.Close()
		log.Printf("%s: Finished!", sh.name)
	}
}
