CREATE PROCEDURE `jm_completejob`(idjob INT, resp MEDIUMTEXT, iserror bit, errorstring MEDIUMTEXT)
BEGIN
	INSERT INTO jobresponse
			(idjob, `time`, response, iserror, `error`)
		VALUES
			(idjob, utc_timestamp(), resp, iserror, errorstring);
END