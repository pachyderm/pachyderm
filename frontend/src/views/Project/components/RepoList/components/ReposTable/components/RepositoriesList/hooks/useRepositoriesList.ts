import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem} from '@pachyderm/components';

const useRepositoriesList = () => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();

  const onOverflowMenuSelect = (repoId: string) => (id: string) => {
    switch (id) {
      case 'dag':
        return browserHistory.push(
          repoRoute({projectId, repoId, branchId: 'default'}),
        );
      default:
        return null;
    }
  };

  const iconItems: DropdownItem[] = [
    {
      id: 'dag',
      content: 'View in DAG',
      closeOnClick: true,
    },
  ];

  return {
    iconItems,
    onOverflowMenuSelect,
  };
};

export default useRepositoriesList;
