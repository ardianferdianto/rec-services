ALTER DATABASE reconciliation_services SET TIMEZONE TO 'Asia/Jakarta';
-- system_transactions: internal system transactions
CREATE TABLE IF NOT EXISTS system_transactions (
    id SERIAL PRIMARY KEY,
    trx_id TEXT NOT NULL UNIQUE,
    amount DECIMAL(18, 2) NOT NULL,
    trx_type TEXT NOT NULL,                       -- e.g., "DEBIT" or "CREDIT"
    transaction_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- bank_statements: external statements from various banks
CREATE TABLE IF NOT EXISTS bank_statements (
    id SERIAL PRIMARY KEY,
    unique_id TEXT NOT NULL,
    amount DECIMAL(18, 2) NOT NULL,               -- can be negative for debits
    statement_time TIMESTAMP NOT NULL,
    bank_code TEXT NOT NULL,                      -- identifies which bank
    hash_code TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_bank_stmt UNIQUE (hash_code)
);

-- reconciliation_jobs: track each reconciliation run (for auditing/resume)
CREATE TABLE IF NOT EXISTS reconciliation_jobs (
    job_id UUID PRIMARY KEY,                      -- unique job ID
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- reconciliation_results: summary of each job
CREATE TABLE IF NOT EXISTS reconciliation_results (
    id SERIAL PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES reconciliation_jobs(job_id),
    total_system_tx_count INT NOT NULL,
    total_bank_tx_count INT NOT NULL,
    matched_count INT NOT NULL,
    unmatched_system_count INT NOT NULL,
    unmatched_bank_count INT NOT NULL,
    total_discrepancies DECIMAL(18, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- reconciliation_matched_records: 1-to-1 matches for each job
CREATE TABLE IF NOT EXISTS reconciliation_matched_records (
    id SERIAL PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES reconciliation_jobs(job_id),
    system_tx_id INT NOT NULL REFERENCES system_transactions(id),
    bank_statement_id INT NOT NULL REFERENCES bank_statements(id),
    discrepancy DECIMAL(18, 2) NOT NULL,          -- absolute difference
    matched_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- 1) each transaction can appear only once per job
    CONSTRAINT unique_system_tx_per_job UNIQUE (job_id, system_tx_id),

    -- 2) each bank statement can appear only once per job
    CONSTRAINT unique_statement_per_job UNIQUE (job_id, bank_statement_id)
);

-- unmatched system tx details
CREATE TABLE IF NOT EXISTS reconciliation_unmatched_system_tx (
    id SERIAL PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES reconciliation_jobs(job_id),
    trx_id TEXT NOT NULL,
    amount DECIMAL(18, 2) NOT NULL,
    trx_type TEXT NOT NULL,
    transaction_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- unmatched bank tx details
CREATE TABLE IF NOT EXISTS reconciliation_unmatched_bank_tx (
    id SERIAL PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES reconciliation_jobs(job_id),
    unique_id TEXT NOT NULL,
    amount DECIMAL(18, 2) NOT NULL,
    statement_time TIMESTAMP NOT NULL,
    bank_code TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);