CREATE PROCEDURE `jm_startjob`(arn VARCHAR(2048), payload MEDIUMTEXT, idschedule INT, `when` DATETIME(6))
BEGIN
	INSERT `job` SET arn=arn, payload=payload, idschedule=idschedule, `when`=`when`;
	SELECT LAST_INSERT_ID() as idjob;
END