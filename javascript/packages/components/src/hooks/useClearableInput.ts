import {useCallback, useMemo} from 'react';

const useClearableInput = (
  name: string,
  setValue: (name: string, value: string) => void,
  currentValue: string,
): [boolean, (e: React.MouseEvent<Element, MouseEvent>) => void] => {
  const handleButtonClick = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      setValue(name, '');
      const input = e.currentTarget.previousSibling as HTMLElement;
      input.focus();
      input.blur(); // necessary to trigger a "touched" form state
      input.focus();
    },
    [name, setValue],
  );

  const hasInput = useMemo(() => {
    if (currentValue) return currentValue.length > 0;
    return false;
  }, [currentValue]);

  return [hasInput, handleButtonClick];
};

export default useClearableInput;
