# API Stability Designations

We use semantic versioning.

**Stable:**
- `GET /api/v1/version`
- `GET /api/v1/status`
- `GET /api/v1/degradation`
- `POST /api/v1/ha/evaluate`
- `GET /api/v1/ha/tasks`

**Beta:**
- `GET /api/v1/ha/plan/:id`
- `GET /api/v1/sweeper/status`
- `GET /api/v1/audit/query`

gRPC endpoints share identical status mappings with REST.
