CREATE TABLE `joblease` (
  `idjoblease` INT NULL AUTO_INCREMENT,
  `lockedby` VARCHAR(256) NOT NULL,
  `at` DATETIME NOT NULL,
  `until` DATETIME NOT NULL,
  PRIMARY KEY (`idjoblease`));
