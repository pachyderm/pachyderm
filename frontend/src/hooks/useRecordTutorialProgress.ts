import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useRecordTutorialProgress = (
  tutorialName: string,
  story: number,
  task: number,
  action?: () => void,
) => {
  const {projectId} = useUrlState();
  const [, setTutorialsProgress] = useLocalProjectSettings({
    projectId,
    key: 'tutorial_progress',
  });

  return () => {
    action && action();
    setTutorialsProgress({
      [tutorialName]: {story, task},
    });
  };
};

export default useRecordTutorialProgress;
