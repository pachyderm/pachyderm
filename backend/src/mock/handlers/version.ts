import {VersionIAPIServer} from '@dash-backend/proto';

import MockState from './MockState';

const version = () => {
  return {
    getService: (): Pick<VersionIAPIServer, 'getVersion'> => {
      return {
        getVersion: (_call, callback) => {
          callback(null, MockState.state.version);
        },
      };
    },
  };
};

export default version();
