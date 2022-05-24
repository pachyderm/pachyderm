import React, {useCallback, useEffect} from 'react';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import ImageProcessing from './ImageProcessing';

export type TutorialProps = {
  onClose: () => void;
};

type TutorialMap = {
  [key: string]: React.FC<TutorialProps>;
};

const TUTORIALS: TutorialMap = {
  'image-processing': ImageProcessing,
};

const ProjectTutorial: React.FC = () => {
  const {projectId} = useUrlState();
  const {viewState, updateViewState} = useUrlQueryState();
  const [, setTutorialData] = useLocalProjectSettings({
    projectId: 'account-data',
    key: 'tutorial_id',
  });
  const [, setTutorialProgress] = useLocalProjectSettings({
    projectId,
    key: 'tutorial_progress',
  });
  const [id, setActiveTutorial] = useLocalProjectSettings({
    projectId,
    key: 'active_tutorial',
  });

  const onClose = useCallback(() => {
    updateViewState({tutorialId: undefined});
    setActiveTutorial(null);
    setTutorialData(null);
    setTutorialProgress(null);
  }, [
    updateViewState,
    setActiveTutorial,
    setTutorialData,
    setTutorialProgress,
  ]);

  useEffect(() => {
    if (viewState.tutorialId && viewState.tutorialId !== id) {
      setActiveTutorial(viewState.tutorialId);
    }
  }, [viewState.tutorialId, setActiveTutorial, id]);

  if (id) {
    const Tutorial = TUTORIALS[id];
    if (Tutorial) {
      return <Tutorial onClose={onClose} />;
    }
  }
  return null;
};

export default ProjectTutorial;
