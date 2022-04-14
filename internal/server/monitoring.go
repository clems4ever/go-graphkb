package server

import (
	"context"
	"time"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/metrics"
	"github.com/clems4ever/go-graphkb/internal/sources"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func getMonitoringIntervalSeconds() time.Duration {
	intervalDuration := viper.GetDuration("monitoring_interval_duration")
	if intervalDuration == 0 {
		intervalDuration = 60 * time.Second
	}
	return intervalDuration
}

// startGraphSizeMonitoring start a background process regularly checking the size of the graph to expose it as a prometheus metric
func startGraphSizeMonitoring(interval time.Duration, database knowledge.GraphDB, sourcesRegistry sources.Registry) {
	monitorForOneIteration := func() {
		ctx, cancel := context.WithTimeout(context.Background(), interval)
		defer cancel()

		logrus.Debug("Start monitoring the graph size...")
		assetsCount, err := database.CountAssets(ctx)
		if err != nil {
			metrics.GraphAssetsAggregatedGauge.Set(-1)
		} else {
			metrics.GraphAssetsAggregatedGauge.Set(float64(assetsCount))
		}

		relationsCount, err := database.CountRelations(ctx)
		if err != nil {
			metrics.GraphRelationsAggregatedGauge.Set(-1)
		} else {
			metrics.GraphRelationsAggregatedGauge.Set(float64(relationsCount))
		}

		sources, err := sourcesRegistry.ListSources(ctx)
		if err != nil {
			logrus.Errorf("Unable to list sources for monitoring: %v", err)
			metrics.GraphAssetsTotalGauge.Reset()
			metrics.GraphRelationsTotalGauge.Reset()
		}

		for s := range sources {
			assetsCount, err := database.CountAssetsBySource(ctx, s)
			if err != nil {
				logrus.Errorf("Unable to count assets of source %s for monitoring: %s", s, err)
				metrics.GraphAssetsTotalGauge.Reset()
			} else {
				metrics.GraphAssetsTotalGauge.With(prometheus.Labels{"source": s}).Set(float64(assetsCount))
			}

			relationsCount, err := database.CountRelationsBySource(ctx, s)
			if err != nil {
				logrus.Errorf("Unable to count relations of source %s for monitoring: %s", s, err)
				metrics.GraphRelationsTotalGauge.Reset()
			} else {
				metrics.GraphRelationsTotalGauge.With(prometheus.Labels{"source": s}).Set(float64(relationsCount))
			}
		}

		m, err := database.CollectMetrics(ctx)
		if err != nil {
			logrus.Errorf("Unable to collect metrics from the database: %v", err)
			metrics.DatabaseMetricsGauge.Reset()
		} else {
			for k, v := range m {
				metrics.DatabaseMetricsGauge.With(prometheus.Labels{"name": k}).Set(float64(v))
			}
		}
	}

	logrus.Infof("Monitoring of the graph size will happen every %ds", int(interval/time.Second))
	go func() {
		for {
			monitorForOneIteration()
			time.Sleep(interval)
		}
	}()
}
