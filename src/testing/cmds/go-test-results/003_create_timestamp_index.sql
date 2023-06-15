UPDATE public.ci_jobs SET job_timestamp = job_timestamp + interval '7 hours';
ALTER TABLE public.ci_jobs ALTER job_timestamp TYPE timestamptz USING job_timestamp AT TIME ZONE 'UTC';
CREATE INDEX ci_job_timestamp_index ON public.ci_jobs(job_timestamp);