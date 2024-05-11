package main

import (
	excel_file_reader "app/services/excel-file-reader"
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var forever chan int

	conn, err := amqp.Dial("amqp://user:password@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_read_excel_file", // name
		false,                 // durable
		true,                  // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for d := range msgs {
			response := map[string]any{}

			err := json.Unmarshal(d.Body, &response)

			fileName := response["file_name"].(string)
			path, _ := filepath.Abs("./storage/app/public/excels")

			studySessionId := int(response["study_session_id"].(float64))
			scheduleFileReader := excel_file_reader.NewScheduleExcelFileReader(path+"/"+fileName /*"/home/hai/Downloads/MHT_HK 2_NH2023_2024_Lienthong.xlsx"*/, studySessionId)
			data := scheduleFileReader.HandleData()
			dataJson, _ := json.Marshal(data)

			err = ch.PublishWithContext(ctx,
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          dataJson,
				})
			failOnError(err, "Failed to publish a message")

			err = d.Ack(false)
			if err != nil {
				failOnError(err, "")
				return
			}
		}
	}()

	log.Printf(" [*] Awaiting RPC requests")

	<-forever
}

func WriteFile(string2 string) {
	f, err := os.Create("test.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	l, err := f.WriteString(string2)
	if err != nil {
		fmt.Println(err)
		err := f.Close()
		if err != nil {
			return
		}
		return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
