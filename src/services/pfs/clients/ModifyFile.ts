import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {Commit, ModifyFileRequest} from '../../..';
import {
  branchFromObject,
  BranchObject,
  commitFromObject,
} from '../../../builders/pfs';
import {deriveObserversFromPlugins} from '../lib/deriverObserversFromPlugins';
import {FileClient, FileClientConstructorArgs} from '../lib/FileClient';
export class ModifyFile extends FileClient<Empty.AsObject> {
  commit: Commit.AsObject | undefined;
  constructor({
    pachdAddress,
    channelCredentials,
    credentialMetadata,
    plugins = [],
  }: FileClientConstructorArgs) {
    super({
      pachdAddress,
      channelCredentials,
      credentialMetadata,
      plugins,
    });
    const {onCallObservers, onCompleteObservers, onErrorObservers} =
      deriveObserversFromPlugins(plugins);

    this.promise = new Promise<Empty.AsObject>((resolve, reject) => {
      onCallObservers.forEach((cb) => cb({requestName: 'modifyFile'}));
      this.stream = this.client.modifyFile(credentialMetadata, (err) => {
        if (err) {
          reject(err);
          onErrorObservers.forEach((cb) =>
            cb({error: err, requestName: 'modifyFile'}),
          );
          return;
        } else {
          resolve({});
          onCompleteObservers.forEach((cb) => cb({requestName: 'modifyFile'}));
          return;
        }
      });
    });
  }

  autoCommit(branch: BranchObject) {
    this.commit = new Commit().setBranch(branchFromObject(branch)).toObject();
    this.stream.write(
      new ModifyFileRequest().setSetCommit(commitFromObject(this.commit)),
    );
    return this;
  }

  setCommit(commit: Commit.AsObject) {
    this.commit = commit;
    this.stream.write(
      new ModifyFileRequest().setSetCommit(commitFromObject(commit)),
    );
    return this;
  }
}
