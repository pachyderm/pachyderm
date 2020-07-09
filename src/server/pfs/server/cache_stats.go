package server

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/fatih/camelcase"
	"github.com/golang/groupcache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

type cacheStats struct {
	cacheName      string
	descriptions   map[string]*prometheus.Desc
	stats          *groupcache.Stats
	descriptionsMu sync.RWMutex
}

// RegisterCacheStats creates a new wrapper for groupcache stats that implements
// the prometheus.Collector interface, and registers it
func RegisterCacheStats(cacheName string, groupCacheStats *groupcache.Stats) {
	c := &cacheStats{
		cacheName:    cacheName,
		descriptions: make(map[string]*prometheus.Desc),
		stats:        groupCacheStats,
	}
	if err := prometheus.Register(c); err != nil {
		// metrics may be redundantly registered; ignore these errors
		if !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
			logrus.Infof("error registering prometheus metric: %v", err)
		}
	}
}

func (c *cacheStats) Describe(ch chan<- *prometheus.Desc) {
	for _, statFieldName := range groupCacheStatFields() {
		statName := c.statName(statFieldName)
		desc := prometheus.NewDesc(
			statName,
			fmt.Sprintf("groupcache %v", statFieldName),
			[]string{},
			nil,
		)
		func() {
			c.descriptionsMu.Lock()
			defer c.descriptionsMu.Unlock()
			c.descriptions[statName] = desc
		}()
		ch <- desc
	}
}

func (c *cacheStats) Collect(ch chan<- prometheus.Metric) {
	r := reflect.ValueOf(c.stats)
	for _, statFieldName := range groupCacheStatFields() {
		value := reflect.Indirect(r).FieldByName(statFieldName)
		func() {
			c.descriptionsMu.RLock()
			defer c.descriptionsMu.RUnlock()
			metric, err := prometheus.NewConstMetric(
				c.descriptions[c.statName(statFieldName)],
				prometheus.GaugeValue,
				float64(value.Int()),
			)
			if err != nil {
				// metrics may be redundantly registered; ignore these errors
				if !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
					logrus.Infof("error reporting prometheus cache metric %v: %v", c.statName(statFieldName), err)
				}
			} else {
				ch <- metric
			}
		}()
	}
}

func (c *cacheStats) statName(fieldName string) string {
	var tokens []string
	for _, token := range camelcase.Split(fieldName) {
		tokens = append(tokens, strings.ToLower(token))
	}
	groupCacheStatName := strings.Join(tokens, "_")

	return fmt.Sprintf("pachyderm_pachd_cache_%v_%v_gauge", c.cacheName, groupCacheStatName)
}

func groupCacheStatFields() (fields []string) {
	e := reflect.ValueOf(&groupcache.Stats{}).Elem()
	t := e.Type()
	for i := 0; i < e.NumField(); i++ {
		fields = append(fields, t.Field(i).Name)
	}
	return fields
}
