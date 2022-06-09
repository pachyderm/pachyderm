import {useCallback} from 'react';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useOnCloseTutorial = ({clearProgress}: {clearProgress?: boolean}) => {
  const {projectId} = useUrlState();
  const {updateViewState} = useUrlQueryState();
  const [, setTutorialData] = useLocalProjectSettings({
    projectId: 'account-data',
    key: 'tutorial_id',
  });
  const [tutorialProgress, setTutorialProgress] = useLocalProjectSettings({
    projectId,
    key: 'tutorial_progress',
  });
  const [, setActiveTutorial] = useLocalProjectSettings({
    projectId,
    key: 'active_tutorial',
  });

  const onCloseTutorial = useCallback(
    (tutorialName: string) => {
      updateViewState({tutorialId: undefined});
      setActiveTutorial(null);
      setTutorialData(null);
      if (clearProgress) {
        setTutorialProgress({...tutorialProgress, [tutorialName]: null});
      } else {
        setTutorialProgress({
          ...tutorialProgress,
          [tutorialName]: {
            ...tutorialProgress[tutorialName],
            completed: true,
          },
        });
      }
    },
    [
      updateViewState,
      setActiveTutorial,
      setTutorialData,
      clearProgress,
      setTutorialProgress,
      tutorialProgress,
    ],
  );

  return onCloseTutorial;
};

export default useOnCloseTutorial;
