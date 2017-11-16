BEGIN;

CREATE TABLE lease (
  idlease INT NULL AUTO_INCREMENT,
  `type` VARCHAR(256) NOT NULL,
  lockedby VARCHAR(256) NOT NULL,
  `at` DATETIME NOT NULL,
  `until` DATETIME NOT NULL,
  PRIMARY KEY (`idlease`));

CREATE INDEX idx_lease_type_until ON lease (`type`, `until`);

COMMIT;