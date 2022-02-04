import React, {useEffect} from 'react';

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

  const [id, setActiveTutorial] = useLocalProjectSettings({
    projectId,
    key: 'active_tutorial',
  });

  useEffect(() => {
    if (viewState.tutorialId && viewState.tutorialId !== id) {
      setActiveTutorial(viewState.tutorialId);
    }
  }, [viewState.tutorialId, setActiveTutorial, id]);

  const onClose = () => {
    updateViewState({tutorialId: undefined});
    setActiveTutorial(null);
  };

  if (id) {
    const Tutorial = TUTORIALS[id];
    if (Tutorial) {
      return <Tutorial onClose={onClose} />;
    }
  }
  return null;
};

export default ProjectTutorial;
