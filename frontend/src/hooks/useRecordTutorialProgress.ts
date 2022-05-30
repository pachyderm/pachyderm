import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useRecordTutorialProgress = (
  tutorialName: string,
  story: number,
  task: number,
) => {
  const {projectId} = useUrlState();
  const [, setTutorialsProgress] = useLocalProjectSettings({
    projectId,
    key: 'tutorial_progress',
  });

  return () => {
    setTutorialsProgress({
      [tutorialName]: {story, task},
    });
  };
};

export default useRecordTutorialProgress;
