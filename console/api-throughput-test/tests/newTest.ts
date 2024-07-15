import {group} from 'k6';
import {
  createGraphqlClient,
  // sleep,
  Operation,
} from '../utils';

const DURATION = 10;
const VUS = 1;
// const SLEEP_DURATION_BETWEEN_ITERATIONS = 0.0;

function generateScenarios(
  operations: Operation[],
  duration: number,
  baseVUs: number,
) {
  let scenarios: Record<any, any> = {};
  let accumulatedDuration = 0;

  for (const operation in operations) {
    const operationName = operations[operation];
    const scenarioName = `${operationName}Test`;

    scenarios[operationName] = {
      executor: 'constant-vus',
      vus: baseVUs,
      duration: `${duration}s`,
      exec: scenarioName,
      startTime: `${accumulatedDuration}s`,
      gracefulStop: '5s',
    };

    accumulatedDuration += duration;
  }

  return scenarios;
}

// exchangeCode failing because I am not in auth? Not sure why jobSet was failing.
const excludedOperations = [Operation.jobSet, Operation.exchangeCode];

export let options = {
  scenarios: generateScenarios(
    Object.values(Operation)
      .sort()
      .filter((op) => !excludedOperations.includes(op)),
    DURATION,
    VUS,
  ),
};

const pp = (item: any) => {
  console.log(JSON.stringify(item, null, 2));
};

function runTest(op: Operation) {
  const {query} = createGraphqlClient();
  group(op, () => {
    query(op);
  });
  // sleep(SLEEP_DURATION_BETWEEN_ITERATIONS);
}

export function getAccountTest() {
  runTest(Operation.getAccount);
}

export function projectTest() {
  runTest(Operation.project);
}

export function projectsTest() {
  runTest(Operation.projects);
}

export function projectDetailsTest() {
  runTest(Operation.projectDetails);
}

export function jobSetsTest() {
  runTest(Operation.jobSets);
}

export function jobSetTest() {
  runTest(Operation.jobSet);
}

export function jobsTest() {
  runTest(Operation.jobs);
}

export function jobTest() {
  runTest(Operation.job);
}

export function repoTest() {
  runTest(Operation.repo);
}

export function reposTest() {
  runTest(Operation.repos);
}

export function getCommitsTest() {
  runTest(Operation.getCommits);
}

export function exchangeCodeTest() {
  runTest(Operation.exchangeCode);
}

export function pipelineTest() {
  runTest(Operation.pipeline);
}

export function getFilesTest() {
  runTest(Operation.getFiles);
}

const formatTable = (data: any) => {
  const totalNumberOfGroups = data.root_group.groups.length;
  const testRunDurationSec =
    data?.state?.testRunDurationMs / 1000 / totalNumberOfGroups; // Convert duration to seconds

  const groupResults = data?.root_group?.groups?.map(({name, checks}: any) => ({
    name,
    passes: checks?.[0]?.passes,
    fails: checks?.[0]?.fails,
    rps: (checks?.[0]?.passes + checks?.[0]?.fails) / testRunDurationSec,
  }));

  // Calculate the maximum length of the group names for formatting
  const maxNameLength = groupResults.reduce(
    (max: number, {name}: {name: string}) => Math.max(max, name.length),
    'Query'.length,
  );

  // Calculate the sum totals
  const totalPasses = groupResults.reduce(
    (total: number, {passes}: {passes: number}) => total + passes,
    0,
  );
  const totalFails = groupResults.reduce(
    (total: number, {fails}: {fails: number}) => total + fails,
    0,
  );
  const totalRPS =
    groupResults.reduce(
      (total: number, {rps}: {rps: number}) => total + rps,
      0,
    ) / totalNumberOfGroups;

  // Header row and separator
  const header = `Query${' '.repeat(
    maxNameLength - 5,
  )} | Passes | Failures | Aprox RPS`;
  const separator = '-'.repeat(header.length);

  // Create the table rows
  const rows = groupResults.map(
    ({
      name,
      passes,
      fails,
      rps,
    }: {
      name: string;
      passes: number;
      fails: number;
      rps: number;
    }) =>
      `${name.padEnd(maxNameLength)} | ${String(passes).padStart(6)} | ${String(
        fails,
      ).padStart(8)} | ${rps.toFixed(2).padStart(6)}`,
  );

  // Combine everything into a table string, including the totals
  return `${header}\n${separator}\n${rows.join(
    '\n',
  )}\n${separator}\n${'Total'.padEnd(maxNameLength)} | ${String(
    totalPasses,
  ).padStart(6)} | ${String(totalFails).padStart(8)} | ${totalRPS
    .toFixed(2)
    .padStart(6)}\n`;
};

export const handleSummary = (data: any) => {
  // console.log(JSON.stringify(data, null, 2));

  const summary = `
*Concurrent virtual users*: ${data?.metrics?.vus?.values?.value}

*Requests*: ${data?.metrics?.http_reqs?.values?.count} total with *Failures*: ${
    data?.metrics?.checks?.values?.fails
  } failures.

${formatTable(data)}

*Avg response time*: ${data?.metrics?.http_req_duration?.values?.avg?.toFixed(
    2,
  )}ms
*95% response time*: ${data?.metrics?.http_req_duration?.values[
    'p(95)'
  ].toFixed(2)}ms
  `;

  return {
    stdout: summary,
  };
};
