CREATE PROCEDURE `jm_getjobresponse`(idjob int)
BEGIN
	SELECT
		j.idjob, j.idschedule, j.`when`, j.arn, j.payload, 
		jr.idjobresponse, jr.idjob, jr.`time`, jr.response, jr.iserror, jr.`error` 
		FROM `job` j 
		LEFT JOIN `jobresponse` jr ON jr.idjob = j.idjob 
		WHERE 
			j.idjob=idjob;
END