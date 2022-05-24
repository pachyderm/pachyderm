import {useModal} from '@pachyderm/components';
import {useCallback, useMemo} from 'react';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import stories from '../stories';

const useImageProcessing = () => {
  const {projectId} = useUrlState();
  const {updateViewState} = useUrlQueryState();
  const [, setActiveTutorial] = useLocalProjectSettings({
    projectId,
    key: 'active_tutorial',
  });
  const [tutorialProgress] = useLocalProjectSettings({
    projectId,
    key: 'tutorial_progress',
  });

  const {openModal: openExitSurvey, isOpen: isExitSurveyOpen} = useModal(false);
  const {
    openModal: openConfirmationModal,
    isOpen: isConfirmationModalOpen,
    closeModal: closeConfirmationModal,
  } = useModal(false);

  const handleSkipTutorial = useCallback(() => {
    closeConfirmationModal();
    openExitSurvey();
  }, [closeConfirmationModal, openExitSurvey]);

  const closeTutorial = useCallback(() => {
    updateViewState({tutorialId: undefined});
    setActiveTutorial(null);
  }, [setActiveTutorial, updateViewState]);

  const initialProgress = useMemo(() => {
    const {story: currentStory, task: completedTask} = tutorialProgress
      ? tutorialProgress['image-processing']
      : {story: 0, task: -1};

    let initialTask = 0;
    let initialStory = 0;

    if (
      stories[currentStory].sections.filter((s) => !!s.Task).length ===
      completedTask + 1
    ) {
      if (currentStory + 1 <= stories.length - 1) {
        initialStory = currentStory + 1;
      } else {
        initialStory = currentStory;
        initialTask = completedTask;
      }
    } else {
      initialTask = completedTask + 1;
    }

    return {initialTask, initialStory};
  }, [tutorialProgress]);

  return {
    isExitSurveyOpen,
    openExitSurvey,
    openConfirmationModal,
    isConfirmationModalOpen,
    closeConfirmationModal,
    handleSkipTutorial,
    closeTutorial,
    initialProgress,
  };
};

export default useImageProcessing;
