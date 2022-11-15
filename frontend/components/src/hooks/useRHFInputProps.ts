import React, {useCallback} from 'react';
import {UseFormRegisterReturn} from 'react-hook-form';

interface UseRHFInputPropsOpts<T extends Element> {
  onChange: React.ChangeEventHandler<T> | undefined;
  onBlur: React.FocusEventHandler<T> | undefined;
  registerOutput: UseFormRegisterReturn;
}

const useRHFInputProps = <T extends Element = HTMLInputElement>({
  onChange,
  onBlur,
  registerOutput,
}: UseRHFInputPropsOpts<T>) => {
  const {
    onChange: rhfOnChange,
    onBlur: rhfOnBlur,
    ...inputProps
  } = registerOutput;

  const handleChange = useCallback(
    (e: React.ChangeEvent<T>) => {
      rhfOnChange(e);
      onChange && onChange(e);
    },
    [onChange, rhfOnChange],
  );

  const handleBlur = useCallback(
    (e: React.FocusEvent<T>) => {
      rhfOnBlur(e);
      onBlur && onBlur(e);
    },
    [onBlur, rhfOnBlur],
  );

  return {
    handleChange,
    handleBlur,
    ...inputProps,
  };
};

export default useRHFInputProps;
