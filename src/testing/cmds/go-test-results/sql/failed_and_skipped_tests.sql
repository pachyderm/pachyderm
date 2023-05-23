SELECT
  test_results.test_name,
  action as result,
  count(test_results.action)
FROM
  test_results
JOIN ci_jobs ON test_results.workflow_id=ci_jobs.workflow_id AND test_results.job_id=ci_jobs.job_id
WHERE
  $__timeFilter(ci_jobs.job_timestamp) and not test_name = '' and (action = 'fail' or action = 'skip')
GROUP BY  test_results.test_name, action
