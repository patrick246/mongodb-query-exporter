package exporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mongodb-query-exporter/aggregations"
	"path/filepath"
)

type Exporter struct {
	aggregations []prometheus.Collector
}

func NewExporter(glob string, client *mongo.Client) (*Exporter, error) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("could not glob for aggregation files. glob=%s", glob)
	}

	var ags []prometheus.Collector

	for _, match := range matches {
		ag, err := aggregations.FromFile(match, client)
		if err != nil {
			log.Printf("could not parse aggregation file. filename=%s", match)
			continue
		}
		ags = append(ags, ag)
	}
	return &Exporter{aggregations: ags}, nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, ag := range e.aggregations {
		ag.Describe(ch)
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for _, ag := range e.aggregations {
		ag.Collect(ch)
	}
}
