import React, {useEffect} from 'react';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import OpenCvTutorial from './OpenCvTutorial';

type TutorialMap = {
  [key: string]: React.FC;
};

const TUTORIALS: TutorialMap = {
  'open-cv': OpenCvTutorial,
};

const ProjectTutorial: React.FC = () => {
  const {projectId} = useUrlState();
  const {viewState} = useUrlQueryState();

  const [id, setActiveTutorial] = useLocalProjectSettings({
    projectId,
    key: 'active_tutorial',
  });

  useEffect(() => {
    if (viewState.tutorialId && viewState.tutorialId !== id) {
      setActiveTutorial(viewState.tutorialId);
    }
  }, [viewState.tutorialId, setActiveTutorial, id]);

  if (id) {
    const Tutorial = TUTORIALS[id];
    if (Tutorial) {
      return <Tutorial />;
    }
  }
  return null;
};

export default ProjectTutorial;
