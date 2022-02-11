import {useModal} from '@pachyderm/components';
import {useCallback} from 'react';

const useImageProcessing = () => {
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

  return {
    isExitSurveyOpen,
    openExitSurvey,
    openConfirmationModal,
    isConfirmationModalOpen,
    closeConfirmationModal,
    handleSkipTutorial,
  };
};

export default useImageProcessing;
