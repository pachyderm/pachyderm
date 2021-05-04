import {RefObject, useCallback, useEffect, useMemo} from 'react';
import {useFormContext} from 'react-hook-form';

import {Keys} from 'lib/types';

import useDropdown from './useDropdown';
import useSelectedId from './useSelectedId';

interface UseDropdownMenuItemOpts {
  id: string;
  onClick: () => void;
  closeOnClick: boolean;
  ref: RefObject<HTMLButtonElement>;
  value: string;
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
  value,
}: UseDropdownMenuItemOpts) => {
  const {closeDropdown, filter, setFilteredResults} = useDropdown();
  const {selectedId, setSelectedId} = useSelectedId();
  const {watch} = useFormContext();

  const searchValue = watch('search');

  const shown = useMemo(() => filter({id, value}, searchValue), [
    searchValue,
    filter,
    id,
    value,
  ]);

  useEffect(() => {
    if (shown) {
      setFilteredResults((results) => {
        return [...results, {id, value}];
      });
    } else {
      setFilteredResults((results) => {
        return results.filter((result) => result.id !== id);
      });
    }
  }, [shown, id, setFilteredResults, value]);

  useEffect(() => {
    // We want to remove the unmounted item from the
    // filtered results. This is important in case the
    // Dropdown itself has dynamic items.

    return () => {
      setFilteredResults((results) => {
        return results.filter((result) => result.id !== id);
      });
    };
  }, [id, setFilteredResults]);

  const selectItem = useCallback(() => {
    setSelectedId(id);
  }, [id, setSelectedId]);

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

  return {
    selectItem,
    isSelected,
    closeDropdown,
    handleClick,
    handleKeyDown,
    shown,
  };
};

export default useDropdownMenuItem;
