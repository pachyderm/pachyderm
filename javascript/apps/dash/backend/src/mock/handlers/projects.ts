import {ServiceError} from '@grpc/grpc-js';
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
          const projectId = call.request.getProjectid();

          callback(null, projectFixtures[projectId]);
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
