package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// package level constants
const (
	webPort  = "80"
	rpcPort  = "5001"
	gRpcPort = "50001"
	mongoUrl = "mongodb://mongo:27017"
)

// package level variable
var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// connect to mongodb
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	// create a context to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Close disconnect
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// Register RPC server
	err = rpc.Register(new(RPCServer))
	go app.rpcListen()

	// gRPC server
	go app.gRPCListen()

	// start web server
	// go app.serve()
	app.serve()

}

// Set up web server
func (app *Config) serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Router(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func (app *Config) rpcListen() error {
	log.Println("Start listening RPC calls on port:", rpcPort)

	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rpcConn)
	}

	// return nil
}

func connectToMongo() (*mongo.Client, error) {
	// mongo connection options
	clientOptions := options.Client().ApplyURI(mongoUrl)
	clientOptions.SetAuth(options.Credential{
		// In production, it should be defined by env or other method.
		// It is unsafe to write sensitive information in hard core.
		Username: "admin",
		Password: "password",
	})

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connect to mongo:", err)
		return nil, err
	}

	return c, nil
}
