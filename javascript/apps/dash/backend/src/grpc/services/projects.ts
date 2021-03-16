import {Metadata} from '@grpc/grpc-js';
import {APIClient} from '@pachyderm/proto/pb/projects/projects_grpc_pb';
import {
  Project,
  ProjectRequest,
  Projects,
} from '@pachyderm/proto/pb/projects/projects_pb';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {ServiceArgs} from '@dash-backend/lib/types';

const projects = ({pachdAddress, channelCredentials, log}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    inspectProject: (projectId = '') => {
      return new Promise<Project.AsObject>((resolve, reject) => {
        const metadata = new Metadata();

        const request = new ProjectRequest();
        request.setProjectid(projectId);

        log.info(
          {
            meta: {
              args: {projectId},
            },
          },
          'inspectProject request',
        );

        client.inspectProject(request, metadata, (error, res) => {
          if (error) {
            log.error(
              {
                error: error.message,
              },
              'inspectProject request failed',
            );
            return reject(error);
          }

          log.info('inspectProject request succeeded');
          return resolve(res.toObject());
        });
      });
    },
    listProject: () => {
      return new Promise<Projects.AsObject>((resolve, reject) => {
        const metadata = new Metadata();
        const empty = new Empty();

        log.info('listProject request');

        client.listProject(empty, metadata, (error, res) => {
          if (error) {
            log.error(
              {
                error: error.message,
              },
              'listProject request failed',
            );
            return reject(error);
          }

          log.info('listProject request succeeded');
          return resolve(res.toObject());
        });
      });
    },
  };
};

export default projects;
