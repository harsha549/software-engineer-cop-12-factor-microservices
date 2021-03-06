package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"
)

var db *sql.DB
var response top10Response

type top10Response struct {
	Boys  []top10ResponseEntry `json:"boys"`
	Girls []top10ResponseEntry `json:"girls"`
}

type top10ResponseEntry struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func main() {
	bindAddr := ":8080"
	connect := "postgres://postgres:password@localhost/babynames?sslmode=disable"
	zookeeperHost := "localhost:2181"

	if v := os.Getenv("BIND_ADDR"); len(v) > 0 {
		bindAddr = v
	}
	if v := os.Getenv("DB_CONNECT"); len(v) > 0 {
		connect = v
	}
	if v := os.Getenv("ZOOKEEPER_HOST"); len(v) > 0 {
		zookeeperHost = v
	}

	var err error
	db, err = sql.Open("postgres", connect)
	if err != nil {
		log.Fatal(err)
	}

	response, err = getResponse()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/top10", top10Handler)

	go startKafka(zookeeperHost)

	log.Printf("Listening on %s", bindAddr)
	if err := http.ListenAndServe(bindAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func startKafka(zookeeperHost string) {
	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetNewest
	config.Offsets.ProcessingTimeout = 10 * time.Second

	var zookeeperNodes []string
	zookeeperNodes, config.Zookeeper.Chroot = kazoo.ParseConnectionString(zookeeperHost)

	consumer, err := consumergroup.JoinConsumerGroup("cg2", []string{"top10changed"}, zookeeperNodes, config)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for err := range consumer.Errors() {
			log.Println("ERROR:", err.Error())
		}
	}()

	for msg := range consumer.Messages() {
		log.Println("Top 10 changed")
		response, err = getResponse()
		if err != nil {
			log.Println("ERROR:", err.Error())
		}
		consumer.CommitUpto(msg)
	}
}

func top10Handler(w http.ResponseWriter, req *http.Request) {
	b, err := json.Marshal(&response)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write(b)
}

func getResponse() (response top10Response, err error) {
	response.Boys, err = getTop10("male")
	if err != nil {
		return
	}

	response.Girls, err = getTop10("female")
	return
}

func getTop10(sex string) (entries []top10ResponseEntry, err error) {
	var rows *sql.Rows
	if rows, err = db.Query("SELECT name, total FROM baby_names WHERE sex=$1 ORDER BY total DESC LIMIT 10", sex); err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var entry top10ResponseEntry
		if err = rows.Scan(&entry.Name, &entry.Count); err != nil {
			return
		}
		entries = append(entries, entry)
	}

	err = rows.Err()
	return
}
