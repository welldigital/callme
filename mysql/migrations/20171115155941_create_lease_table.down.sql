BEGIN;

DROP TABLE lease;
DROP INDEX idx_lease_type_until ON lease;

COMMIT;
