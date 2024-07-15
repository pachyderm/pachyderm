import sample from 'lodash/sample';

import {ADJECTIVES, PACHYDERMS} from './constants/names';

const getRandomWorkspaceName = () => {
  const adjective = sample(ADJECTIVES);
  const pachyderm = sample(PACHYDERMS);
  return `${adjective}-${pachyderm}`;
};

export default getRandomWorkspaceName;
