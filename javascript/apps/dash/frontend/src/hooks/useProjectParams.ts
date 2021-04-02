import {useParams} from 'react-router';

import {ProjectRouteParams} from '@dash-frontend/lib/types';

const useProjectParams = () => {
  return useParams<ProjectRouteParams>();
};

export default useProjectParams;
