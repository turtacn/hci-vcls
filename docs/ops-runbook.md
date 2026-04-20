# Operational Runbook

## Health Checks

```sh
curl http://localhost:8080/api/v1/status
curl http://localhost:8080/api/v1/degradation
```

## Failure Recovery

**MySQL Unavailable**
- Degradation Level shifts to \`Major\`.
- HA engine halts new boot claims to avoid split brain.
- Restore MySQL link and metrics will automatically clear.

**Split-Brain Mitigation**
- Handled safely natively via MySQL Optimistic Locking natively via \`token\` comparisons across the \`ha_vm_state\` table.

## Monitoring

**Prometheus Keys:**
- \`hci_vcls_election_total\`
- \`hci_vcls_ha_execution_duration_seconds\`
- \`hci_vcls_degradation_level\`
- \`hci_vcls_sweeper_last_run_timestamp_seconds\`
