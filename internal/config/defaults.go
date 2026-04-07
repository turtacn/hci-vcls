package config

import (
	"time"

	"github.com/spf13/viper"
)

func setDefaults(v *viper.Viper) {
	v.SetDefault("node.node_id", "")
	v.SetDefault("node.cluster_id", "")
	v.SetDefault("node.host_ip", "127.0.0.1")

	v.SetDefault("server.http_addr", ":8080")
	v.SetDefault("server.grpc_addr", ":9090")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")

	v.SetDefault("zk.endpoints", []string{"127.0.0.1:2181"})
	v.SetDefault("zk.session_timeout", 5*time.Second)

	v.SetDefault("mysql.dsn", "user:pass@tcp(127.0.0.1:3306)/vcls")
	v.SetDefault("mysql.max_open_conns", 20)

	v.SetDefault("heartbeat.interval", 5*time.Second)
	v.SetDefault("heartbeat.timeout", 15*time.Second)
	v.SetDefault("heartbeat.window_size", 10)

	v.SetDefault("election.lease_ttl", 10*time.Second)
	v.SetDefault("election.renew_interval", 3*time.Second)

	v.SetDefault("fdm.eval_interval", 10*time.Second)
	v.SetDefault("fdm.quorum_ratio", 0.51)
	v.SetDefault("fdm.critical_threshold", 3)

	v.SetDefault("ha.batch_size", 5)
	v.SetDefault("ha.batch_interval", 2*time.Second)
	v.SetDefault("ha.cooldown", 30*time.Second)
	v.SetDefault("ha.fail_fast", true)

	v.SetDefault("cache.default_ttl", 60*time.Second)
	v.SetDefault("cache.max_entries", 1000)

	v.SetDefault("metrics.addr", ":9091")
	v.SetDefault("metrics.namespace", "hci_vcls")

	v.SetDefault("grpc.max_recv_msg_size", 4194304) // 4MB
}

