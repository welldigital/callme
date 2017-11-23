CREATE PROCEDURE `sm_getschedule`(lockedby varchar(256))
BEGIN
	START TRANSACTION;
		SET @lastID = LAST_INSERT_ID(0);

		INSERT INTO crontablease (idcrontab, lockedby, `at`, `until`) 				
		SELECT 
			ct.idcrontab,
			lockedby,
			utc_timestamp(),
			TIMESTAMPADD(HOUR, 1, utc_timestamp())
		FROM
			`crontab` ct
		 	INNER JOIN `schedule` sc ON sc.idschedule = ct.idschedule
			LEFT JOIN crontablease ctl ON ctl.idcrontab = ct.idcrontab
		WHERE
			ct.next <= utc_timestamp() AND
			sc.active = 1 AND
			(ctl.idcrontab IS NULL OR ctl.rescinded = 1 OR ctl.until < utc_timestamp())
		ORDER BY ct.next ASC
		LIMIT 1;

		IF LAST_INSERT_ID() > 0 THEN
			SELECT 
				ctl.idcrontablease,
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
				INNER JOIN crontablease ctl ON ctl.idcrontab = ct.idcrontab
			WHERE 
				ctl.idcrontablease = LAST_INSERT_ID();
		END IF;
    COMMIT;
END