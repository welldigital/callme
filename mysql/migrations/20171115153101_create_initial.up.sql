START TRANSACTION;

	CREATE TABLE `job` (
	  `idjob` INT NOT NULL AUTO_INCREMENT,
	  `idschedule` INT NULL,
	  `when` DATETIME(6) NOT NULL,
	  `arn` VARCHAR(2048) NULL,
	  `payload` MEDIUMTEXT NULL,
	  PRIMARY KEY (`idjob`));

	CREATE TABLE joblease (
	  idjoblease INT NOT NULL AUTO_INCREMENT,
	  idjob INT NOT NULL,
	  lockedby VARCHAR(256) NOT NULL,
	  `at` DATETIME(6) NOT NULL,
	  `until` DATETIME(6) NOT NULL,
	  PRIMARY KEY (idjoblease));

	CREATE TABLE `jobresponse` (
	  `idjobresponse` INT NOT NULL AUTO_INCREMENT,
	  `idjobid` INT NOT NULL,
	  `time` DATETIME(6) NOT NULL,
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
	  `created` DATETIME(6) NOT NULL,
	  `active` BIT NOT NULL,
	  `deactivateddate` DATETIME(6),
	  PRIMARY KEY (`idschedule`));

	CREATE TABLE `crontab` (
	  `idcrontab` INT NOT NULL AUTO_INCREMENT,
	  `idschedule` INT NOT NULL,
	  `crontab` VARCHAR(256) NOT NULL,
	  `previous` DATETIME(6) NOT NULL,
	  `next` DATETIME(6) NOT NULL,
	  `lastupdated` DATETIME(6) NOT NULL,
	  PRIMARY KEY (idcrontab));

	CREATE TABLE crontablease (
	  idcrontablease INT NOT NULL AUTO_INCREMENT,
	  idcrontab INT NOT NULL,
	  lockedby VARCHAR(256) NOT NULL,
	  `at` DATETIME(6) NOT NULL,
	  `until` DATETIME(6) NOT NULL,
		rescinded BIT NOT NULL,
	  PRIMARY KEY (idcrontablease));

	CREATE INDEX idx_joblease_idjob_until ON joblease (`idjob`, `until`);
  CREATE INDEX idx_crontablease_idschedule_until ON crontablease (`idcrontab`, `until`);

	ALTER TABLE `job`
		ADD CONSTRAINT fk_job_idschedule 
		FOREIGN KEY (idschedule) REFERENCES schedule(idschedule);

	ALTER TABLE joblease
		ADD CONSTRAINT fk_joblease_job
		FOREIGN KEY (idjob) REFERENCES `job`(idjob);

	ALTER TABLE jobresponse
		ADD CONSTRAINT fk_jobresponse_idjob
		FOREIGN KEY (idjobid) REFERENCES `job`(idjob);

	ALTER TABLE crontab 
		ADD CONSTRAINT fk_crontab_idschedule 
		FOREIGN KEY (idschedule) REFERENCES `schedule`(idschedule);
		
	ALTER TABLE crontablease
		ADD CONSTRAINT fk_crontablease_crontab
		FOREIGN KEY (idcrontab) REFERENCES crontab(idcrontab);

COMMIT;
