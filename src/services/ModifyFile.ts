import {APIClient} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {
  AddFile,
  Commit,
  ModifyFileRequest,
} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';

import {commitFromObject} from 'builders/pfs';
import {GRPCPlugin, ServiceArgs} from 'lib/types';

import {GRPC_MAX_MESSAGE_LENGTH} from './constants/pfs';

interface ModifyFileConstructorArgs extends ServiceArgs {
  plugins: GRPCPlugin[];
}

export class ModifyFile {
  private client: APIClient;
  private promise: Promise<Empty.AsObject>;
  stream: ReturnType<APIClient['modifyFile']>;

  constructor({
    pachdAddress,
    channelCredentials,
    credentialMetadata,
    plugins = [],
  }: ModifyFileConstructorArgs) {
    this.client = new APIClient(pachdAddress, channelCredentials, {
      /* eslint-disable @typescript-eslint/naming-convention */
      'grpc.max_receive_message_length': GRPC_MAX_MESSAGE_LENGTH,
      'grpc.max_send_message_length': GRPC_MAX_MESSAGE_LENGTH,
      /* eslint-enable @typescript-eslint/naming-convention */
    });

    const onCallObservers = plugins.flatMap((p) =>
      p.onCall ? [p.onCall] : [],
    );
    const onCompleteObservers = plugins.flatMap((p) =>
      p.onCompleted ? [p.onCompleted] : [],
    );
    const onErrorObservers = plugins.flatMap((p) =>
      p.onError ? [p.onError] : [],
    );
    this.promise = new Promise<Empty.AsObject>((resolve, reject) => {
      onCallObservers.forEach((cb) => cb({requestName: 'modifyFile'}));
      this.stream = this.client.modifyFile(credentialMetadata, (err) => {
        if (err) {
          onErrorObservers.forEach((cb) =>
            cb({error: err, requestName: 'modifyFile'}),
          );
          reject(err);
        } else {
          onCompleteObservers.forEach((cb) => cb({requestName: 'modifyFile'}));
          return resolve({});
        }
      });
    });
  }

  setCommit(commit: Commit.AsObject) {
    this.stream.write(
      new ModifyFileRequest().setSetCommit(commitFromObject(commit)),
    );
    return this;
  }

  putFileFromBytes(path: string, bytes: Buffer) {
    const addFile = new AddFile()
      .setPath(path)
      .setRaw(new BytesValue().setValue(bytes));
    this.stream.write(new ModifyFileRequest().setAddFile(addFile));
    return this;
  }

  putFileFromURL(path: string, url: string) {
    const addFile = new AddFile()
      .setPath(path)
      .setUrl(new AddFile.URLSource().setUrl(url));
    this.stream.write(new ModifyFileRequest().setAddFile(addFile));
    return this;
  }

  end() {
    this.stream.end();
    return this.promise;
  }
}
