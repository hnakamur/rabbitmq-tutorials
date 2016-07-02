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
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := "hello"
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	log.Printf(" [x] Sent %s", body)
	failOnError(err, "Failed to publish a message")
}
