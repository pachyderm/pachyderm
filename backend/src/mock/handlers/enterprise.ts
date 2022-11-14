import {EnterpriseIAPIServer} from '@dash-backend/proto';

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
