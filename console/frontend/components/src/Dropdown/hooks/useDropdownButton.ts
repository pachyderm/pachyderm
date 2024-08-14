import {useCallback} from 'react';

import {Keys} from '@pachyderm/components/lib/types';

import findFocusableChild from '../utils/findFocusableChild';

import useDropdown from './useDropdown';

const useDropdownButton = (ref: React.RefObject<HTMLButtonElement>) => {
  const {
    toggleDropdown,
    isOpen,
    openDropdown,
    closeDropdown,
    sideOpen,
    openUpwards,
  } = useDropdown();

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (!isOpen && e.key === Keys.Down) {
        openDropdown();
      }

      if (isOpen && (e.key === Keys.Up || e.key === Keys.Escape)) {
        closeDropdown();
      }

      if (isOpen && e.key === Keys.Down) {
        const menuItems = Array.from(
          ref.current?.nextSibling?.childNodes || [],
        );

        const firstActiveItem = menuItems.find((item) => {
          return !(item as HTMLButtonElement).disabled;
        });

        if (firstActiveItem) {
          findFocusableChild(firstActiveItem)?.focus();
        }
      }
    },
    [closeDropdown, isOpen, openDropdown, ref],
  );

  return {toggleDropdown, isOpen, handleKeyDown, sideOpen, openUpwards};
};

export default useDropdownButton;
