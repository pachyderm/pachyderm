import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {deriveObserversFromPlugins} from '../lib/deriverObserversFromPlugins';
import {FileClient, FileClientConstructorArgs} from '../lib/FileClient';

export class ModifyFile extends FileClient<Empty.AsObject> {
  constructor({credentialMetadata, plugins = []}: FileClientConstructorArgs) {
    super();
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
}
