CREATE TABLE IF NOT EXISTS ingestion_jobs (
    job_id UUID PRIMARY KEY,
    file_type TEXT NOT NULL,  -- "SYSTEM_TX" or "BANK_STMT"
    file_name TEXT NOT NULL, -- key in MinIO, e.g. "uploads/system/...csv"
    total_lines_processed BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL,     -- "PENDING", "IN_PROGRESS", "COMPLETED", "FAILED", ...
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);