import React, {useCallback} from 'react';
import {FieldError, useFormContext} from 'react-hook-form';

import {ErrorText} from '../Text';

const useFormField = (name: string) => {
  const {
    register,
    reset,
    watch,
    setValue,
    formState: {errors, touchedFields},
  } = useFormContext();

  const fieldError: FieldError | undefined = errors[name];
  const hasError = Boolean(fieldError);
  const hasErrorMessage = Boolean(hasError && fieldError?.message);
  const errorId = hasErrorMessage ? `${name}-error` : undefined;
  const isTouched = touchedFields[name];

  const error = useCallback(
    ({...rest}) => {
      if (!hasErrorMessage) {
        return null;
      }

      return (
        <ErrorText
          id={errorId}
          {...rest}
          role="alert"
          style={{marginLeft: '1rem', lineHeight: '2.5rem'}}
        >
          {fieldError?.message}
        </ErrorText>
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
