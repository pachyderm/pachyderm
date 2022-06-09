import {useHistory} from 'react-router';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';

const useTutorialsMenu = (projectId?: string) => {
  const routerHistory = useHistory();
  const {updateViewState} = useUrlQueryState();
  const [, setTutorialData] = useLocalProjectSettings({
    projectId: 'account-data',
    key: 'tutorial_id',
  });
  const [activeTutorial, setActiveTutorial] = useLocalProjectSettings({
    projectId: projectId || 'default',
    key: 'active_tutorial',
  });
  const [tutorialProgress, setTutorialProgress] = useLocalProjectSettings({
    projectId: projectId || 'default',
    key: 'tutorial_progress',
  });

  const startTutorial = (tutorial: string) => {
    setActiveTutorial(tutorial);
    routerHistory.push(`/lineage/${projectId || 'default'}`);
  };

  const deleteTutorialResources = (tutorial: string) => {
    updateViewState({tutorialId: undefined});
    setTutorialProgress({...tutorialProgress, [tutorial]: null});
    setActiveTutorial(null);
    setTutorialData(null);
  };

  return {
    activeTutorial,
    startTutorial,
    tutorialProgress,
    deleteTutorialResources,
  };
};

export default useTutorialsMenu;
