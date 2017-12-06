CREATE PROCEDURE `jm_getavailablejobcount`()
BEGIN
	SELECT COUNT(*) FROM `job` j
		LEFT JOIN `jobresponse` jr on jr.idjob = j.idjob
		WHERE
			jr.idjob IS NULL AND
			j.`when` <= utc_timestamp();
END