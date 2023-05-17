SELECT
  $__timeGroup(ci_jobs.job_timestamp,'30m'),
  count(test_results.test_name) AS metric,
  ci_jobs.branch AS value
FROM test_results
JOIN ci_jobs ON test_results.workflow_id=ci_jobs.workflow_id AND test_results.job_id=ci_jobs.job_id
WHERE
  action = 'fail' and not test_name = ''
GROUP BY ci_jobs.job_timestamp, ci_jobs.branch
ORDER BY 1,2
