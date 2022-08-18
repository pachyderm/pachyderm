import {AdminIAPIServer} from '@pachyderm/node-pachyderm';

import MockState from './MockState';

const admin = () => {
  return {
    getService: (): Pick<AdminIAPIServer, 'inspectCluster'> => {
      return {
        inspectCluster: (_call, callback) => {
          callback(null, MockState.state.admin);
        },
      };
    },
  };
};

export default admin();
