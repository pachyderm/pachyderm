SELECT
  test_results.test_name,
  test_results.elapsed_seconds,
  ci_jobs.job_name
FROM
  test_results
JOIN ci_jobs ON test_results.workflow_id=ci_jobs.workflow_id AND test_results.job_id=ci_jobs.job_id
WHERE
  $__timeFilter(ci_jobs.job_timestamp) and elapsed_seconds > ${min_elapsed_seconds}
GROUP BY test_results.test_name, elapsed_seconds, ci_jobs.job_name
ORDER BY elapsed_seconds
