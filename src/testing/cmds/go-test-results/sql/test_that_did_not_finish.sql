select 
  $__timeGroup(cj.job_timestamp, '$tp') as "time",
  count(tr.test_name),
  cj.job_name
from ci_jobs cj 
JOIN test_results tr  ON cj.workflow_id=tr.workflow_id and cj.job_id=tr.job_id and cj.job_executor=tr.job_executor
where $__timeFilter(cj.job_timestamp) and 
not exists (
  select 1 --t1.job_id, t1.test_name, t1.workflow_id, t1.job_executor, t1."action" ,t1.id, t2.id, t2."action" 
  from test_results t2
  where 
  --not (tr."action" = 'fail' or tr."action" = 'pass' or tr."action" = 'skip') and
  (t2."action" = 'fail' or t2."action" = 'pass' or t2."action" = 'skip')
  and tr.id <> t2.id 
  and tr.job_id=t2.job_id and tr.test_name=t2.test_name and tr.workflow_id=t2.workflow_id and tr.job_executor=t2.job_executor
) and tr."action" IN ('cont','pause','run','start') 
and not test_name=''
GROUP BY time, cj.job_name
ORDER BY 1,2 