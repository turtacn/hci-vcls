package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusMetrics struct {
	haBootTotal      *prometheus.CounterVec
	haBootDuration   *prometheus.HistogramVec
	fdmHeartbeatLost *prometheus.CounterVec
	degradationLevel *prometheus.GaugeVec
	cacheAgeSeconds  *prometheus.GaugeVec
	electionTotal    *prometheus.CounterVec
	leaderChanges    *prometheus.CounterVec
}

func NewPrometheusMetrics(registerer prometheus.Registerer) (*PrometheusMetrics, error) {
	m := &PrometheusMetrics{
		haBootTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: MetricHABootTotal,
			Help: "Total number of HA boot attempts",
		}, []string{LabelClusterID, LabelNodeID, LabelVMID, LabelBootPath, LabelResult}),
		haBootDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: MetricHABootDuration,
			Help: "Duration of HA boot attempts in seconds",
		}, []string{LabelClusterID, LabelNodeID, LabelVMID, LabelBootPath, LabelResult}),
		fdmHeartbeatLost: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: MetricFDMHeartbeatLost,
			Help: "Total number of FDM heartbeats lost",
		}, []string{LabelClusterID, LabelNodeID}),
		degradationLevel: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: MetricDegradationLevel,
			Help: "Current degradation level of the cluster",
		}, []string{LabelClusterID, LabelDegradationLevel}),
		cacheAgeSeconds: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: MetricCacheAgeSeconds,
			Help: "Age of the cache in seconds",
		}, []string{LabelClusterID, LabelNodeID}),
		electionTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: MetricElectionTotal,
			Help: "Total number of elections held",
		}, []string{LabelClusterID, LabelNodeID}),
		leaderChanges: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: MetricLeaderChanges,
			Help: "Total number of leader changes",
		}, []string{LabelClusterID, LabelNodeID}),
	}

	if registerer != nil {
		if err := registerer.Register(m.haBootTotal); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.haBootDuration); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.fdmHeartbeatLost); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.degradationLevel); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.cacheAgeSeconds); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.electionTotal); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.leaderChanges); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *PrometheusMetrics) IncHABootTotal(labels MetricLabels) {
	m.haBootTotal.With(prometheus.Labels(labels)).Inc()
}

func (m *PrometheusMetrics) ObserveHABootDuration(durationSeconds float64, labels MetricLabels) {
	m.haBootDuration.With(prometheus.Labels(labels)).Observe(durationSeconds)
}

func (m *PrometheusMetrics) IncFDMHeartbeatLost(labels MetricLabels) {
	m.fdmHeartbeatLost.With(prometheus.Labels(labels)).Inc()
}

func (m *PrometheusMetrics) SetDegradationLevel(level float64, labels MetricLabels) {
	m.degradationLevel.With(prometheus.Labels(labels)).Set(level)
}

func (m *PrometheusMetrics) SetCacheAgeSeconds(ageSeconds float64, labels MetricLabels) {
	m.cacheAgeSeconds.With(prometheus.Labels(labels)).Set(ageSeconds)
}

func (m *PrometheusMetrics) IncElectionTotal(labels MetricLabels) {
	m.electionTotal.With(prometheus.Labels(labels)).Inc()
}

func (m *PrometheusMetrics) IncLeaderChanges(labels MetricLabels) {
	m.leaderChanges.With(prometheus.Labels(labels)).Inc()
}

//Personal.AI order the ending