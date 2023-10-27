import {useHistory} from 'react-router';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {lineageRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem} from '@pachyderm/components';

const useRunsList = () => {
  const {projectId} = useUrlState();
  const {getUpdatedSearchParams} = useUrlQueryState();
  const browserHistory = useHistory();

  const globalIdRedirect = (runId: string) => {
    const searchParams = getUpdatedSearchParams(
      {
        globalIdFilter: runId,
      },
      true,
    );

    return browserHistory.push(
      `${lineageRoute(
        {
          projectId,
        },
        false,
      )}?${searchParams}`,
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
