CREATE PROCEDURE `sm_startjobandupdatecron`(idschedule int, idcrontab int, idcrontablease int, nextjob timestamp)
BEGIN
	START TRANSACTION;
		SET @lastID=LAST_INSERT_ID(0);

		INSERT INTO `job` (arn, payload, idschedule, `when`)
		SELECT 
			s.arn, 
			s.payload, 
			s.idschedule, 
			utc_timestamp() 
		FROM schedule s
		WHERE 
			s.idschedule=idschedule;

		SET @lastID=LAST_INSERT_ID();

		UPDATE crontab ct
		SET
			ct.previous=ct.next,
			ct.next=nextjob,
			ct.lastupdated=utc_timestamp()
		WHERE
			ct.idcrontab=idcrontab;

		UPDATE crontablease ctl
		SET 
			rescinded = 1
		WHERE 
			ctl.idcrontablease=idcrontablease;

		SELECT @lastID;
    COMMIT;
END