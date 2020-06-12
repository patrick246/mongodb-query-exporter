package main

import (
	"context"
	"github.com/namsral/flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"mongodb-query-exporter/aggregations"
	exporter2 "mongodb-query-exporter/exporter"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	signalListener(exporter)

	log.Println("MongoDB query exporter starting on " + *addr + ", exporting metrics on /metrics")
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func connectToMongoDb() *mongo.Client {
	log.Printf("Connecting to MongoDB")
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
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("Tried to establish connection to mongodb, but failed: reason=%v", err)
	}
	log.Printf("Connected to MongoDB")
	return client
}

func signalListener(exporter *exporter2.Exporter) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGUSR1)

	go func() {
		for {
			<-signals
			log.Printf("Reloading config")
			exporter.Reload()
		}
	}()
}
