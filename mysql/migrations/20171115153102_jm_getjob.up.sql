CREATE PROCEDURE `jm_getjob`(lockedby varchar(256))
BEGIN
	START TRANSACTION;
		SET @lastID = LAST_INSERT_ID(0);

		INSERT INTO joblease (idjob, lockedby, `at`, `until`) 				
		SELECT 
			j.idjob,
			lockedby,
			utc_timestamp(),
			TIMESTAMPADD(HOUR, 1, utc_timestamp())
		FROM `job` j
			LEFT JOIN jobresponse jr ON jr.idjobid = j.idjob
			LEFT JOIN joblease jl ON jl.idjob = j.idjob
		WHERE
			jr.idjobid IS NULL AND
			j.when <= utc_timestamp() AND
			((jl.idjoblease IS NULL) OR (jl.until < utc_timestamp()))
		ORDER BY j.when ASC
		LIMIT 1;

		IF LAST_INSERT_ID() > 0 THEN
			SELECT 
				j.idjob, 
				j.idschedule, 
				j.`when`, 
				j.arn, 
				j.payload 
			FROM 
				`job` j
				INNER JOIN joblease jl ON j.idjob = jl.idjob
			WHERE 
				jl.idjoblease = LAST_INSERT_ID();
		END IF;
    COMMIT;
END