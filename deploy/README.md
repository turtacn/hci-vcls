# HCI vCLS Deployment

## Local Startup

You can start the full stack (ZooKeeper, MySQL, HCI vCLS) locally using Docker Compose:

1. Clone the repository
```bash
git clone https://github.com/turtacn/hci-vcls.git
cd hci-vcls
```

2. Start docker-compose
```bash
docker compose -f deploy/docker-compose.dev.yml up --build -d
```

3. Check Status
```bash
curl -X GET http://localhost:8080/api/v1/status
```

## API Testing

You can use `curl` to interact with the API endpoints once the service is up.

1. **Check Status**
```bash
curl -X GET http://localhost:8080/api/v1/status
```

2. **Fetch Degradation Level**
```bash
curl -X GET http://localhost:8080/api/v1/degradation
```

3. **Manual HA Trigger**
```bash
curl -X POST http://localhost:8080/api/v1/ha/evaluate \
     -H "Content-Type: application/json" \
     -d '{"cluster_id": "cluster-1"}'
```

## Architecture

```text
    [ REST API / gRPC ] ---> (App Orchestration Layer)
                                      |
         +----------------------------+-----------------------+
         |                            |                       |
    [ HA Planner ]             [ FDM Evaluator ]        [ State Machine ]
    (Score & Route)          (Health rules & level)    (State transitions)
         |                            |                       |
  [ Infrastr. Adapters ]     [ Heartbeat & Election ]  [ VCLS Aggregate View ]
    (CFS, QM, Witness)          (Nodes Health)            (Protected VMs)
```

## Key Configurations

| Config Key | Default | Description |
|---|---|---|
| `node.node_id` | "" | Unique Node ID |
| `node.cluster_id` | "" | Cluster ID |
| `fdm.eval_interval` | 10s | FDM Loop frequency |
| `ha.batch_size` | 5 | HA Task batch execution limit |
| `election.lease_ttl` | 10s | TTL for election leasing |
