import {EnterpriseIAPIServer} from '@pachyderm/node-pachyderm';

import enterpriseStates from '../fixtures/enterprise';

const enterprise = () => {
  return {
    getService: (): Pick<EnterpriseIAPIServer, 'getState'> => {
      return {
        getState: (_call, callback) => {
          callback(null, enterpriseStates.active);
        },
      };
    },
  };
};

export default enterprise();
