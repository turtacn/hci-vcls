# API Reference

## REST API (v1)

### `GET /api/v1/version`
Returns build metadata.

### `GET /api/v1/status`
Returns global cluster leader state, protected vm count, etc.

### `GET /api/v1/degradation`
Returns FDM node perspective degradation view (`None`, `Minor`, `Major`, `Critical`).

### `POST /api/v1/ha/evaluate`
Payload: `{"cluster_id": "c1"}`
Manually trigger HA planner validation.

### `GET /api/v1/ha/tasks`
Returns current processing tasks.

### `GET /api/v1/ha/plan/:id`
Fetch specific plan output by uuid.

### `GET /api/v1/sweeper/status`
Returns last run timestamp and total claims released.

### `GET /api/v1/audit/query`
Audit HA decision loops.

## gRPC Interface

Available on default `9090` or `cfg.Server.GRPCAddr`.
Supports matching methods directly equivalent to the REST endpoints.
