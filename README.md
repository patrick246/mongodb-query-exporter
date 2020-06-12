# mongodb-query-exporter
Export data from MongoDB as prometheus metrics.

This is still in development and a proof of concept for now. It has never been tested on a real application database, just test datasets.

## How does it work?
You write an aggregation pipeline, that returns data in the form of
```json 
{
    "labels": {
        "labelKey": "labelValue",
        "anotherKey": "anotherValue"
    },
    "value": 1234
}
```

MongoDB Query Exporter takes these results and exports every result of the pipeline as a prometheus metric.

The pipeline needs to be placed in a JSON file with the following properties.

| Property | Data type | Description |
| ---------|-----------|--------------|
| type | string | The prometheus metric type to export. Only gauge is supported currently. Support for more types, e.g. histograms are planned. |
| name | string | The name of the exported prometheus metric |
| help | string | The description text in the prometheus exposition format. This explains what the metric is supposed to represent. |
| labels | string[] | A string array of label names. These need to match exactly the labels returned from the aggregation pipeline. |
| database | string | The MongoDB database to run this aggregation pipeline in. |
| collection | string | The MongoDB collection in the database to run this aggregation pipeline on. |
| pipeline | Array of pipeline stages | This is the actual aggregation pipeline. It is executed every scrape. MongoDB Extended JSON is supported. See [MongoDB docs](https://docs.mongodb.com/manual/reference/operator/aggregation-pipeline/) for more information on how to write his.|


## Why would I want that?
You can generate nice dashboards in Grafana, based on your data in Prometheus. Also alerting on stuff happening in the database, but I think this is better handled by the application. It is however an option with this exporter.

## Roadmap
* Pipeline validation
* Testing this with real application databases
* Support for other metric types than gauges
