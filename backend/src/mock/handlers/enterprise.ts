import {EnterpriseIAPIServer} from '@pachyderm/node-pachyderm';

import MockState from './MockState';

const enterprise = () => {
  return {
    getService: (): Pick<EnterpriseIAPIServer, 'getState'> => {
      return {
        getState: (_call, callback) => {
          callback(null, MockState.state.enterprise);
        },
      };
    },
  };
};

export default enterprise();
