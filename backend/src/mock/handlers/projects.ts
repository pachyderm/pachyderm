import {ServiceError} from '@grpc/grpc-js';
import {Status} from '@grpc/grpc-js/build/src/constants';
import {IAPIServer} from '@pachyderm/proto/pb/projects/projects_grpc_pb';

import {
  default as projectFixtures,
  projectInfo,
} from '@dash-backend/mock/fixtures/projects';

const defaultState: {error: ServiceError | null} = {
  error: null,
};

const projects = () => {
  let state = {...defaultState};

  return {
    getService: (): Pick<IAPIServer, 'inspectProject' | 'listProject'> => {
      return {
        inspectProject: (call, callback) => {
          if (state.error) {
            return callback(state.error);
          }

          const projectId = call.request.getProjectid();

          if (projectFixtures[projectId]) {
            callback(null, projectFixtures[projectId]);
          } else {
            callback({code: Status.NOT_FOUND, details: 'Project not found'});
          }
        },
        listProject: (_call, callback) => {
          if (state.error) {
            return callback(state.error);
          }
          return callback(null, projectInfo);
        },
      };
    },
    setError: (error: ServiceError | null) => {
      state.error = error;
    },
    resetState: () => {
      state = {...defaultState};
    },
  };
};

export default projects();
