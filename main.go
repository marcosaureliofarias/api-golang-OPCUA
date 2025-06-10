// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"os"
// 	"time"

// 	mqtt "github.com/eclipse/paho.mqtt.golang"
// 	"github.com/gopcua/opcua"
// 	"github.com/gopcua/opcua/ua"
// )

// type OpcuaData struct {
// 	NodeID string      `json:"node_id"`
// 	Value  interface{} `json:"value"`
// 	Time   time.Time   `json:"timestamp"`
// }

// var mqttClient mqtt.Client

// func main() {
// 	// Conexão MQTT
// 	opts := mqtt.NewClientOptions().
// 		AddBroker("tcp://broker.hivemq.com:1883").
// 		SetClientID("opcua-client-test")
// 	mqttClient = mqtt.NewClient(opts)
// 	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("Erro ao conectar ao broker MQTT: %v", token.Error())
// 	}
// 	fmt.Println("✅ Conectado ao broker MQTT")

// 	// Conexão OPC-UA
// 	endpoint := "opc.tcp://0.0.0.0:4840"
// 	if len(os.Args) > 1 {
// 		endpoint = os.Args[1]
// 	}

// 	ctx := context.Background()

// 	client, err := opcua.NewClient(endpoint,
// 		opcua.SecurityMode(ua.MessageSecurityModeNone),
// 		opcua.AuthAnonymous(),
// 	)
// 	if err != nil {
// 		log.Fatalf("Erro ao criar o cliente OPC-UA: %v", err)
// 	}

// 	if err := client.Connect(ctx); err != nil {
// 		log.Fatalf("Erro ao conectar com o servidor OPC-UA: %v", err)
// 	}
// 	defer client.Close(ctx)

// 	fmt.Println("✅ Conectado ao servidor OPC-UA:", endpoint)

// 	nodeIDs := []string{
// 		"ns=2;s=Demo.Static.Scalar.Double",
// 		"ns=2;s=Demo.Static.Scalar.Boolean",
// 		"ns=2;s=Demo.Static.Scalar.Int32",
// 	}

// 	for {
// 		for _, nodeID := range nodeIDs {
// 			readAndPublish(client, nodeID)
// 		}
// 		time.Sleep(2 * time.Second)
// 	}
// }

// func readAndPublish(client *opcua.Client, nodeID string) {
// 	id, err := ua.ParseNodeID(nodeID)
// 	if err != nil {
// 		log.Printf("❌ Erro ao parsear NodeID (%s): %v", nodeID, err)
// 		return
// 	}

// 	req := &ua.ReadRequest{
// 		NodesToRead: []*ua.ReadValueID{
// 			{NodeID: id, AttributeID: ua.AttributeIDValue},
// 		},
// 		TimestampsToReturn: ua.TimestampsToReturnBoth,
// 	}

// 	resp, err := client.Read(context.Background(), req)
// 	if err != nil || resp.Results[0].Status != ua.StatusOK {
// 		log.Printf("⚠️ Erro ao ler valor de %s: %v", nodeID, err)
// 		return
// 	}

// 	value := resp.Results[0].Value.Value()
// 	timestamp := time.Now()

// 	data := OpcuaData{
// 		NodeID: nodeID,
// 		Value:  value,
// 		Time:   timestamp,
// 	}

// 	payload, err := json.Marshal(data)
// 	if err != nil {
// 		log.Printf("❌ Erro ao serializar JSON: %v", err)
// 		return
// 	}

// 	// Publica no tópico MQTT
// 	topic := "teste/opcua"
// 	token := mqttClient.Publish(topic, 0, false, payload)
// 	token.Wait()

// 	fmt.Printf("📤 Publicado em %s: %s\n", topic, payload)
// }

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type OpcuaData struct {
	NodeID    string      `json:"node_id"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

var mqttClient mqtt.Client

func main() {
	// Configura o cliente MQTT
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("go-subscriber")
	mqttClient = mqtt.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("❌ Erro ao conectar ao broker MQTT: %v", token.Error())
	}
	fmt.Println("✅ Conectado ao broker MQTT")

	// Conexão com o servidor OPC UA
	endpoint := "opc.tcp://DESKTOP-PO6FJ9U:4334/UA/SecadorSimulado"
	if len(os.Args) > 1 {
		endpoint = os.Args[1]
	}

	ctx := context.Background()
	client, err := opcua.NewClient(endpoint,
		opcua.SecurityMode(ua.MessageSecurityModeNone),
		opcua.AuthAnonymous(),
	)
	if err != nil {
		log.Fatalf("❌ Erro ao criar cliente OPC UA: %v", err)
	}
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("❌ Erro ao conectar ao servidor OPC UA: %v", err)
	}
	defer client.Close(ctx)
	fmt.Println("✅ Conectado ao servidor OPC UA:", endpoint)

	// NodeIDs simulados
	nodeIDs := []string{
		"ns=1;s=Temperatura",
		"ns=1;s=Umidade",
		"ns=1;s=Status",
	}

	for {
		for _, nodeID := range nodeIDs {
			readAndPublish(client, nodeID)
		}
		time.Sleep(2 * time.Second)
	}
}

func readAndPublish(client *opcua.Client, nodeID string) {
	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		log.Printf("❌ Erro ao parsear NodeID %s: %v", nodeID, err)
		return
	}

	req := &ua.ReadRequest{
		NodesToRead: []*ua.ReadValueID{
			{NodeID: id, AttributeID: ua.AttributeIDValue},
		},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	resp, err := client.Read(context.Background(), req)
	if err != nil || resp.Results[0].Status != ua.StatusOK {
		log.Printf("⚠️ Erro ao ler valor de %s: %v", nodeID, err)
		return
	}

	val := resp.Results[0]
	data := OpcuaData{
		NodeID:    nodeID,
		Value:     val.Value.Value(),
		Timestamp: val.ServerTimestamp, // ou val.SourceTimestamp
	}

	jsonPayload, err := json.Marshal(data)
	if err != nil {
		log.Printf("❌ Erro ao converter para JSON: %v", err)
		return
	}

	topic := "secador/opcua"
	token := mqttClient.Publish(topic, 0, false, jsonPayload)
	token.Wait()
	fmt.Printf("📤 Publicado em %s: %s\n", topic, jsonPayload)
}
