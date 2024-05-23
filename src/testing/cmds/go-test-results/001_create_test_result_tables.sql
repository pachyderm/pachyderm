CREATE TABLE public.ci_jobs (
	workflow_id varchar(64) NOT NULL,
	job_id varchar(64) NOT NULL,
	job_executor int NOT NULL,
	job_name varchar(1024) NOT NULL,
	job_timestamp timestamp NOT NULL,
	job_num_executors int NULL,
	commit_sha varchar(512) NULL,
	branch varchar(1024) NULL,
	tag varchar(1024) NULL,
	pull_requests varchar(2048) NULL,
    CONSTRAINT ci_jobs_pk PRIMARY KEY (workflow_id, job_id, job_executor)
);

CREATE INDEX ci_job_dk_index ON public.ci_jobs(workflow_id, job_id);

CREATE TABLE public.test_results (
	id varchar(64) NOT NULL, 
	workflow_id varchar(64) NOT NULL, 
	job_id varchar(64) NOT NULL, 
	job_executor int NOT NULL,
	test_name varchar(1024) NOT NULL,
	package varchar(2048) NULL,
	"action" varchar(64) NOT NULL,
	elapsed_seconds numeric NULL,
	"output" varchar NULL,
	CONSTRAINT test_results_pk PRIMARY KEY (id),
	CONSTRAINT test_results_fk FOREIGN KEY (workflow_id, job_id, job_executor) REFERENCES public.ci_jobs(workflow_id, job_id, job_executor)
);
