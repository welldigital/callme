CREATE PROCEDURE `sm_getschedulebyid`(idschedule INT)
BEGIN
	SELECT 
		sc.`idschedule`, 
		sc.`externalid`, 
		sc.`by`,
		sc.`arn`,
		sc.`payload`,
		sc.`created`,
		sc.`active`,
		sc.`deactivateddate`,
		ct.`idcrontab`,
		ct.`idschedule`, 
		ct.`crontab`,
		ct.`previous`,
		ct.`next`,
		ct.`lastupdated`
	FROM 
		`crontab` ct
		INNER JOIN `schedule` sc ON sc.idschedule = ct.idschedule
	WHERE 
		sc.idschedule = idschedule;
END