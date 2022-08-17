import {useCallback} from 'react';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useImageProcessing = () => {
  const {projectId} = useUrlState();
  const {updateViewState} = useUrlQueryState();
  const [, setActiveTutorial] = useLocalProjectSettings({
    projectId,
    key: 'active_tutorial',
  });

  const closeTutorial = useCallback(() => {
    updateViewState({tutorialId: undefined});
    setActiveTutorial(null);
  }, [setActiveTutorial, updateViewState]);

  return {
    closeTutorial,
  };
};

export default useImageProcessing;
