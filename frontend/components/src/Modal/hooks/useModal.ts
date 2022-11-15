import {useState, useCallback} from 'react';

const useModal = (initializer: boolean | (() => boolean) = false) => {
  const [isOpen, setIsOpen] = useState<boolean>(initializer);

  const openModal = useCallback(() => {
    setIsOpen(true);
  }, [setIsOpen]);

  const closeModal = useCallback(() => {
    setIsOpen(false);
  }, [setIsOpen]);

  return {isOpen, openModal, closeModal};
};

export default useModal;
