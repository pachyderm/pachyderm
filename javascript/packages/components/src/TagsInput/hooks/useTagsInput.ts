import {useCallback, useEffect, useMemo, useState, useRef} from 'react';
import {useFormContext} from 'react-hook-form';

import {Keys} from '../../lib/types';

type UseTagsInputProps = {
  defaultValues: string[];
  name: string;
  validate: (value: string) => boolean;
};

export const useTagsInput = ({
  defaultValues,
  name,
  validate,
}: UseTagsInputProps) => {
  const {
    clearErrors,
    errors: formErrors,
    setError,
    setValue: setFormValue,
    register,
    formState,
  } = useFormContext();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [focused, setFocused] = useState(false);
  const [values, setValues] = useState<string[]>(defaultValues);
  const {isSubmitted} = formState;

  const errors = useMemo(() => {
    return values.map((v) => !validate(v));
  }, [validate, values]);
  const hasVisibleErrors = useMemo(
    () => (!values.length && isSubmitted) || errors.some((v) => v),
    [errors, values, isSubmitted],
  );

  const handleTagSave = useCallback(() => {
    const inputValue = inputRef.current?.value || '';
    if (!inputValue && values.length) return;

    const newValues = inputValue
      ? [
          ...values,
          ...inputValue.split(/[\s,]+/), // Support pasting CSV
        ]
      : values;
    setValues(newValues);
    setFormValue(name, newValues);

    if (inputRef.current) {
      inputRef.current.value = '';
    }
  }, [name, setFormValue, values]);

  const handleDelete = useCallback(
    (i: number) => {
      setValues((currValues) => {
        const newValues = [...currValues];
        newValues.splice(i, 1);
        setFormValue(name, newValues);

        return newValues;
      });
    },
    [name, setFormValue],
  );

  const handleKeyDown = useCallback(
    (evt: React.KeyboardEvent<HTMLInputElement>) => {
      if (
        evt.key === Keys.Enter ||
        evt.key === Keys.Comma ||
        evt.key === Keys.Space
      ) {
        evt.preventDefault();
        handleTagSave();
      }

      if (evt.key === Keys.Tab) {
        handleTagSave();
      }

      if (evt.key === 'Backspace' && (inputRef.current?.value || '') === '') {
        handleDelete(values.length - 1);
      }
    },
    [handleDelete, handleTagSave, values],
  );

  const handleBlur = useCallback(() => {
    setFocused(false);
    handleTagSave();
  }, [handleTagSave]);

  const handleFocus = useCallback(() => {
    setFocused(true);
  }, []);

  const handleReset = useCallback(() => {
    setValues([]);
    setFormValue(name, []);
  }, [name, setFormValue]);

  const ariaDescribedByIds = useMemo(
    () => values.map((v, i) => `tag-${v}-${i}`),
    [values],
  );

  const ariaDescribedBy = useMemo(() => ariaDescribedByIds.join(' '), [
    ariaDescribedByIds,
  ]);

  useEffect(() => {
    if (errors.some((error) => error)) {
      setError(name, {
        type: 'notMatch',
        message: 'Please remove invalid tags.',
      });
    } else if (formErrors[name]) {
      clearErrors(name);
    }
  }, [clearErrors, errors, formErrors, name, setError]);

  return {
    ariaDescribedBy,
    ariaDescribedByIds,
    errors,
    focused,
    handleBlur,
    handleDelete,
    handleFocus,
    handleKeyDown,
    hasVisibleErrors,
    register,
    values,
    inputRef,
    handleReset,
  };
};
