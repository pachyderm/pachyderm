import http from 'k6/http';
import { check, sleep } from 'k6';
import { Options } from 'k6/options';
import {print} from 'graphql/language/printer';
import {GET_DAG_QUERY} from '@queries/GetDagQuery';
import { ASTNode } from 'graphql/language/ast';

const GRAPHQL_URL = 'http://localhost:3000/graphql';
const PACHD_ADDRESS = 'localhost:50051';
const TEST_CRITERIA = 'retrieved DAG successfully';


export let options: Options = {
  stages: [
    { duration: '5m', target: 100 }, // simulate ramp-up of traffic from 1 to 100 users over 5 minutes.
    { duration: '10m', target: 100 }, // stay at 100 users for 10 minutes
    { duration: '5m', target: 0 }, // ramp-down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(99)<1500'], // 99% of requests must complete below 1.5s
    [`${TEST_CRITERIA}`]: ['p(99)<1500'], // 99% of requests must complete below 1.5s
  },
};

export default () => {
  // This is likely due to a mismatch in types between load-tests's graphql-tag,
  // and the one apollo pulls in imperatively.
  const query = print(GET_DAG_QUERY as ASTNode);

  const options = {
    headers: {
      ['pachd-address']: PACHD_ADDRESS,
    },
  };

  const result = http.post(
    `${GRAPHQL_URL}`,
    JSON.stringify({ query, variables: { projectId: 'customerTeam' } }),
    options
  );

  check(result, {
    [`${TEST_CRITERIA}`]: (resp) => resp.json('dag') !== '',
  });

  sleep(1);
};
