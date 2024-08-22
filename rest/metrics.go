package rest

import (
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
)

var (
	HTTPCollector *Collector
)

func init() {
	serviceMetricCollector := NewPrometheusServiceMetricCollector()
	HTTPCollector = NewCollector(Default(), "http_client", serviceMetricCollector)
}

type ClientMetricCollector interface {
	IncrementCounter(clientName string, eventType string, eventSubType string, value ...float64)
	RecordExecutionTime(clientName, eventType string, eventSubType string, elapsedTime time.Duration)
}

type Config struct {
	Environment string
	Application string
}

func Default() *Config {
	return &Config{
		Environment: strings.ToLower(os.Getenv("ENV")),
		Application: os.Getenv("APP_NAME"),
	}
}

type Collector struct {
	config      *Config
	serviceType string
	collector   ServiceCollector
}

func NewCollector(config *Config, serviceType string, collector ServiceCollector) *Collector {
	return &Collector{
		config:      config,
		serviceType: serviceType,
		collector:   collector,
	}
}

func (c Collector) IncrementCounter(clientName string, eventType string, eventSubType string, value ...float64) {
	c.collector.IncrementCounter(
		CounterDto{
			metricDto: metricDto{
				serviceType:  c.serviceType,
				environment:  c.config.Environment,
				application:  c.config.Application,
				clientName:   clientName,
				eventType:    eventType,
				eventSubType: eventSubType,
			},
			values: value,
		})
}

func (c Collector) RecordExecutionTime(clientName, eventType string, eventSubType string, elapsedTime time.Duration) {
	c.collector.RecordExecutionTime(
		TimerDto{
			metricDto: metricDto{
				serviceType:  c.serviceType,
				environment:  c.config.Environment,
				application:  c.config.Application,
				clientName:   clientName,
				eventType:    eventType,
				eventSubType: eventSubType,
			},
			elapsedTime: elapsedTime,
		})
}

type ServiceCollector interface {
	IncrementCounter(metric CounterDto)
	RecordExecutionTime(metric TimerDto)
}

type ServiceType string

type Mapper interface {
	BuildLabels() []string
}

type metricDto struct {
	serviceType  string
	environment  string
	application  string
	clientName   string
	eventType    string
	eventSubType string
}

func (m *metricDto) BuildLabels() []string {
	return []string{
		m.getValue(m.serviceType),
		m.getValue(m.environment),
		m.getValue(m.application),
		m.getValue(m.clientName),
		m.getValue(m.eventType),
		m.getValue(m.eventSubType),
	}
}

func (m *metricDto) getValue(value string) string {
	if value == "" {
		return "undefined"
	}

	return value
}

type CounterDto struct {
	values []float64

	metricDto
}

func (c *CounterDto) BuildLabels() []string {
	return c.metricDto.BuildLabels()
}

type TimerDto struct {
	elapsedTime time.Duration

	metricDto
}

func (t *TimerDto) BuildLabels() []string {
	return t.metricDto.BuildLabels()
}

func getNamespace() string {
	return "services_dashboard"
}

func getLabels() []string {
	return []string{
		"service_type",
		"environment",
		"application",
		"client_name",
		"event_type",
		"event_subtype",
	}
}

type PrometheusServiceMetricCollector struct {
	Counter *prometheus.CounterVec
	Gauge   *prometheus.GaugeVec
	Summary *prometheus.SummaryVec
}

func (p *PrometheusServiceMetricCollector) IncrementCounter(counterDto CounterDto) {
	if len(counterDto.values) > 0 {
		p.Counter.WithLabelValues(counterDto.BuildLabels()...).Add(counterDto.values[0])
	} else {
		p.Counter.WithLabelValues(counterDto.BuildLabels()...).Inc()
	}
}

func (p *PrometheusServiceMetricCollector) RecordExecutionTime(timerDto TimerDto) {
	p.Summary.WithLabelValues(timerDto.BuildLabels()...).Observe(float64(timerDto.elapsedTime.Milliseconds()))
}

func NewPrometheusServiceMetricCollector() *PrometheusServiceMetricCollector {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: getNamespace(),
			Name:      "services_counters_total",
			Help:      "Service counters",
		}, getLabels(),
	)

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: getNamespace(),
			Name:      "services_gauges",
			Help:      "Service gauges",
		}, getLabels(),
	)

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: getNamespace(),
			Name:      "services_timers",
			Help:      "Service timers",
			Objectives: map[float64]float64{
				0.5:  0.05,  // Average
				0.95: 0.01,  // P95
				0.99: 0.001, // P99
			},
		}, getLabels(),
	)

	register(
		counter,
		summary,
		gauge,
	)

	return &PrometheusServiceMetricCollector{
		Counter: counter,
		Gauge:   gauge,
		Summary: summary,
	}
}

func register(collectors ...prometheus.Collector) {
	for i := range len(collectors) {
		err := prometheus.Register(collectors[i])
		if err != nil {
			log.Error(err)
			continue
		}
	}
}
