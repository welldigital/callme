BEGIN;

CREATE TABLE `job` (
  `idjob` INT NOT NULL AUTO_INCREMENT,
  `idschedule` INT NULL,
  `when` DATETIME NOT NULL,
  `arn` VARCHAR(2048) NULL,
  `payload` MEDIUMTEXT NULL,
  PRIMARY KEY (`idjob`));

CREATE TABLE `jobresponse` (
  `idjobresponse` INT NOT NULL AUTO_INCREMENT,
  `idlease` INT NOT NULL,
  `idjobid` INT NOT NULL,
  `time` DATETIME NOT NULL,
  `response` MEDIUMTEXT NOT NULL,
  `iserror` BIT NOT NULL,
  `error` MEDIUMTEXT NOT NULL,
  PRIMARY KEY (`idjobresponse`));

CREATE TABLE `schedule` (
  `idschedule` INT NOT NULL AUTO_INCREMENT,
  `externalid` VARCHAR(256) NOT NULL,
  `by` VARCHAR(256) NOT NULL,
  `arn` TEXT NOT NULL,
  `payload` MEDIUMTEXT NOT NULL,
  `created` DATETIME NOT NULL,
  `from` DATETIME NOT NULL,
  `active` BIT NOT NULL,
  `deactivateddate` DATETIME,
  PRIMARY KEY (`idschedule`));

CREATE TABLE `crontab` (
  `idcrontab` INT NOT NULL AUTO_INCREMENT,
  `idschedule` INT NOT NULL,
  `crontab` VARCHAR(256) NOT NULL,
  `previous` DATETIME NOT NULL,
  `next` DATETIME NOT NULL,
  `lastupdated` DATETIME NOT NULL,
  PRIMARY KEY (`idcrontab`));

CREATE TABLE lease (
  idlease INT NOT NULL AUTO_INCREMENT,
  `type` VARCHAR(256) NOT NULL,
  lockedby VARCHAR(256) NOT NULL,
  `at` DATETIME NOT NULL,
  `until` DATETIME NOT NULL,
  PRIMARY KEY (`idlease`));

CREATE INDEX idx_lease_type_until ON lease (`type`, `until`);

ALTER TABLE `job`
    ADD CONSTRAINT fk_job_idschedule 
    FOREIGN KEY (idschedule) REFERENCES schedule(idschedule);

ALTER TABLE jobresponse
    ADD CONSTRAINT fk_jobresponse_idlease 
    FOREIGN KEY (idlease) REFERENCES lease(idlease);

ALTER TABLE jobresponse
    ADD CONSTRAINT fk_jobresponse_idjob
    FOREIGN KEY (idjobid) REFERENCES job(idjob);

ALTER TABLE crontab 
    ADD CONSTRAINT fk_crontab_idschedule 
    FOREIGN KEY (idschedule) REFERENCES schedule(idschedule);

COMMIT;