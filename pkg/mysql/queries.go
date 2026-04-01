package mysql

const (
	queryUpsertVM         = `INSERT INTO vm_records (vm_id, cluster_id, current_host, power_state, protected, updated_at) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE current_host=?, power_state=?, protected=?, updated_at=?`
	queryGetVM            = `SELECT * FROM vm_records WHERE vm_id = ?`
	queryListVMsByCluster = `SELECT * FROM vm_records WHERE cluster_id = ?`
	queryListProtectedVMs = `SELECT * FROM vm_records WHERE protected = 1 AND cluster_id = ?`

	queryCreateTask      = `INSERT INTO ha_tasks (id, vm_id, cluster_id, source_host, target_host, boot_path, status, batch_no, retry_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	queryUpdateTaskStatus = `UPDATE ha_tasks SET status = ?, updated_at = ? WHERE id = ?`
	queryListTasksByPlan = `SELECT * FROM ha_tasks WHERE plan_id = ?`

	queryCreatePlan = `INSERT INTO ha_plans (id, cluster_id, trigger, degradation, task_count, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	queryGetPlan    = `SELECT * FROM ha_plans WHERE id = ?`
)

// Personal.AI order the ending
