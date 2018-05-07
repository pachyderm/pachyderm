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

type cacheStats struct {
	cacheName    string
	descriptions map[string]*prometheus.Desc
	*groupcache.Stats
}

func NewCacheStats(cacheName string, groupCacheStats *groupcache.Stats) *cacheStats {
	c := &cacheStats{cacheName, make(map[string]*prometheus.Desc), groupCacheStats}
	if err := prometheus.Register(c); err != nil {
		logrus.Infof("error registering prometheus metric: %v", err)
	}
	return c
}

func (c *cacheStats) Describe(ch chan<- *prometheus.Desc) {
	for _, statFieldName := range groupCacheStatFields {
		statName := c.statName(statFieldName)
		desc := prometheus.NewDesc(
			statName,
			fmt.Sprintf("groupcache %v", statFieldName),
			[]string{},
			nil,
		)
		c.descriptions[statName] = desc
		ch <- desc
	}
}

func (c *cacheStats) Collect(ch chan<- prometheus.Metric) {
	for _, statFieldName := range groupCacheStatFields {
		r := reflect.ValueOf(c) // or do I need to ref c.Stats ??
		value := reflect.Indirect(r).FieldByName(statFieldName)
		metric, err := prometheus.NewConstMetric(
			c.descriptions[c.statName(statFieldName)],
			prometheus.GaugeValue,
			float64(value.Int()),
		)
		if err != nil {
			logrus.Infof("error reporting cache prometheus metric: %v", err)
		} else {
			ch <- metric
		}
	}
}

func (c *cacheStats) statName(fieldName string) string {
	var tokens []string
	for _, token := range camelcase.Split(fieldName) {
		tokens = append(tokens, strings.ToLower(token))
	}
	groupCacheStatName := strings.Join(tokens, "_")

	return fmt.Sprintf("pachd_cache_%v_%v_gauge", c.cacheName, groupCacheStatName)
}
