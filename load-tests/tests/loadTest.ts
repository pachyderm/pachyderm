import {group, sleep} from 'k6';
import http from 'k6/http';

const DURATION = 90;
const POLL_INTERVAL = 3;

const POLL_DURATION = DURATION - 30;

import {
  authenticate,
  Authentication,
  createGraphqlClient,
  Operation,
} from '../utils';

export const options = {
  vus: 30,
  duration: `${DURATION}s`,
};

export default function () {
  const {query, pollRequest} = createGraphqlClient();
  const pages = [
    () => {
      group('landing', () => {
        query(Operation.getAccount);
        query(Operation.projects);

        pollRequest(
          () => {
            query(Operation.project);
          },
          POLL_INTERVAL,
          POLL_DURATION,
        );
      });
    },
    () => {
      group('lineage', () => {
        query(Operation.getAccount);
        query(Operation.projects);
        query(Operation.project);
        query(Operation.repos);

        pollRequest(
          () => {
            query(Operation.jobSets);
          },
          POLL_INTERVAL,
          POLL_DURATION,
        );
      });
    },
    () => {
      group('repo', () => {
        query(Operation.getAccount);
        query(Operation.project);
        query(Operation.repos);

        pollRequest(
          () => {
            query(Operation.repo);
            query(Operation.jobSets);
            query(Operation.getCommits);
          },
          POLL_INTERVAL,
          POLL_DURATION,
        );
      });
    },
    () => {
      group('pipeline', () => {
        query(Operation.getAccount);
        query(Operation.project);
        query(Operation.repos);
        query(Operation.pipeline);

        pollRequest(
          () => {
            query(Operation.jobs);
            query(Operation.jobSets);
          },
          POLL_INTERVAL,
          POLL_DURATION,
        );
      });
    },
    () => {
      group('files', () => {
        query(Operation.getAccount);
        query(Operation.project);
        query(Operation.repos);
        query(Operation.getFiles);

        pollRequest(
          () => {
            query(Operation.repo);
            query(Operation.jobSets);
            query(Operation.getCommits);
          },
          POLL_INTERVAL,
          POLL_DURATION,
        );
      });
    },
  ];

  const pageIndex = Math.floor(Math.random() * pages.length);
  pages[pageIndex]();
}

export const handleSummary = (data: any) => {
  const summary = `
*Concurrent virtual users*: ${data?.metrics?.vus?.values?.value}
*Requests*: ${data?.metrics?.http_reqs?.values?.count} 
${data?.root_group?.groups
  ?.map(({name, checks}: any) => {
    return `  ${name}: ${checks?.[0]?.passes}`;
  })
  .join(`\n`)}
*Failures*: ${data?.metrics?.checks?.values?.fails} ${
    data?.metrics?.checks?.values?.fails
      ? data?.root_group?.groups
          ?.map(({name, checks}: any) => {
            return `\n  ${name}: ${checks?.[0]?.fails}`;
          })
          .join(``)
      : ''
  }
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
