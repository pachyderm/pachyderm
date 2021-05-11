import {RefObject, useCallback, useEffect, useMemo} from 'react';
import {useFormContext} from 'react-hook-form';

import useDropdown from './useDropdown';
import useItemKeyController from './useItemKeyController';
import useSelectedId from './useSelectedId';

interface UseDropdownMenuItemOpts {
  id: string;
  onClick: () => void;
  closeOnClick: boolean;
  ref: RefObject<HTMLButtonElement>;
  value: string;
}

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
  const {handleKeyDown} = useItemKeyController({ref});

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
