CREATE TABLE IF NOT EXISTS reconciliation_workflows (
    workflow_id UUID PRIMARY KEY,
    system_ingestion_job_id UUID,
    bank_ingestion_job_id UUID,
    reconciliation_job_id UUID,
    status TEXT NOT NULL,                -- e.g. "IN_PROGRESS", "COMPLETED", "FAILED"
    start_date DATE,                     -- the date range for reconciliation
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);