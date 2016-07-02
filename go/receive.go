package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	var caDir string
	flag.StringVar(&caDir, "ca-dir", "./testca", "CA directory")
	var clientCertDir string
	flag.StringVar(&clientCertDir, "client-cert-dir", "./client", "client certificate directory")
	var clientKeyDir string
	flag.StringVar(&clientKeyDir, "client-key-dir", "./client", "client certificate key directory")
	flag.Parse()

	cfg := new(tls.Config)
	cfg.RootCAs = x509.NewCertPool()

	cacertPath := filepath.Join(caDir, "cacert.pem")
	if ca, err := ioutil.ReadFile(cacertPath); err == nil {
		cfg.RootCAs.AppendCertsFromPEM(ca)
	}
	clientCertPath := filepath.Join(clientCertDir, "cert.pem")
	clientKeyPath := filepath.Join(clientKeyDir, "key.pem")
	if cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath); err == nil {
		cfg.Certificates = append(cfg.Certificates, cert)
	}

	conn, err := amqp.DialTLS("amqps://guest:guest@localhost:5671/", cfg)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when usused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
