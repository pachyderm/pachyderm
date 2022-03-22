import {match, useRouteMatch, useLocation} from 'react-router';

import {ConsoleRouteParams} from '@dash-frontend/lib/types';

const getDecodedRouteParam = (
  match: match<ConsoleRouteParams> | null,
  key: keyof ConsoleRouteParams,
) => {
  const param = match?.params[key];
  return param ? decodeURIComponent(param) : '';
};

const useLineageListViewSwap = () => {
  const PATH = '/:view/:projectId/';

  const location = useLocation();
  const match = useRouteMatch<ConsoleRouteParams>({
    path: PATH,
  });
  const view = getDecodedRouteParam(match, 'view');
  const projectId = getDecodedRouteParam(match, 'projectId');

  if (view === 'lineage') {
    return (
      location.pathname.replace(
        `/lineage/${projectId}`,
        `/project/${projectId}`,
      ) + location.search
    );
  }
  if (view === 'project') {
    return (
      location.pathname.replace(
        `/project/${projectId}`,
        `/lineage/${projectId}`,
      ) + location.search
    );
  }
  return `/${view}/${projectId}`;
};

export default useLineageListViewSwap;
