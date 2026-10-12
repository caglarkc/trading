-- +goose Up
ALTER TABLE decision_runs ADD COLUMN runtime_events jsonb;
ALTER TABLE decision_runs ADD COLUMN duration_ms bigint;
ALTER TABLE decision_runs ADD COLUMN usage jsonb;
CREATE TRIGGER immutable_memory BEFORE UPDATE OR DELETE ON memory_versions FOR EACH ROW EXECUTE FUNCTION reject_journal_mutation();
-- +goose Down
SELECT 1;
