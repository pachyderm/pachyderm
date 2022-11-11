import {Status} from '@grpc/grpc-js/build/src/constants';

import {ProjectsIAPIServer} from '@dash-backend/proto';

import MockState from './MockState';

const projects = () => {
  return {
    getService: (): Pick<
      ProjectsIAPIServer,
      'inspectProject' | 'listProject'
    > => {
      return {
        inspectProject: (call, callback) => {
          if (MockState.state.error) {
            return callback(MockState.state.error);
          }

          const projectId = call.request.getProjectid();

          if (MockState.state.projects[projectId]) {
            callback(null, MockState.state.projects[projectId]);
          } else {
            callback({code: Status.NOT_FOUND, details: 'Project not found'});
          }
        },
        listProject: (_call, callback) => {
          if (MockState.state.error) {
            return callback(MockState.state.error);
          }
          return callback(null, MockState.state.projectInfo);
        },
      };
    },
  };
};

export default projects();
