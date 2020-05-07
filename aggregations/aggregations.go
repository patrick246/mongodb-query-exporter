package aggregations

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"log"
)

type aggregation struct {
	Type       string   `json:"type"`
	Name       string   `json:"name"`
	Help       string   `json:"help"`
	Labels     []string `json:"labels"`
	Database   string   `json:"database"`
	Collection string   `json:"collection"`
	Pipeline   bson.A   `json:"pipeline"`
}

var ExecutionHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "aggregation_execution_seconds",
	Help:    "Execution time for the aggregations",
	Buckets: prometheus.ExponentialBuckets(1e-6, 2, 25),
}, []string{"name"})

func FromFile(filename string, client *mongo.Client) (prometheus.Collector, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("Error reading file")
		return nil, err
	}

	var data aggregation

	err = bson.UnmarshalExtJSON(content, true, &data)
	if err != nil {
		log.Println("Error unmarshalling ext json")
		return nil, err
	}

	switch data.Type {
	case "gauge":
		return newGaugeAggregation(data, client), nil
	}
	return nil, fmt.Errorf("could not create metric by type: type=%s", data.Type)
}
