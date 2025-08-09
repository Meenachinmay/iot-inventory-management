package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// Read CA certificate
	caCert, err := os.ReadFile("./internal/config/certs/emqx-ca.crt")
	if err != nil {
		log.Fatal("Failed to read CA cert:", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker("ssl://q15c1f50.ala.asia-southeast1.emqxsl.com:8883")
	opts.SetClientID("iot-backend")
	opts.SetUsername("iot_backend") // Replace with your EMQX username
	opts.SetPassword("Chinmay1234") // Replace with your EMQX password
	opts.SetTLSConfig(tlsConfig)

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Connection failed:", token.Error())
	}

	fmt.Println("Successfully connected to EMQX Cloud!")

	// Test publish
	token := client.Publish("test/topic", 0, false, "Hello EMQX!")
	token.Wait()

	time.Sleep(2 * time.Second)
	client.Disconnect(250)
}
