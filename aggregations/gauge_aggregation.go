package aggregations

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type gaugeAggregation struct {
	gauge      *prometheus.GaugeVec
	name       string
	client     *mongo.Client
	database   string
	collection string
	pipeline   bson.A
}

type gaugeAggregationResult struct {
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
}

func newGaugeAggregation(data aggregation, client *mongo.Client) *gaugeAggregation {
	return &gaugeAggregation{
		gauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: data.Name,
			Help: data.Help,
		}, data.Labels),
		name:       data.Name,
		client:     client,
		database:   data.Database,
		collection: data.Collection,
		pipeline:   data.Pipeline,
	}
}

func (a *gaugeAggregation) Describe(ch chan<- *prometheus.Desc) {
	a.gauge.Describe(ch)
}

func (a *gaugeAggregation) Collect(ch chan<- prometheus.Metric) {
	up := a.scrape()
	ch <- prometheus.MustNewConstMetric(prometheus.NewDesc("up", "MongoDB Query Exporter scrape result", nil, nil), prometheus.GaugeValue, up)
	a.gauge.Collect(ch)
}

func (a *gaugeAggregation) scrape() (up float64) {
	start := time.Now()
	defer ExecutionHistogram.WithLabelValues(a.name).Observe(time.Since(start).Seconds())

	aggregationComment := "Prometheus MongoDB Query Exporter. github.com/patrick246/mongodb-query-exporter"
	it, err := a.client.Database(a.database).Collection(a.collection).Aggregate(context.Background(), a.pipeline, &options.AggregateOptions{Comment: &aggregationComment})
	if err != nil {
		log.Printf("Error running aggregation pipeline: metric=%s database=%s collection=%s error=%v", a.name, a.database, a.collection, err)
		return 0
	}
	defer it.Close(context.Background())

	for it.Next(context.Background()) {
		var result gaugeAggregationResult
		err = it.Decode(&result)
		if err != nil {
			log.Printf("Error getting aggregation pipeline result: metric=%s database=%s collection=%s error=%v", a.name, a.database, a.collection, err)
			continue
		}
		a.gauge.With(result.Labels).Set(result.Value)
	}

	return 1
}
