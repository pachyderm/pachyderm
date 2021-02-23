import {useContext, useCallback} from 'react';

import DropdownContext from '../contexts/DropdownContext';

const useDropdown = () => {
  const {isOpen, setIsOpen, ...rest} = useContext(DropdownContext);

  const toggleDropdown = useCallback(() => {
    setIsOpen((prevIsOpen: boolean) => !prevIsOpen);
  }, [setIsOpen]);

  const openDropdown = useCallback(() => {
    setIsOpen(true);
  }, [setIsOpen]);

  const closeDropdown = useCallback(() => {
    setIsOpen(false);
  }, [setIsOpen]);

  return {
    isOpen,
    toggleDropdown,
    openDropdown,
    closeDropdown,
    ...rest,
  };
};

export default useDropdown;
