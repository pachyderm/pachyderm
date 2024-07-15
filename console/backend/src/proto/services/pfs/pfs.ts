import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {
  branchFromObject,
  CommitObject,
  commitFromObject,
  commitSetFromObject,
  CommitSetObject,
} from '@dash-backend/proto/builders/pfs';
import {
  ServiceArgs,
  FinishCommitRequestArgs,
  StartCommitRequestArgs,
  RenewFileSetRequestArgs,
  AddFileSetRequestArgs,
} from '@dash-backend/proto/lib/types';
import {APIClient} from '@dash-backend/proto/proto/pfs/pfs_grpc_pb';
import {
  Commit,
  ClearCommitRequest,
  DropCommitSetRequest,
  StartCommitRequest,
  FinishCommitRequest,
  AddFileSetRequest,
  RenewFileSetRequest,
  ComposeFileSetRequest,
  Project,
} from '@dash-backend/proto/proto/pfs/pfs_pb';
import {grpcApiConstructorArgs} from '@dash-backend/proto/utils/createGrpcApiClient';

import {FileSet} from './clients/FileSet';
import {ModifyFile} from './clients/ModifyFile';
import {GRPC_MAX_MESSAGE_LENGTH} from './lib/constants';

let client: APIClient;

const pfsServiceRpcHandler = ({
  credentialMetadata,
  plugins = [],
}: ServiceArgs) => {
  client =
    client ??
    new APIClient(...grpcApiConstructorArgs(), {
      'grpc.max_receive_message_length': GRPC_MAX_MESSAGE_LENGTH,
      'grpc.max_send_message_length': GRPC_MAX_MESSAGE_LENGTH,
    });

  const pfsService = {
    startCommit: ({
      projectId,
      branch,
      parent,
      description = '',
    }: StartCommitRequestArgs & {projectId: string}) => {
      return new Promise<Commit.AsObject>((resolve, reject) => {
        const startCommitRequest = new StartCommitRequest();

        startCommitRequest.setBranch(branchFromObject(branch));
        startCommitRequest
          .getBranch()
          ?.getRepo()
          ?.setProject(new Project().setName(projectId));

        if (parent) {
          startCommitRequest.setParent(commitFromObject(parent));
          startCommitRequest
            .getParent()
            ?.getBranch()
            ?.getRepo()
            ?.setProject(new Project().setName(projectId));
        }
        startCommitRequest.setDescription(description);
        startCommitRequest.getBranch();

        client.startCommit(
          startCommitRequest,
          credentialMetadata,
          (error, res) => {
            if (error) {
              return reject(error);
            }
            return resolve(res.toObject());
          },
        );
      });
    },
    dropCommitSet: (commitSet: CommitSetObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const request = new DropCommitSetRequest();
        request.setCommitSet(commitSetFromObject(commitSet));
        client.dropCommitSet(request, credentialMetadata, (error) => {
          if (error) return reject(error);
          return resolve({});
        });
      });
    },
    finishCommit: ({
      projectId,
      error,
      force = false,
      commit,
      description = '',
    }: FinishCommitRequestArgs & {projectId: string}) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const finishCommitRequest = new FinishCommitRequest();

        if (error) {
          finishCommitRequest.setError(error);
        }
        if (commit) {
          finishCommitRequest.setCommit(commitFromObject(commit));
          finishCommitRequest
            .getCommit()
            ?.getBranch()
            ?.getRepo()
            ?.setProject(new Project().setName(projectId));
        }
        finishCommitRequest.setForce(force);
        finishCommitRequest.setDescription(description);

        client.finishCommit(
          finishCommitRequest,
          credentialMetadata,
          (error) => {
            if (error) {
              return reject(error);
            }
            return resolve({});
          },
        );
      });
    },
    clearCommit: (params: CommitObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const clearCommitRequest = new ClearCommitRequest().setCommit(
          commitFromObject(params),
        );

        client.clearCommit(clearCommitRequest, credentialMetadata, (error) => {
          if (error) {
            return reject(error);
          }
          return resolve({});
        });
      });
    },
    addFileSet: ({
      projectId,
      fileSetId,
      commit,
    }: AddFileSetRequestArgs & {projectId: string}) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const request = new AddFileSetRequest()
          .setCommit(commitFromObject(commit))
          .setFileSetId(fileSetId);

        request
          .getCommit()
          ?.getBranch()
          ?.getRepo()
          ?.setProject(new Project().setName(projectId));

        client.addFileSet(request, credentialMetadata, (err) => {
          if (err) reject(err);
          else resolve({});
        });
      });
    },
    renewFileSet: ({fileSetId, duration = 600}: RenewFileSetRequestArgs) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const request = new RenewFileSetRequest()
          .setFileSetId(fileSetId)
          .setTtlSeconds(duration);

        client.renewFileSet(request, credentialMetadata, (err) => {
          if (err) reject(err);
          else resolve({});
        });
      });
    },
    composeFileSet: (fileSets: string[]) => {
      return new Promise<string>((resolve, reject) => {
        const request = new ComposeFileSetRequest().setFileSetIdsList(fileSets);

        client.composeFileSet(request, credentialMetadata, (err, res) => {
          if (err) reject(err);
          else resolve(res.getFileSetId());
        });
      });
    },
    modifyFile: async () => {
      return new ModifyFile({
        credentialMetadata,
        plugins,
      });
    },
    fileSet: async () => {
      return new FileSet({
        credentialMetadata,
        plugins,
      });
    },
  };

  return pfsService;
};

export default pfsServiceRpcHandler;
