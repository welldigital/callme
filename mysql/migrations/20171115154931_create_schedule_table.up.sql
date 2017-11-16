CREATE TABLE `schedule` (
  `idschedule` INT NULL AUTO_INCREMENT,
  `externalid` VARCHAR(256) NOT NULL,
  `by` VARCHAR(256) NOT NULL,
  `arn` TEXT NOT NULL,
  `payload` MEDIUMTEXT NOT NULL,
  `created` DATETIME NOT NULL,
  `from` DATETIME NOT NULL,
  `active` BIT NOT NULL,
  `deactivateddate` DATETIME,
  PRIMARY KEY (`idschedule`));
