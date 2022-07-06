import {useModal} from '@pachyderm/components';
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

  const {openModal: openExitSurvey, isOpen: isExitSurveyOpen} = useModal(false);

  const closeTutorial = useCallback(() => {
    updateViewState({tutorialId: undefined});
    setActiveTutorial(null);
  }, [setActiveTutorial, updateViewState]);

  return {
    isExitSurveyOpen,
    openExitSurvey,
    closeTutorial,
  };
};

export default useImageProcessing;
