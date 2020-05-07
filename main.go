package main

import (
	"context"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"mongodb-query-exporter/aggregations"
	exporter2 "mongodb-query-exporter/exporter"
	"net/http"
	"time"
)

var addr = flag.String("listen-address", ":9736", "The address to listen for HTTP requests")
var aggregationPath = flag.String("config-path", "./data/*.json", "The glob pattern to the metric config files")
var mongoDbUri = flag.String("mongodb-uri", "mongodb://localhost:27017", "The connection string to connect to the mongodb")

func main() {
	flag.Parse()
	http.Handle("/metrics", promhttp.Handler())

	client := connectToMongoDb()

	exporter, err := exporter2.NewExporter(*aggregationPath, client)

	if err != nil {
		log.Println(err)
	}
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(aggregations.ExecutionHistogram)

	log.Println("MongoDB query exporter starting on " + *addr + ", exporting metrics on /metrics")
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func connectToMongoDb() *mongo.Client {
	log.Printf("Connection to MongoDB")
	client, err := mongo.NewClient(options.Client().ApplyURI(*mongoDbUri))
	if err != nil {
		log.Fatalf("could not create mongodb client. reason=%v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("could not connect to mongodb. reason=%v", err)
	}
	log.Printf("Connected to MongoDB")
	return client
}
