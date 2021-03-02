import {ChannelCredentials, Metadata} from '@grpc/grpc-js';
import {APIClient} from '@pachyderm/proto/pb/projects/projects_grpc_pb';
import {
  Project,
  ProjectRequest,
  Projects,
} from '@pachyderm/proto/pb/projects/projects_pb';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

const projects = (
  pachdAddress: string,
  channelCredentials: ChannelCredentials,
) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    inspectProject: (projectId = '') => {
      return new Promise<Project.AsObject>((resolve, reject) => {
        const metadata = new Metadata();

        const request = new ProjectRequest();
        request.setProjectid(projectId);

        client.inspectProject(request, metadata, (err, res) => {
          if (err) return reject(err);

          return resolve(res.toObject());
        });
      });
    },
    listProject: () => {
      return new Promise<Projects.AsObject>((resolve, reject) => {
        const metadata = new Metadata();
        const empty = new Empty();

        client.listProject(empty, metadata, (err, res) => {
          if (err) return reject(err);

          return resolve(res.toObject());
        });
      });
    },
  };
};

export default projects;
