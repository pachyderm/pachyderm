import http, {RefinedResponse, request} from 'k6/http';
import {check, sleep} from 'k6';

const IDP_URL = 'https://hub-e2e-testing.us.auth0.com';
const DEX_URL = 'http://localhost:4000';
const URL = 'http://localhost:80';

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

// type QueryMap = {
//   [key in Operation]: QueryParams;
// };

const JOB_ID = '4f89627c0d7447418cf46fb8f463f26f';
const PIPELINE = 'dag-0-pipeline-0';
const REPO = 'repo-0';
const PROJECT = 'default';

const QUERIES: Record<any, {url: string; variables: Record<any, any>}> = {
  // [Operation.getAccount]: {
  //   query: GET_ACCOUNT_QUERY,
  //   variables: {},
  // },

  [Operation.project]: {
    url: '/pfs_v2.API/InspectProject',
    variables: {project: {name: PROJECT}},
  },
  [Operation.projects]: {
    url: '/pfs_v2.API/ListProject',
    variables: {},
  },

  [Operation.jobSet]: {
    url: '/pps_v2.API/InspectJobSet',
    variables: {
      jobSet: {id: JOB_ID},
      details: true,
    },
  },
  [Operation.jobSets]: {
    url: '/pps_v2.API/ListJobSet',
    variables: {project: {name: PROJECT}},
  },

  [Operation.job]: {
    url: '/pps_v2.API/InspectJob',
    variables: {
      job: {
        id: JOB_ID,
        pipeline: {
          name: PIPELINE,
          project: {name: PROJECT},
        },
      },
    },
  },
  [Operation.jobs]: {
    url: '/pps_v2.API/ListJob',
    variables: {
      job: {
        pipeline: {
          name: PIPELINE,
          project: {name: PROJECT},
        },
      },
    },
  },

  [Operation.repo]: {
    url: '/pfs_v2.API/InspectRepo',
    variables: {
      repo: {
        name: REPO,
        type: 'user',
        project: {name: PROJECT},
      },
    },
  },
  [Operation.repos]: {
    url: '/pfs_v2.API/ListRepo',
    variables: {
      repo: {
        type: 'user',
        project: {name: PROJECT},
      },
    },
  },

  [Operation.getCommits]: {
    // This may not be a 1:1 to the previous
    url: '/pfs_v2.API/ListCommit',
    variables: {
      repo: {
        name: REPO,
        type: 'user',
        project: {name: PROJECT},
      },
      number: 100,
    },
  },
  // [Operation.exchangeCode]: {
  //   query: EXCHANGE_CODE_MUTATION,
  //   variables: {},
  // },
  [Operation.pipeline]: {
    url: '/pps_v2.API/InspectPipeline',
    variables: {
      details: true,
      pipeline: {
        name: PIPELINE,
        project: {
          name: PROJECT,
        },
      },
    },
  },
  [Operation.getFiles]: {
    url: '/pfs_v2.API/ListFile',
    variables: {
      file: {
        commit: {
          branch: {
            name: 'master',
            repo: {
              name: REPO,
              project: {
                name: PROJECT,
              },
              type: 'user',
            },
          },
          id: '^',
          repo: {
            name: REPO,
            project: {
              name: PROJECT,
            },
            type: 'user',
          },
        },
      },
    },
  },
};

const logObject = (object: any) => {
  console.log(JSON.stringify(object, null, 4));
};

export const createRestClient = () => ({
  query: (operationName: Operation) => {
    return query(operationName);
  },
});

const query = (operationName: Operation): RefinedResponse<any> => {
  const {url, variables} = QUERIES[operationName];
  return makeRestRequest(operationName, url, variables);
};

const makeRestRequest = (
  operationName: Operation,
  url: string,
  variables = {},
): RefinedResponse<any> => {
  const response = http.post(
    URL + '/api' + url,
    JSON.stringify({
      ...variables,
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

const checkWasSuccessful = (res: RefinedResponse<any>) => {
  check(res, {
    successfulResponse: (res) => {
      const successful = res.status < 400;
      if (!successful) {
        logObject(res);
      }
      return successful;
    },
  });
};
