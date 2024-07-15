import {useModal} from '@pachyderm/components';

const useRerunPipelineButton = () => {
  const {openModal, closeModal, isOpen} = useModal(false);

  return {
    openModal,
    closeModal,
    isOpen,
  };
};

export default useRerunPipelineButton;
