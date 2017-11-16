ALTER TABLE `job`
    ADD CONSTRAINT fk_job_idschedule 
    FOREIGN KEY (idschedule) REFERENCES schedule(idschedule);

ALTER TABLE jobresponse
    ADD CONSTRAINT fk_jobresponse_idjoblease 
    FOREIGN KEY (idjoblease) REFERENCES joblease(idjoblease);

ALTER TABLE jobresponse
    ADD CONSTRAINT fk_jobresponse_idjob
    FOREIGN KEY (idjobid) REFERENCES job(idjob);

ALTER TABLE crontab 
    ADD CONSTRAINT fk_crontab_idschedule 
    FOREIGN KEY (idschedule) REFERENCES schedule(idschedule);
