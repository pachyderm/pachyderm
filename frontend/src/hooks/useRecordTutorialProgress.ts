import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useRecordTutorialProgress = (
  tutorialName: string,
  story: number,
  task: number,
) => {
  const {projectId} = useUrlState();
  const [tutorialId] = useLocalProjectSettings({
    projectId: 'account-data',
    key: 'tutorial_id',
  });
  const [tutorialProgress, setTutorialsProgress] = useLocalProjectSettings({
    projectId,
    key: 'tutorial_progress',
  });

  return () => {
    setTutorialsProgress({
      ...tutorialProgress,
      [tutorialName]: {story, task, tutorialId},
    });
  };
};

export default useRecordTutorialProgress;
