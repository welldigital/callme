CREATE PROCEDURE `jm_deletejob`(idjob int)
BEGIN
	DELETE j.* FROM `job` j
		LEFT OUTER JOIN joblease jl ON jl.idjob = j.idjob
		LEFT OUTER JOIN jobresponse jr ON jr.idjob = j.idjob
		WHERE 
			j.idjob = idjob AND
			jl.idjob IS NULL AND
			jr.idjob IS NULL;
END