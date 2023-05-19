SELECT
  test_results.test_name,
  test_results.elapsed_seconds,
  ci_jobs.job_name
FROM
  test_results
JOIN ci_jobs ON test_results.workflow_id=ci_jobs.workflow_id AND test_results.job_id=ci_jobs.job_id AND test_results.job_executor=ci_jobs.job_executor
WHERE
  $__timeFilter(ci_jobs.job_timestamp) and elapsed_seconds > '$min_test_duration' and not test_name = ''
GROUP BY test_results.test_name,ci_jobs.job_name,test_results.elapsed_seconds
ORDER BY elapsed_seconds
