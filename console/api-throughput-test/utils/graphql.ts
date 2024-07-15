import http, {RefinedResponse, request} from 'k6/http';
import EXCHANGE_CODE_MUTATION from '@mutations/ExchangeCode';
import {DocumentNode} from 'graphql/language/ast';
import {check, sleep} from 'k6';
import {print} from 'graphql';

import {GET_PROJECTS_QUERY} from '@queries/GetProjectsQuery';
import {GET_ACCOUNT_QUERY} from '@queries/GetAccountQuery';
import {GET_PROJECT_DETAILS_QUERY} from '@queries/GetProjectDetailsQuery';
import {GET_PROJECT_QUERY} from '@queries/GetProjectQuery';
import {GET_REPOS_QUERY} from '@queries/GetReposQuery';
import {GET_REPO_QUERY} from '@queries/GetRepoQuery';
import {JOB_SET_QUERY} from '@queries/GetJobsetQuery';
import {JOBS_QUERY} from '@queries/GetJobsQuery';
import {JOB_SETS_QUERY} from '@queries/GetJobSetsQuery';
import {JOB_QUERY} from '@queries/GetJobQuery';
import {GET_COMMITS_QUERY} from '@queries/GetCommitsQuery';
import {GET_PIPELINE_QUERY} from '@queries/GetPipelineQuery';
import {GET_FILES_QUERY} from '@queries/GetFilesQuery';

const IDP_URL = 'https://hub-e2e-testing.us.auth0.com';
const DEX_URL = 'http://localhost:4000';
const URL = 'http://localhost:3000';

const JOB_ID = '4f89627c0d7447418cf46fb8f463f26f';
const PIPELINE = 'dag-0-pipeline-0';
const REPO = 'repo-0';
const PROJECT = 'default';

export enum Operation {
  getAccount = 'getAccount',
  projects = 'projects',
  project = 'project',
  projectDetails = 'projectDetails',
  jobSets = 'jobSets',
  jobSet = 'jobSet',
  jobs = 'jobs',
  job = 'job',
  repos = 'repos',
  repo = 'repo',
  getCommits = 'getCommits',
  exchangeCode = 'exchangeCode',
  pipeline = 'pipeline',
  getFiles = 'getFiles',
}

type QueryVariables = {
  id?: string;
  args?: {};
};

type QueryParams = {
  query: DocumentNode;
  variables: QueryVariables;
};

type QueryMap = {
  [key in Operation]: QueryParams;
};

const QUERIES: QueryMap = {
  [Operation.getAccount]: {
    query: GET_ACCOUNT_QUERY,
    variables: {},
  },
  [Operation.projects]: {
    query: GET_PROJECTS_QUERY,
    variables: {},
  },
  [Operation.project]: {
    query: GET_PROJECT_QUERY,
    variables: {id: PROJECT},
  },
  [Operation.projectDetails]: {
    query: GET_PROJECT_DETAILS_QUERY,
    variables: {args: {projectId: PROJECT}},
  },
  [Operation.jobSets]: {
    query: JOB_SETS_QUERY,
    variables: {
      args: {projectId: PROJECT},
    },
  },
  [Operation.jobSet]: {
    query: JOB_SET_QUERY,
    variables: {
      args: {
        id: JOB_ID,
        projectId: PROJECT,
      },
    },
  },
  [Operation.job]: {
    query: JOB_QUERY,
    variables: {
      args: {
        id: JOB_ID,
        pipelineName: PIPELINE,
        projectId: PROJECT,
      },
    },
  },
  [Operation.repos]: {
    query: GET_REPOS_QUERY,
    variables: {
      args: {projectId: PROJECT},
    },
  },
  [Operation.repo]: {
    query: GET_REPO_QUERY,
    variables: {
      args: {projectId: PROJECT, id: REPO},
    },
  },
  [Operation.getCommits]: {
    query: GET_COMMITS_QUERY,
    variables: {
      args: {
        projectId: PROJECT,
        repoName: REPO,
        branchName: 'master',
        number: 100,
      },
    },
  },
  [Operation.exchangeCode]: {
    query: EXCHANGE_CODE_MUTATION,
    variables: {},
  },
  [Operation.pipeline]: {
    query: GET_PIPELINE_QUERY,
    variables: {
      args: {
        id: PIPELINE,
        projectId: PROJECT,
      },
    },
  },
  [Operation.jobs]: {
    query: JOBS_QUERY,
    variables: {
      args: {
        pipelineId: PIPELINE,
        projectId: PROJECT,
      },
    },
  },
  [Operation.getFiles]: {
    query: GET_FILES_QUERY,
    variables: {
      args: {
        branchName: 'master',
        commitId: '^',
        path: '/',
        projectId: PROJECT,
        repoName: REPO,
      },
    },
  },
};

const logObject = (object: any) => {
  console.log(JSON.stringify(object, null, 4));
};

type ParamsObject = {
  [key: string]: string;
};
const buildParams = (object: ParamsObject) => {
  return Object.keys(object)
    .map((key) => key + '=' + object[key])
    .join('&');
};

const parseParams = (url: string) => {
  const paramString = url.split('?')[1];
  return paramString.split('&').reduce((memo: ParamsObject, param) => {
    const [key, value] = param.split('=');
    memo[decodeURIComponent(key)] = decodeURIComponent(value);
    return memo;
  }, {});
};

export type Authentication = {
  pachToken: string;
  idToken: string;
};

export const authenticate = (): Authentication => {
  if (
    !__ENV.PACHYDERM_AUTH_CLIENT_ID ||
    !__ENV.PACHYDERM_AUTH_CLIENT_SECRET ||
    !__ENV.PACHYDERM_AUTH_EMAIL ||
    !__ENV.PACHYDERM_AUTH_PASSWORD
  ) {
    throw 'environment credentials are not configured';
  }
  try {
    const url = `${DEX_URL}/auth?${buildParams({
      client_id: 'dash',
      redirect_uri: 'http://localhost:4000/oauth/callback/?inline=true',
      response_type: 'code',
      scope: 'openid+email+profile+groups+audience:server:client_id:pachd',
    })}`;
    const loginPage = http.get(url);

    const loginPageParams = parseParams(loginPage.url);
    const loginBody = loginPageParams;
    loginBody.client_id = loginBody.client;
    loginBody.connection = 'Username-Password-Authentication';
    loginBody.username = __ENV.PACHYDERM_AUTH_EMAIL;
    loginBody.password = __ENV.PACHYDERM_AUTH_PASSWORD;
    loginBody.tenant = 'hub-e2e-testing';

    const loginResponse = http.post(
      `${IDP_URL}/usernamepassword/login`,
      JSON.stringify(loginBody),
      {
        headers: {
          referer: loginPage.url,
          origin: IDP_URL,
          'content-type': 'application/json',
          cookie: loginPage.headers['Set-Cookie'],
        },
      },
    );
    const formResponse = loginResponse.submitForm();
    const code = parseParams(formResponse.url).code;

    const auth = exchangeCode(code);

    return auth;
  } catch (e) {
    throw e;
  }
};

type ExchangeCodeData = {
  exchangeCode: Authentication;
};
type ExchangeCodeResponse = {
  data: ExchangeCodeData;
};

const exchangeCode = (code: string): Authentication => {
  const res = makeGraphqlRequest(
    Operation.exchangeCode,
    EXCHANGE_CODE_MUTATION,
    {
      code,
    },
  );
  const exchangeResponse = res.json();

  return (exchangeResponse as ExchangeCodeResponse)?.data?.exchangeCode;
};

export const createGraphqlClient = () => ({
  query: (operationName: Operation) => {
    return query(operationName);
  },
  pollRequest: (requests: () => void, interval: number, duration: number) => {
    return pollRequest(interval, duration / interval, requests);
  },
});

const query = (operationName: Operation): RefinedResponse<any> => {
  const {query, variables} = QUERIES[operationName];
  return makeGraphqlRequest(operationName, query, variables);
};

const makeGraphqlRequest = (
  operationName: Operation,
  query: DocumentNode,
  variables = {},
): RefinedResponse<any> => {
  const response = http.post(
    URL + '/graphql',
    JSON.stringify({
      operationName,
      variables,
      query: print(query),
    }),
    {
      headers: {
        'Content-Type': 'application/json',
        'auth-token': '',
        'id-token': '',
        cookie: `dashAuthToken=;`,
      },
    },
  );

  checkWasSuccessful(response);
  return response;
};

const pollRequest = (
  interval: number,
  iteration: number,
  request: () => void,
) => {
  if (iteration === 0) return;
  sleep(interval);
  request();
  pollRequest(interval, iteration - 1, request);
};

const checkWasSuccessful = (res: RefinedResponse<any>) => {
  check(res, {
    successfulResponse: (res) => {
      const bodyString = res.body
        ? typeof res.body === 'string'
          ? res.body
          : String.fromCharCode(...res.body)
        : '{}';
      const successful = res.status === 200 && !JSON.parse(bodyString).errors;
      if (!successful) {
        logObject(res);
      }
      return successful;
    },
  });
};
