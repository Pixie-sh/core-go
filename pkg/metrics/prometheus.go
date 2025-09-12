package metrics

import "github.com/prometheus/client_golang/prometheus"

type Registry struct {
	*prometheus.Registry
}
