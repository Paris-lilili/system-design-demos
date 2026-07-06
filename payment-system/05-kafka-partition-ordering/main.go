package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

type Event struct {
	OrderId string `json:"order_id"`
	Seq     int    `json:"seq"`  // event sequence
	Type    string `json:"type"` // event type
}

func main() {
	// 1. connect to kafka
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// controller is kafka.Broker
	controller, err := conn.Controller()
	if err != nil {
		panic(err)
	}

	// dail again to connect address according to host and post
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		panic(err)
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             "payment-events",
		NumPartitions:     3,
		ReplicationFactor: 1, // single broker only 1
	})
	if err != nil {
		panic(err)
	}

	// 2. create writer to send event to kafka
	writer := &kafka.Writer{
		Addr: kafka.TCP("localhost:9092"),
		Topic: "payment-events",
		Balancer: &kafka.Hash{},
	} 

	// total two orders, 5 events per order
	ctx := context.Background()
	orders := []string{"order_A", "order_B"}
	for _, orderID := range orders {
		for seq := 1; seq <= 5; seq++ {
			event := Event {OrderId: orderID, Seq: seq, Type: "processing" }
			value, err := json.Marshal(event) // struct -> JSON byte
			if err != nil {
				panic(err)
			}

			err = writer.WriteMessages(ctx, kafka.Message{
				Key: []byte(orderID), // partition key
				Value: value,
			})
			if err != nil {
				panic(err)
			}
		}
	}
	fmt.Println("produced 10 msgs")

	// close (and flush) the writer now, before consuming, to make sure all produced messages are actually sent to Kafka before we start reading.
	writer.Close()

	// 3. create a reader to comsume all msg from order_A and order_B. (set GroupID rather than set single partition)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic: "payment-events",
		GroupID: "demo-consumer",
	})
	defer reader.Close()

	// kafka is a continuous stream, so we consume one message at a time
	for i := 0; i < 10; i++ {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("topic= %s, partition= %d, key= %s, value= %s\n", string(msg.Topic), msg.Partition, string(msg.Key), string(msg.Value))
	}
}
