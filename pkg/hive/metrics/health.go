package metrics

import (
	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/metrics/metric"
)

var Cell = cell.Metric(newHealthMetrics)

type HealthMetrics struct {
	HealthStatusGauge metric.Vec[metric.Gauge]
}

func newHealthMetrics() *HealthMetrics {
	return &HealthMetrics{
		HealthStatusGauge: metric.NewGaugeVec(metric.GaugeOpts{
			ConfigName: "hive_health_status_levels",
			Namespace:  "dolphin",
			Subsystem:  "hive",
			Name:       "status",
			Help:       "Counts of health status levels of Hive components",
		}, []string{"status"}),
	}
}
