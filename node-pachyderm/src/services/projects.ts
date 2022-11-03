import {Metadata} from '@grpc/grpc-js';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {ServiceArgs} from '../lib/types';
import {APIClient} from '../proto/projects/projects_grpc_pb';
import {Project, ProjectRequest, Projects} from '../proto/projects/projects_pb';

const projects = ({pachdAddress, channelCredentials}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    inspectProject: (projectId = '') => {
      return new Promise<Project.AsObject>((resolve, reject) => {
        const metadata = new Metadata();

        const request = new ProjectRequest();
        request.setProjectid(projectId);

        client.inspectProject(request, metadata, (error, res) => {
          if (error) {
            return reject(error);
          }

          return resolve(res.toObject());
        });
      });
    },
    listProject: () => {
      return new Promise<Projects.AsObject>((resolve, reject) => {
        const metadata = new Metadata();
        const empty = new Empty();

        client.listProject(empty, metadata, (error, res) => {
          if (error) {
            return reject(error);
          }

          return resolve(res.toObject());
        });
      });
    },
  };
};

export default projects;
