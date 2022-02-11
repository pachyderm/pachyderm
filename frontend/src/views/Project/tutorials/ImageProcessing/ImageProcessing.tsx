import {BasicModal, TutorialModal} from '@pachyderm/components';
import React from 'react';
import {Prompt} from 'react-router';

import {TutorialProps} from '../ProjectTutorial';

import ExitSurvey from './ExitSurvey';
import useImageProcessing from './hooks/useImageProcessing';
import stories from './stories';

const WARNING_MESSAGE =
  'Are you sure you want to skip this tutorial? You will not be able to resume from where you left off.';

const ImageProcessing: React.FC<TutorialProps> = ({onClose}) => {
  const {
    isExitSurveyOpen,
    openExitSurvey,
    openConfirmationModal,
    isConfirmationModalOpen,
    closeConfirmationModal,
    handleSkipTutorial,
  } = useImageProcessing();
  return (
    <>
      <BasicModal
        show={isConfirmationModalOpen}
        headerContent={'Skip Tutorial'}
        actionable
        loading={false}
        small
        onHide={closeConfirmationModal}
        onConfirm={handleSkipTutorial}
        confirmText="Skip"
      >
        {WARNING_MESSAGE}
      </BasicModal>
      {!isExitSurveyOpen && !isConfirmationModalOpen && (
        <Prompt message={WARNING_MESSAGE} />
      )}
      {isExitSurveyOpen ? <ExitSurvey onTutorialClose={onClose} /> : null}
      <TutorialModal
        onTutorialComplete={openExitSurvey}
        stories={stories}
        onSkip={openConfirmationModal}
      />
    </>
  );
};

export default ImageProcessing;
