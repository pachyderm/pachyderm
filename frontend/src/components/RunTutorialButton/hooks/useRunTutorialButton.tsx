import {useHistory} from 'react-router';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';

const useRunTutorialButton = (projectId?: string) => {
  const routerHistory = useHistory();
  const [activeTutorial, setActiveTutorial] = useLocalProjectSettings({
    projectId: projectId || 'default',
    key: 'active_tutorial',
  });
  const [tutorialProgress] = useLocalProjectSettings({
    projectId: projectId || 'default',
    key: 'tutorial_progress',
  });

  const startTutorial = () => {
    setActiveTutorial('image-processing');
    routerHistory.push(`/lineage/${projectId || 'default'}`);
  };

  return {
    activeTutorial,
    startTutorial,
    tutorialProgress: !!tutorialProgress,
  };
};

export default useRunTutorialButton;
