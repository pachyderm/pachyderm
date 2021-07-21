import {APIClient} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {
  CommitInfo,
  CommitSetInfo,
  FileInfo,
  GetFileRequest,
  InspectRepoRequest,
  ListCommitRequest,
  ListCommitSetRequest,
  ListFileRequest,
  ListRepoRequest,
  RepoInfo,
  SquashCommitSetRequest,
} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';
import {extract} from 'tar-stream';

import {
  commitSetFromObject,
  CommitSetObject,
  fileFromObject,
  FileObject,
  inspectCommitSetRequestFromObject,
  InspectCommitSetRequestObject,
  repoFromObject,
  RepoObject,
} from '../builders/pfs';
import {ServiceArgs} from '../lib/types';
import streamToObjectArray from '../utils/streamToObjectArray';

const pfs = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    listFile: (params: FileObject) => {
      const listFileRequest = new ListFileRequest();
      const file = fileFromObject(params);

      listFileRequest.setFile(file);
      listFileRequest.setDetails(true);

      const stream = client.listFile(listFileRequest, credentialMetadata);

      return streamToObjectArray<FileInfo, FileInfo.AsObject>(stream);
    },
    getFile: (params: FileObject) => {
      const getFileRequest = new GetFileRequest();
      const file = fileFromObject(params);

      getFileRequest.setFile(file);

      const stream = client.getFileTAR(getFileRequest, credentialMetadata);

      return new Promise<Buffer>((resolve, reject) => {
        // The GetFile request returns a tar stream.
        // We have to untar it before we can read it.
        const buffers: Buffer[] = [];
        const extractor = extract();

        extractor.on('entry', (_, estream, next) => {
          estream.on('data', (buffer: Buffer) => buffers.push(buffer));
          estream.on('end', () => next());
          estream.resume();
        });

        extractor.on('finish', () => {
          if (buffers.length) {
            return resolve(buffers[0]);
          } else {
            return reject(new Error('File does not exist.'));
          }
        });

        stream.on('data', (chunk: BytesValue) =>
          extractor.write(chunk.getValue()),
        );
        stream.on('end', () => extractor.end());
        stream.on('error', (err) => reject(err));
      });
    },
    listCommit: (
      repoName: RepoObject['name'],
      limit?: number,
      reverse = true,
    ) => {
      const listCommitRequest = new ListCommitRequest();
      const repo = repoFromObject({name: repoName});

      listCommitRequest.setRepo(repo);
      listCommitRequest.setReverse(reverse);

      if (limit) {
        listCommitRequest.setNumber(limit);
      }

      const stream = client.listCommit(listCommitRequest, credentialMetadata);

      return streamToObjectArray<CommitInfo, CommitInfo.AsObject>(stream);
    },
    inspectCommitSet: (request: InspectCommitSetRequestObject) => {
      const inspectCommitSetRequest =
        inspectCommitSetRequestFromObject(request);
      const stream = client.inspectCommitSet(
        inspectCommitSetRequest,
        credentialMetadata,
      );

      return streamToObjectArray<CommitInfo, CommitInfo.AsObject>(stream);
    },
    listCommitSet: () => {
      const listCommitSetRequest = new ListCommitSetRequest();
      const stream = client.listCommitSet(
        listCommitSetRequest,
        credentialMetadata,
      );

      return streamToObjectArray<CommitSetInfo, CommitSetInfo.AsObject>(stream);
    },
    squashCommitSet: (commitSet: CommitSetObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const squashCommitSetRequest =
          new SquashCommitSetRequest().setCommitSet(
            commitSetFromObject(commitSet),
          );

        client.squashCommitSet(
          squashCommitSetRequest,
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
    listRepo: (type = 'user') => {
      const listRepoRequest = new ListRepoRequest().setType(type);
      const stream = client.listRepo(listRepoRequest, credentialMetadata);

      return streamToObjectArray<RepoInfo, RepoInfo.AsObject>(stream);
    },
    inspectRepo: (name: RepoObject['name']) => {
      return new Promise<RepoInfo.AsObject>((resolve, reject) => {
        const inspectRepoRequest = new InspectRepoRequest();
        const repo = repoFromObject({name});

        inspectRepoRequest.setRepo(repo);

        client.inspectRepo(
          inspectRepoRequest,
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
  };
};

export default pfs;
