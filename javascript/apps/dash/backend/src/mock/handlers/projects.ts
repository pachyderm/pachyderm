import {IAPIServer} from '@pachyderm/proto/pb/projects/projects_grpc_pb';

import {default as projectFixtures, projectInfo} from '@dash-backend/mock/fixtures/projects';

const projects: Pick<IAPIServer, 'inspectProject' | 'listProject'> = {
  inspectProject: (call, callback) => {
    const projectId = call.request.getProjectid();

    callback(null, projectFixtures[projectId]);
  },
  listProject: (_call, callback) => {
    return callback(null, projectInfo);
  },
};

export default projects;
