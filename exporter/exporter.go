package exporter

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/remeh/sizedwaitgroup"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mongodb-query-exporter/aggregations"
	"path/filepath"
)

type Exporter struct {
	aggregations []prometheus.Collector
	operations   chan aggregationOperation
}

type aggregationOperationType uint64

const (
	get aggregationOperationType = iota
	reload
)

type aggregationOperation struct {
	operationType aggregationOperationType
	result        chan prometheus.Collector
}

var maxConcurrentQueries = flag.Int("max-concurrent-queries", 0, "Maximal number of queries that are run concurrently against the database server. Zero disables limiting")

func NewExporter(glob string, client *mongo.Client) (*Exporter, error) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("could not glob for aggregation files. glob=%s", glob)
	}

	exporter := &Exporter{aggregations: nil, operations: make(chan aggregationOperation)}
	load := func() {
		var ags []prometheus.Collector
		for _, match := range matches {
			ag, err := aggregations.FromFile(match, client)
			if err != nil {
				log.Printf("could not parse aggregation file. filename=%s", match)
				continue
			}
			ags = append(ags, ag)
		}
		exporter.aggregations = ags
	}
	load()

	go func() {
		for {
			operation := <-exporter.operations
			switch operation.operationType {
			case get:
				for _, elem := range exporter.aggregations {
					operation.result <- elem
				}
				close(operation.result)
			case reload:
				load()
			}
		}
	}()
	return exporter, nil
}

func (e *Exporter) get() chan prometheus.Collector {
	result := make(chan prometheus.Collector)
	e.operations <- aggregationOperation{
		operationType: get,
		result:        result,
	}
	return result
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	result := e.get()
	for ag := range result {
		ag.Describe(ch)
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	wg := sizedwaitgroup.New(*maxConcurrentQueries)
	defer wg.Wait()
	result := e.get()
	for ag := range result {
		wg.Add()
		go func(collector prometheus.Collector) {
			defer wg.Done()
			collector.Collect(ch)
		}(ag)
	}
}

func (e *Exporter) Reload() {
	e.operations <- aggregationOperation{
		operationType: reload,
		result:        nil,
	}
}
