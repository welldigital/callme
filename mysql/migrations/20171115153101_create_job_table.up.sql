CREATE TABLE `job` (
  `idjob` INT NOT NULL,
  `idschedule` INT NULL,
  `when` DATETIME NOT NULL,
  `arn` VARCHAR(2048) NULL,
  `payload` MEDIUMTEXT NULL,
  PRIMARY KEY (`idjob`));
