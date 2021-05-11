import {RefObject, useCallback} from 'react';

import findFocusableChild from 'Dropdown/utils/findFocusableChild';
import {Keys} from 'lib/types';

import useDropdown from './useDropdown';

const findNextActiveSibling = (
  currentElem: HTMLButtonElement | ChildNode | null | undefined,
) => {
  let nextSibling = findFocusableChild(currentElem?.nextSibling);

  // continue traversing siblings as long as the
  // focusable child of the next sibling is disabled
  while (nextSibling && (nextSibling as HTMLButtonElement).disabled) {
    nextSibling = nextSibling?.nextSibling as HTMLElement | null;
  }

  return nextSibling;
};

const findPreviousActiveSibling = (
  currentElem: HTMLButtonElement | ChildNode | null | undefined,
) => {
  let prevSibling = findFocusableChild(currentElem?.previousSibling);

  while (prevSibling && (prevSibling as HTMLButtonElement).disabled) {
    prevSibling = prevSibling?.previousSibling as HTMLElement | null;
  }

  return prevSibling;
};

interface UseItemKeyControllerOpts {
  ref: RefObject<HTMLElement>;
}

const useItemKeyController = ({ref}: UseItemKeyControllerOpts) => {
  const {closeDropdown} = useDropdown();

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      let nextElement;

      if (e.key === Keys.Down) {
        let nextSibling = findNextActiveSibling(ref.current);

        if (nextSibling) {
          nextElement = nextSibling;
        } else {
          const firstItem = ref.current?.parentElement?.firstChild;

          if (firstItem && !(firstItem as HTMLButtonElement).disabled) {
            nextElement = findFocusableChild(firstItem);
          } else {
            nextSibling = findNextActiveSibling(firstItem);

            if (nextSibling) {
              nextElement = nextSibling;
            }
          }
        }
      }

      if (e.key === Keys.Up) {
        let prevSibling = findPreviousActiveSibling(ref.current);

        if (prevSibling) {
          nextElement = prevSibling;
        } else {
          const lastItem = ref.current?.parentElement?.lastChild;

          if (lastItem && !(lastItem as HTMLButtonElement).disabled) {
            nextElement = findFocusableChild(lastItem);
          } else {
            prevSibling = findPreviousActiveSibling(lastItem);

            if (prevSibling) {
              nextElement = prevSibling;
            }
          }
        }
      }

      if (e.key === Keys.Escape) {
        closeDropdown();

        const menuButton = ref.current?.parentElement?.previousSibling;

        if (menuButton) {
          nextElement = menuButton;
        }
      }

      if (nextElement) {
        (nextElement as HTMLElement).focus();
      }
    },
    [closeDropdown, ref],
  );

  return {
    handleKeyDown,
  };
};

export default useItemKeyController;
