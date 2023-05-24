SELECT
  $__timeGroup(ci_jobs.job_timestamp,'30m'),
  count(ci_jobs.job_id) AS result,
  action AS value
FROM test_results
JOIN ci_jobs ON test_results.workflow_id=ci_jobs.workflow_id AND test_results.job_id=ci_jobs.job_id
WHERE
  (action = 'pass' or action = 'skip' or action = 'fail') and not test_name = '' 
  and not ci_jobs.job_name = 'test-go'
GROUP BY ci_jobs.job_timestamp, test_results.action
ORDER BY 1,2
