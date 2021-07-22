import {Account} from '@graphqlTypes';

const defaultAccount = {
  id: '1',
  email: 'cloud.strife@avalanche.net',
  name: 'Cloud Strife',
};

const accounts: {[accountId: string]: Account} = {
  default: defaultAccount,
  '1': defaultAccount,
  '2': {
    id: '2',
    email: 'barret.wallace@avalanche.net',
    name: 'Barret Wallace',
  },
};

export default accounts;
