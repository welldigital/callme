CREATE TABLE `crontab` (
  `idcrontab` INT NULL AUTO_INCREMENT,
  `idschedule` INT NOT NULL,
  `crontab` VARCHAR(256) NOT NULL,
  `previous` DATETIME NOT NULL,
  `next` DATETIME NOT NULL,
  `lastupdated` DATETIME NOT NULL,
  PRIMARY KEY (`idcrontab`));
