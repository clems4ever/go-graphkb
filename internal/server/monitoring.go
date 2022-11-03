package server

import (
	"context"
	"time"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type dbMonitor struct {
	database knowledge.GraphDB

	AssetCount    int64
	RelationCount int64

	AssetsBySource    map[string]int64
	RelationsBySource map[string]int64
}

func newDBMonitor(database knowledge.GraphDB) *dbMonitor {
	return &dbMonitor{
		database:          database,
		AssetsBySource:    map[string]int64{},
		RelationsBySource: map[string]int64{},
	}
}

func (m *dbMonitor) Start() {
	interval := getMonitoringIntervalSeconds()

	logrus.Infof("monitoring of the database every %s", interval)
	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), interval)
			err := m.refresh(ctx)
			if err != nil {
				logrus.Errorf("db monitor: %s", err)
			}
			cancel()

			time.Sleep(interval)
		}
	}()
}

func (m *dbMonitor) refresh(ctx context.Context) error {
	c, err := m.database.CountAssets(ctx)
	if err != nil {
		metrics.GraphAssetsAggregatedGauge.Set(-1)
	} else {
		m.AssetCount = c
		metrics.GraphAssetsAggregatedGauge.Set(float64(c))
	}

	c, err = m.database.CountRelations(ctx)
	if err != nil {
		metrics.GraphRelationsAggregatedGauge.Set(-1)
	} else {
		m.RelationCount = c
		metrics.GraphRelationsAggregatedGauge.Set(float64(c))
	}

	ac, err := m.database.CountAssetsBySource(ctx)
	if err != nil {
		metrics.GraphAssetsTotalGauge.Reset()
	} else {
		m.AssetsBySource = ac
		for name, count := range ac {
			metrics.GraphAssetsTotalGauge.With(prometheus.Labels{"source": name}).Set(float64(count))
		}
	}

	ac, err = m.database.CountRelationsBySource(ctx)
	if err != nil {
		metrics.GraphRelationsTotalGauge.Reset()
	} else {
		m.RelationsBySource = ac
		for name, count := range ac {
			metrics.GraphRelationsTotalGauge.With(prometheus.Labels{"source": name}).Set(float64(count))
		}
	}

	dbMetrics, err := m.database.CollectMetrics(ctx)
	if err != nil {
		logrus.Errorf("Unable to collect metrics from the database: %v", err)
		metrics.DatabaseMetricsGauge.Reset()
	} else {
		for k, v := range dbMetrics {
			metrics.DatabaseMetricsGauge.With(prometheus.Labels{"name": k}).Set(float64(v))
		}
	}

	return nil
}

func getMonitoringIntervalSeconds() time.Duration {
	intervalDuration := viper.GetDuration("monitoring_interval_duration")
	if intervalDuration == 0 {
		intervalDuration = 60 * time.Second
	}
	return intervalDuration
}
