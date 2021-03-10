import {RefObject, useCallback} from 'react';

import {Keys} from 'lib/types';

import useDropdown from './useDropdown';

interface UseDropdownMenuItemOpts {
  id: string;
  onClick: () => void;
  closeOnClick: boolean;
  ref: RefObject<HTMLButtonElement>;
}

const findNextActiveSibling = (
  currentElem: HTMLButtonElement | ChildNode | null | undefined,
) => {
  let nextSibling = currentElem?.nextSibling;

  while (nextSibling && (nextSibling as HTMLButtonElement).disabled) {
    nextSibling = nextSibling?.nextSibling;
  }

  return nextSibling;
};

const findPreviousActiveSibling = (
  currentElem: HTMLButtonElement | ChildNode | null | undefined,
) => {
  let prevSibling = currentElem?.previousSibling;

  while (prevSibling && (prevSibling as HTMLButtonElement).disabled) {
    prevSibling = prevSibling?.previousSibling;
  }

  return prevSibling;
};

const useDropdownMenuItem = ({
  id,
  onClick,
  closeOnClick,
  ref,
}: UseDropdownMenuItemOpts) => {
  const {setSelectedId, onSelect, selectedId, closeDropdown} = useDropdown();

  const selectItem = useCallback(() => {
    setSelectedId(id);
    onSelect(id);
  }, [id, onSelect, setSelectedId]);

  const isSelected = selectedId === id;

  const handleClick = useCallback(() => {
    selectItem();
    onClick();

    if (closeOnClick) {
      closeDropdown();

      const menuButton = ref.current?.parentElement?.previousSibling;

      if (menuButton) {
        (menuButton as HTMLElement).focus();
      }
    }
  }, [closeDropdown, closeOnClick, onClick, ref, selectItem]);

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
            nextElement = firstItem;
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
            nextElement = lastItem;
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

  return {selectItem, isSelected, closeDropdown, handleClick, handleKeyDown};
};

export default useDropdownMenuItem;
