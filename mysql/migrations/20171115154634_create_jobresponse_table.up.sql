CREATE TABLE `jobresponse` (
  `idjobresponse` INT NULL AUTO_INCREMENT,
  `idjoblease` INT NOT NULL,
  `idjobid` INT NOT NULL,
  `time` DATETIME NOT NULL,
  `response` MEDIUMTEXT NOT NULL,
  `iserror` BIT NOT NULL,
  `error` MEDIUMTEXT NOT NULL,
  PRIMARY KEY (`idjobresponse`));
