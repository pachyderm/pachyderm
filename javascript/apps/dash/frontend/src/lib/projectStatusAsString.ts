import {ProjectStatus} from '@graphqlTypes';

const projectStatusAsString = (
  status: ProjectStatus,
): keyof typeof ProjectStatus => {
  if (status.toString() === 'HEALTHY') return 'HEALTHY';
  else return 'UNHEALTHY';
};

export default projectStatusAsString;
