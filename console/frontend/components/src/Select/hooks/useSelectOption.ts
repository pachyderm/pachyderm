import {useContext, useCallback, useState, useEffect, useMemo} from 'react';

import SelectContext from '../contexts/SelectContext';
import {OptionRef} from '../lib/types';

const useSelectOption = (value: string, displayValue: React.ReactNode) => {
  const {
    id,
    activeValue,
    activeOptionRef,
    setActiveIndex,
    setIsOpen,
    setOptions,
    options,
    selectOptionAtIndex,
    value: selectValue,
    setIgnoreBlur,
  } = useContext(SelectContext);
  const [isRegistered, setIsRegistered] = useState(false);

  const index = useMemo(
    () => options.findIndex((option) => option.value === value),
    [options, value],
  );

  const onClick = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();

      setActiveIndex(index);
      selectOptionAtIndex(index);
      setIsOpen(false);
    },
    [index, selectOptionAtIndex, setActiveIndex, setIsOpen],
  );

  const onMouseDown = useCallback(() => {
    setIgnoreBlur(true);
  }, [setIgnoreBlur]);

  useEffect(() => {
    if (!isRegistered) {
      setOptions((_options: OptionRef[]) => {
        return [..._options, {value, displayValue}];
      });
      setIsRegistered(true);
    }
  }, [displayValue, isRegistered, options.length, setOptions, value]);

  const isActive = useMemo(() => activeValue === value, [activeValue, value]);

  return {
    htmlId: `${id}-${value}`,
    isSelected: value === selectValue,
    isActive,
    activeOptionRef,
    onClick,
    onMouseDown,
  };
};

export default useSelectOption;
