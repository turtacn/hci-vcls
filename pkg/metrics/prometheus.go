package metrics

import "github.com/prometheus/client_golang/prometheus"

type PrometheusMetrics struct {
	electionTotal        *prometheus.CounterVec
	leaderChange         *prometheus.CounterVec
	heartbeatLost        *prometheus.CounterVec
	degradationLevel     *prometheus.GaugeVec
	haTaskTotal          *prometheus.CounterVec
	haExecutionDuration  *prometheus.HistogramVec
	protectedVMCount     *prometheus.GaugeVec
	sweeperReleaseOK     prometheus.Counter
	sweeperReleaseFailed prometheus.Counter
	sweeperLastRunUnix   prometheus.Gauge
}

var _ Metrics = &PrometheusMetrics{}

func NewPrometheusMetrics(registerer prometheus.Registerer) (*PrometheusMetrics, error) {
	m := &PrometheusMetrics{
		electionTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "hci_election_total",
			Help: "Total number of elections held",
		}, []string{"node", "result"}),
		leaderChange: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "hci_leader_change_total",
			Help: "Total number of leader changes",
		}, []string{"cluster"}),
		heartbeatLost: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "hci_heartbeat_lost_total",
			Help: "Total number of heartbeats lost",
		}, []string{"node", "cluster"}),
		degradationLevel: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "hci_degradation_level",
			Help: "Current degradation level of the cluster",
		}, []string{"cluster"}),
		haTaskTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "hci_ha_task_total",
			Help: "Total number of HA tasks executed",
		}, []string{"cluster", "status"}),
		haExecutionDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "hci_ha_execution_duration_seconds",
			Help: "Duration of HA task executions in seconds",
		}, []string{"cluster"}),
		protectedVMCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "hci_protected_vm_count",
			Help: "Current number of protected VMs",
		}, []string{"cluster"}),
		sweeperReleaseOK: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "hci_sweeper_release_ok_total",
			Help: "Total number of successful sweeper releases",
		}),
		sweeperReleaseFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "hci_sweeper_release_failed_total",
			Help: "Total number of failed sweeper releases",
		}),
		sweeperLastRunUnix: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "hci_sweeper_last_run_unix",
			Help: "Timestamp of the last sweeper run",
		}),
	}

	if registerer != nil {
		if err := registerer.Register(m.electionTotal); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.leaderChange); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.heartbeatLost); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.degradationLevel); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.haTaskTotal); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.haExecutionDuration); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.protectedVMCount); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.sweeperReleaseOK); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.sweeperReleaseFailed); err != nil {
			return nil, err
		}
		if err := registerer.Register(m.sweeperLastRunUnix); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *PrometheusMetrics) IncElectionTotal(node, result string) {
	m.electionTotal.WithLabelValues(node, result).Inc()
}

func (m *PrometheusMetrics) IncLeaderChange(cluster string) {
	m.leaderChange.WithLabelValues(cluster).Inc()
}

func (m *PrometheusMetrics) IncHeartbeatLost(node, cluster string) {
	m.heartbeatLost.WithLabelValues(node, cluster).Inc()
}

func (m *PrometheusMetrics) SetDegradationLevel(cluster string, level float64) {
	m.degradationLevel.WithLabelValues(cluster).Set(level)
}

func (m *PrometheusMetrics) IncHATaskTotal(cluster, status string) {
	m.haTaskTotal.WithLabelValues(cluster, status).Inc()
}

func (m *PrometheusMetrics) ObserveHAExecutionDuration(cluster string, seconds float64) {
	m.haExecutionDuration.WithLabelValues(cluster).Observe(seconds)
}

func (m *PrometheusMetrics) SetProtectedVMCount(cluster string, count float64) {
	m.protectedVMCount.WithLabelValues(cluster).Set(count)
}

func (m *PrometheusMetrics) IncSweeperReleaseOK() {
	m.sweeperReleaseOK.Inc()
}

func (m *PrometheusMetrics) IncSweeperReleaseFailed() {
	m.sweeperReleaseFailed.Inc()
}

func (m *PrometheusMetrics) SetSweeperLastRunUnix(ts float64) {
	m.sweeperLastRunUnix.Set(ts)
}
