package config

import (
	"github.com/spf13/viper"
)

func setDefaults(v *viper.Viper) {
	v.SetDefault("Node.ID", "local-node")
	v.SetDefault("Node.ClusterID", "default-cluster")
	v.SetDefault("Server.ListenAddr", ":8080")
	v.SetDefault("Metrics.Enabled", true)
	v.SetDefault("Metrics.ListenAddr", ":9090")
	v.SetDefault("GRPC.Enabled", false)
	v.SetDefault("GRPC.ListenAddr", ":50051")
	v.SetDefault("LogLevel", "info")
	v.SetDefault("LogFormat", "text")
	v.SetDefault("ZK.SessionTimeoutMs", 5000)
	v.SetDefault("CFS.TimeoutMs", 3000)
	v.SetDefault("MySQL.MaxOpenConns", 20)
	v.SetDefault("MySQL.MaxIdleConns", 5)
	v.SetDefault("QM.TimeoutMs", 30000)
	v.SetDefault("HA.MaxConcurrentBoots", 5)
	v.SetDefault("HA.BootTimeoutMs", 60000)
	v.SetDefault("HA.MaxRetries", 3)
}

//Personal.AI order the ending
