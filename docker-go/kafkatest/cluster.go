package kafkatest

type GenericKafkaResponse struct {
	Kind     string                   `json:"kind"`
	Metadata interface{}              `json:"metadata"`
	Next     string                   `json:"next"`
	Data     []map[string]interface{} `json:"data"`
}

// {"kind":"KafkaClusterList","metadata":{"self":"http://localhost:8082/v3/clusters","next":null},"data":[{"kind":"KafkaCluster","metadata":{"self":"http://localhost:8082/v3/clusters/_euepX0zTBuxPlkBPEu_nA","resource_name":"crn:///kafka=_euepX0zTBuxPlkBPEu_nA"},"cluster_id":"_euepX0zTBuxPlkBPEu_nA","controller":{"related":"http://localhost:8082/v3/clusters/_euepX0zTBuxPlkBPEu_nA/brokers/1001"},"acls":{"related":"http://localhost:8082/v3/clusters/_euepX0zTBuxPlkBPEu_nA/acls"},"brokers":{"related":"http://localhost:8082/v3/clusters/_euepX0zTBuxPlkBPEu_nA/brokers"},"broker_configs":{"related":"http://localhost:8082/v3/clusters/_euepX0zTBuxPlkBPEu_nA/broker-configs"},"consumer_groups":{"related":"http://localhost:8082/v3/clusters/_euepX0zTBuxPlkBPEu_nA/consumer-groups"},"topics":{"related":"http://localhost:8082/v3/clusters/_euepX0zTBuxPlkBPEu_nA/topics"},"partition_reassignments":{"related":"http://localhost:8082/v3/clusters/_euepX0zTBuxPlkBPEu_nA/topics/-/partitions/-/reassignment"}}]}
