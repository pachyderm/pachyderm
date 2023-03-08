import {useHistory} from 'react-router';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {lineageRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem} from '@pachyderm/components';

const useRunsList = () => {
  const {projectId} = useUrlState();
  const {getNewSearchParamsAndGo} = useUrlQueryState();
  const browserHistory = useHistory();

  const globalIdRedirect = (runId: string) => {
    getNewSearchParamsAndGo({
      globalIdFilter: runId,
    });
    browserHistory.push(
      lineageRoute({
        projectId,
      }),
    );
  };

  const onOverflowMenuSelect = (runId: string) => (id: string) => {
    switch (id) {
      case 'apply-run':
        return globalIdRedirect(runId);
      default:
        return null;
    }
  };

  const iconItems: DropdownItem[] = [
    {
      id: 'apply-run',
      content: 'Apply Global ID and view in DAG',
      closeOnClick: true,
    },
  ];

  return {
    iconItems,
    onOverflowMenuSelect,
  };
};

export default useRunsList;
