import {useCallback} from 'react';

import {Keys} from 'lib/types';

import useDropdown from './useDropdown';

const useDropdownButton = (ref: React.RefObject<HTMLButtonElement>) => {
  const {toggleDropdown, isOpen, openDropdown, closeDropdown} = useDropdown();

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
          (firstActiveItem as HTMLElement).focus();
        }
      }
    },
    [closeDropdown, isOpen, openDropdown, ref],
  );

  return {toggleDropdown, isOpen, handleKeyDown};
};

export default useDropdownButton;
