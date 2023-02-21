import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem} from '@pachyderm/components';

const usePipelineStepsList = () => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();

  const onOverflowMenuSelect = (pipelineId: string) => (id: string) => {
    switch (id) {
      case 'dag':
        return browserHistory.push(pipelineRoute({projectId, pipelineId}));
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

export default usePipelineStepsList;
