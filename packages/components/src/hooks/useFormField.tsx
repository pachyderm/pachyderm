import React, {useCallback} from 'react';
import {FieldError, useFormContext} from 'react-hook-form';

import {FieldError as FieldErrorComponent} from '../Text';

const useFormField = (name: string) => {
  const {
    register,
    errors,
    reset,
    watch,
    setValue,
    formState,
  } = useFormContext();

  const fieldError: FieldError | undefined = errors[name];
  const hasError = Boolean(fieldError);
  const hasErrorMessage = Boolean(hasError && fieldError?.message);
  const errorId = hasErrorMessage ? `${name}-error` : undefined;
  const isTouched = formState.touched[name];

  const error = useCallback(
    ({...rest}) => {
      if (!hasErrorMessage) {
        return null;
      }

      return (
        <FieldErrorComponent id={errorId} {...rest}>
          {fieldError?.message}
        </FieldErrorComponent>
      );
    },
    [errorId, fieldError, hasErrorMessage],
  );

  return {
    register,
    error,
    hasError,
    errorId,
    reset,
    setValue,
    watch,
    isTouched,
  };
};

export default useFormField;
