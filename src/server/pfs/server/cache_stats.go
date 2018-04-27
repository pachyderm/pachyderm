package server

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/golang/groupcache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var groupCacheStatFields = []string{
	"Gets",
	"CacheHits",
	"PeerLoads",
	"PeerErrors",
	"Loads",
	"LoadsDeduped",
	"LocalLoads",
	"LocalLoadErrs",
	"ServerRequests",
}

var cacheStats map[string]prometheus.Gauge

func statName(cacheName string, fieldName string) string {
	var tokens []string
	for _, token := range camelcase.Split(fieldName) {
		tokens = append(tokens, strings.ToLower(token))
	}
	groupCacheStatName := strings.Join(tokens, "_")

	return fmt.Sprintf("cache_%v_%v_gauge", cacheName, groupCacheStatName)
}

func initCacheStats(cacheNames []string) {
	cacheStats = make(map[string]prometheus.Gauge)
	for _, cacheName := range cacheNames {
		for _, statFieldName := range groupCacheStatFields {
			name := statName(cacheName, statFieldName)
			gauge, ok := cacheStats[name]
			if !ok {
				gauge = prometheus.NewGauge(
					prometheus.GaugeOpts{
						Namespace: "pachyderm",
						Subsystem: "pachd",
						Name:      name,
						Help:      fmt.Sprintf("gauge for groupcache %v field", statFieldName),
					},
				)
				if err := prometheus.Register(gauge); err != nil {
					logrus.Infof("error registering prometheus metric: %v", err)
				} else {
					cacheStats[name] = gauge
				}
			}
		}
	}
}

func reportCacheStats(cacheName string, stats *groupcache.Stats) {
	for _, statFieldName := range groupCacheStatFields {
		r := reflect.ValueOf(stats)
		value := reflect.Indirect(r).FieldByName(statFieldName)
		cacheStats[statName(cacheName, statFieldName)].Set(value.Float())
	}
}
