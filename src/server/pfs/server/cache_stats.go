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
)

type cacheStats struct {
	cacheName    string
	descriptions map[string]*prometheus.Desc
	*groupcache.Stats
	mutex *sync.Mutex // For concurrent descriptions map access
}

// RegisterCacheStats creates a new wrapper for groupcache stats that implements
// the prometheus.Collector interface, and registers it
func RegisterCacheStats(cacheName string, groupCacheStats *groupcache.Stats) {
	c := &cacheStats{cacheName, make(map[string]*prometheus.Desc), groupCacheStats, &sync.Mutex{}}
	if err := prometheus.Register(c); err != nil {
		logrus.Infof("error registering prometheus metric: %v", err)
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
			c.mutex.Lock()
			defer c.mutex.Unlock()
			c.descriptions[statName] = desc
		}()
		ch <- desc
	}
}

func (c *cacheStats) Collect(ch chan<- prometheus.Metric) {
	r := reflect.ValueOf(c)
	for _, statFieldName := range groupCacheStatFields() {
		value := reflect.Indirect(r).FieldByName(statFieldName)
		func() {
			c.mutex.Lock()
			defer c.mutex.Unlock()
			metric, err := prometheus.NewConstMetric(
				c.descriptions[c.statName(statFieldName)],
				prometheus.GaugeValue,
				float64(value.Int()),
			)
			if err != nil {
				logrus.Infof("error reporting prometheus cache metric %v: %v", c.statName(statFieldName), err)
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
	s := &groupcache.Stats{}
	e := reflect.ValueOf(s).Elem()
	t := e.Type()
	for i := 0; i < e.NumField(); i++ {
		fields = append(fields, t.Field(i).Name)
	}
	return fields
}
