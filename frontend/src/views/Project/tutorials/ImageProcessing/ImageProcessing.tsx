import {BasicModal, TutorialModal} from '@pachyderm/components';
import React from 'react';

import {TutorialProps} from '../ProjectTutorial';

import ExitSurvey from './ExitSurvey';
import useImageProcessing from './hooks/useImageProcessing';
import stories from './stories';

const WARNING_MESSAGE =
  'Are you sure you want to end this tutorial? You will not be able to resume from where you left off.';

const ImageProcessing: React.FC<TutorialProps> = ({onClose}) => {
  const {
    isExitSurveyOpen,
    openExitSurvey,
    openConfirmationModal,
    closeTutorial,
    isConfirmationModalOpen,
    closeConfirmationModal,
    handleSkipTutorial,
    initialProgress,
  } = useImageProcessing();

  return (
    <>
      <BasicModal
        show={isConfirmationModalOpen}
        headerContent={'End Tutorial'}
        actionable
        loading={false}
        small
        onHide={closeConfirmationModal}
        onConfirm={handleSkipTutorial}
        confirmText="Finish Tutorial"
      >
        {WARNING_MESSAGE}
      </BasicModal>
      {isExitSurveyOpen ? <ExitSurvey onTutorialClose={onClose} /> : null}
      <TutorialModal
        onTutorialComplete={openExitSurvey}
        tutorialName="image-processing"
        stories={stories}
        onSkip={openConfirmationModal}
        onClose={closeTutorial}
        initialStory={initialProgress.initialStory}
        initialTask={initialProgress.initialTask}
      />
    </>
  );
};

export default ImageProcessing;
