## README.md

<div align="center">
  <img src="logo.png" alt="hci-vcls Logo" width="200" height="200">

  # hci-vcls

  [![Build Status](https://img.shields.io/github/actions/workflow/status/turtacn/hci-vcls/ci.yml?branch=main&label=build)](https://github.com/turtacn/hci-vcls/actions)
  [![Go Report Card](https://goreportcard.com/badge/github.com/turtacn/hci-vcls)](https://goreportcard.com/report/github.com/turtacn/hci-vcls)
  [![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
  [![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org/)
  [![Release](https://img.shields.io/github/v/release/turtacn/hci-vcls?include_prereleases)](https://github.com/turtacn/hci-vcls/releases)
  [![中文文档](https://img.shields.io/badge/docs-%E4%B8%AD%E6%96%87-red)](README-zh.md)

  Minority-quorum HA for HCI clusters — inspired by vCenter + vCLS, built for the real world.
</div>

---

> This document is also available in [Simplified Chinese（简体中文）](README-zh.md).

---

## Mission Statement

HCI clusters die in silence when control-plane quorum is lost.
`hci-vcls` gives every surviving node the autonomous ability to detect failures,
maintain a local HA metadata cache, and restart protected VMs — without waiting
for ZooKeeper, CFS, or MySQL to recover.

Inspired by VMware's vCenter + vCLS decoupling philosophy, `hci-vcls` extracts
cluster-services liveness from the availability of the management plane, and
delivers it as a lightweight, embeddable Go library and daemon that integrates
with existing Proxmox-VE-based HCI stacks.

---

## Why hci-vcls?

| Problem | What breaks today | What hci-vcls does |
|---|---|---|
| ZK minority-partition | HA scheduling blocked on ZK write lock | FDM agent runs independently; local cache drives qm start |
| CFS read-only | VM config unreadable; HA cannot determine target | Snapshot cache serves last-known VM config |
| MySQL unavailable | Scheduler cannot commit state transitions | Decoupled boot path bypasses MySQL write requirement for idempotent start |
| svc-master offline | No arbiter for VM placement | vCLS-equivalent agents elect a local master per fault domain |
| 2-node cluster | Any single failure breaks quorum | Asymmetric ZK weight + witness integration path |
| Silent failure | Operator unaware of degraded mode | Explicit state machine with observable degradation levels |

---

## Key Features

- Fault Domain Monitor (FDM) daemon with L1 UDP heartbeat and L2 storage heartbeat,
  independent of ZK/CFS/MySQL health.
- FDM Agent per compute node with autonomous leader election (Raft-lite, no ZK dependency).
- Local HA metadata snapshot cache: persists VM HA configuration to node-local storage,
  readable even when cluster file system is unavailable.
- Minority-quorum boot path: when ZK is in read-only minority mode and MySQL primary
  is directly reachable, protected VMs are started via idempotent qm start with
  MySQL-layer optimistic locking to prevent double-boot.
- vCLS-equivalent cluster services layer: decouples cluster health signalling from
  management plane availability.
- Explicit degradation state machine: every node knows its current degradation level
  (NORMAL / ZK_RO / CFS_RO / MYSQL_UNAVAIL / ISOLATED) and acts accordingly.
- Zero external dependencies for the core HA path: no etcd, no Consul, no additional
  infrastructure required.
- Pluggable witness integration: optional ZK witness node support for 2-node and
  stretched-cluster topologies.
- Full observability: Prometheus metrics, structured JSON logs, and a gRPC status API.

---

## Architecture Overview

See [docs/architecture.md](docs/architecture.md) for the full architecture design,
component breakdown, and sequence diagrams.

High-level layers:

```text
+----------------------------------------------------------+
|                    hci-vcls daemon                       |
|                                                          |
|  +----------------+   +----------------+                |
|  |  vCLS Agent    |   |  FDM Daemon    |                |
|  |  (per cluster) |   |  (per node)    |                |
|  +-------+--------+   +-------+--------+                |
|          |                    |                          |
|  +-------v--------------------v--------+                |
|  |        Cluster State Machine        |                |
|  |   NORMAL / ZK_RO / CFS_RO /        |                |
|  |   MYSQL_UNAVAIL / ISOLATED          |                |
|  +-------+-----------------------------+                |
|          |                                              |
|  +-------v-----------------------------------------+   |
|  |           HA Execution Engine                   |   |
|  |  snapshot cache | minority boot | qm adapter    |   |
|  +-------------------------------------------------+   |
+----------------------------------------------------------+
        |                  |                  |
   ZooKeeper           pmxcfs / CFS        MySQL
   (optional)          (optional)          (optional)
```

---

## Getting Started

### Prerequisites

* Go 1.21 or later
* A Proxmox VE 7.x / 8.x node (for full integration) or any Linux host (for testing)
* `qm` available in PATH for VM lifecycle operations

### Install

```text
go install github.com/turtacn/hci-vcls/cmd/hci-vcls@latest
```

Or build from source:

```shell
git clone https://github.com/turtacn/hci-vcls.git
cd hci-vcls
make build
# binary is at bin/hci-vcls
```

### Quick Start

Start the FDM daemon on a compute node:

```shell
hci-vcls fdm start \
  --node-id=node-01 \
  --cluster-id=hci-prod \
  --peers=192.168.1.1:7946,192.168.1.2:7946,192.168.1.3:7946 \
  --ha-meta-cache-dir=/var/lib/hci-vcls/cache \
  --zk-endpoints=192.168.1.1:2181,192.168.1.2:2181,192.168.1.3:2181
```

Query cluster degradation state:

```shell
hci-vcls status --output=json
```

Example output:

```json
{
  "node_id": "node-01",
  "cluster_id": "hci-prod",
  "degradation_level": "ZK_RO",
  "zk_state": "minority_read_only",
  "cfs_state": "read_only",
  "mysql_state": "primary_reachable",
  "ha_capable": true,
  "protected_vms": 12,
  "cache_age_seconds": 47
}
```

Trigger minority-mode HA boot manually (for testing):

```shell
hci-vcls ha boot --vm-id=101 --dry-run
```

---

## Code Example: Embedding the FDM Agent

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/turtacn/hci-vcls/pkg/fdm"
    "github.com/turtacn/hci-vcls/pkg/cache"
    "github.com/turtacn/hci-vcls/pkg/ha"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    cacheStore, err := cache.NewLocalSnapshotStore(cache.Config{
        Dir:             "/var/lib/hci-vcls/cache",
        MaxAgeSeconds:   300,
        CompressEnabled: true,
    })
    if err != nil {
        logger.Error("failed to init cache", "err", err)
        os.Exit(1)
    }

    agent, err := fdm.NewAgent(fdm.AgentConfig{
        NodeID:    "node-01",
        ClusterID: "hci-prod",
        Peers:     []string{"192.168.1.2:7946", "192.168.1.3:7946"},
        Logger:    logger,
        Cache:     cacheStore,
    })
    if err != nil {
        logger.Error("failed to init FDM agent", "err", err)
        os.Exit(1)
    }

    engine, err := ha.NewMinorityBootEngine(ha.EngineConfig{
        Agent:     agent,
        Cache:     cacheStore,
        QMAdapter: ha.NewQMAdapter("/usr/sbin/qm"),
        Logger:    logger,
    })
    if err != nil {
        logger.Error("failed to init HA engine", "err", err)
        os.Exit(1)
    }

    ctx := context.Background()
    if err := engine.RunProtectedVMs(ctx); err != nil {
        logger.Error("minority HA boot failed", "err", err)
        os.Exit(1)
    }
}
```

---

## Project Structure

```text
hci-vcls/
├── cmd/hci-vcls/          # CLI entry point (Cobra)
├── pkg/
│   ├── fdm/               # Fault Domain Monitor daemon and agent
│   ├── vcls/              # vCLS-equivalent cluster services layer
│   ├── ha/                # HA execution engine and minority boot
│   ├── cache/             # Local HA metadata snapshot cache
│   ├── statemachine/      # Cluster degradation state machine
│   ├── zk/                # ZooKeeper adapter and quorum probe
│   ├── mysql/             # MySQL adapter and boot-path decoupling
│   ├── cfs/               # CFS/pmxcfs adapter
│   ├── qm/                # qm CLI adapter for VM lifecycle
│   ├── witness/           # Optional witness node integration
│   ├── metrics/           # Prometheus metrics
│   └── api/               # gRPC + REST status API
├── internal/
│   ├── election/          # Raft-lite leader election
│   ├── heartbeat/         # L1 UDP + L2 storage heartbeat
│   └── logger/            # Structured logger setup
├── docs/
│   ├── architecture.md
│   └── apis.md
└── test/e2e/              # End-to-end test suite
```

---

## Contributing

Contributions are very welcome. `hci-vcls` is an ambitious project and benefits
from diverse engineering perspectives — especially from engineers operating
real HCI clusters.

How to contribute:

1. Read [docs/architecture.md](docs/architecture.md) to understand design principles.
2. Open an issue to discuss your proposed change before writing code.
3. Fork the repository and create a feature branch.
4. Write tests. Every PR must include unit tests; E2E tests for behaviour changes.
5. Run `make lint && make test` before submitting.
6. Submit a pull request with a clear description.

Please follow the Contributor Covenant Code of Conduct.
See [CONTRIBUTING.md](CONTRIBUTING.md) for full guidelines.

---

## License

Copyright 2024 turtacn contributors.

Licensed under the Apache License, Version 2.0.
See [LICENSE](LICENSE) for the full license text.