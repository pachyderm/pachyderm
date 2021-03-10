import {useCallback, useContext} from 'react';

import {Keys} from '../../lib/types';
import SelectContext from '../contexts/SelectContext';
import getActionFromKey from '../lib/getActionFromKey';
import getUpdatedIndex from '../lib/getUpdatedIndex';
import {MenuActions} from '../lib/types';

const useComboBox = () => {
  const {
    activeIndex,
    displayValue,
    id,
    ignoreBlur,
    isOpen,
    options,
    selectOptionAtIndex,
    setActiveIndex,
    setActiveValue,
    setIgnoreBlur,
    setIsOpen,
    activeValue,
    value,
  } = useContext(SelectContext);

  const onClick = useCallback(() => setIsOpen((prevState) => !prevState), [
    setIsOpen,
  ]);

  const onBlur = useCallback(() => {
    if (ignoreBlur) {
      setIgnoreBlur(false);
      return;
    }

    if (isOpen) {
      setActiveIndex(options.findIndex((option) => option.value === value));
      setActiveValue(value);
      setIsOpen(false);
    }
  }, [
    ignoreBlur,
    isOpen,
    options,
    setActiveIndex,
    setActiveValue,
    setIgnoreBlur,
    setIsOpen,
    value,
  ]);

  const onKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key !== Keys.Tab) {
        e.preventDefault();
      }

      const max = options.length - 1;

      const action = getActionFromKey(e, isOpen);

      switch (action) {
        case MenuActions.Next:
        case MenuActions.Last:
        case MenuActions.First:
        case MenuActions.Previous:
          setActiveIndex(getUpdatedIndex(max, action, activeIndex));
          setActiveValue(
            options[getUpdatedIndex(max, action, activeIndex) || 0].value,
          );
          break;
        case MenuActions.CloseSelect:
        case MenuActions.Space:
          if (activeIndex !== undefined) {
            selectOptionAtIndex(activeIndex);
            setActiveIndex(getUpdatedIndex(max, action, activeIndex));

            setActiveValue(
              options[getUpdatedIndex(max, action, activeIndex) || 0].value,
            );
          }
          break;
        case MenuActions.Close:
          setIsOpen(false);
          break;
        case MenuActions.Open:
          setIsOpen(true);
          setActiveValue(
            options[getUpdatedIndex(max, action, activeIndex) || 0].value,
          );
          break;
        default:
        // do nothing.
      }
    },
    [
      activeIndex,
      isOpen,
      options,
      selectOptionAtIndex,
      setActiveIndex,
      setActiveValue,
      setIsOpen,
    ],
  );

  return {
    displayValue,
    activeValue,
    id,
    isOpen,
    onBlur,
    onClick,
    onKeyDown,
    value,
  };
};

export default useComboBox;
