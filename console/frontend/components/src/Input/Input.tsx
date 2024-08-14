import classNames from 'classnames';
import React, {forwardRef, InputHTMLAttributes, useMemo} from 'react';
import {RegisterOptions} from 'react-hook-form';

import useRHFInputProps from '@pachyderm/components/hooks/useRHFInputProps';

import useClearableInput from '../hooks/useClearableInput';
import useFormField from '../hooks/useFormField';
import {CloseSVG} from '../Svg';

import styles from './Input.module.css';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  clearable?: boolean;
  validationOptions?: RegisterOptions;
  name: string;
  showErrors?: boolean;
}

const Input: React.ForwardRefRenderFunction<HTMLInputElement, InputProps> = (
  {
    className,
    clearable = false,
    defaultValue,
    disabled,
    name,
    readOnly,
    type,
    validationOptions = {},
    showErrors = true,
    onChange,
    onBlur,
    ...rest
  },
  ref,
) => {
  const {
    register,
    error: ErrorComponent,
    errorId,
    hasError,
    setValue,
    watch,
    isTouched,
  } = useFormField(name);

  const currentValue = watch(name, defaultValue);

  const [hasInput, handleButtonClick] = useClearableInput(
    name,
    setValue,
    currentValue,
  );

  const hidden = useMemo(() => type === 'hidden', [type]);
  const showButton = useMemo(() => {
    return (
      !hidden &&
      clearable &&
      !disabled &&
      !readOnly &&
      (isTouched ? hasInput : Boolean(defaultValue))
    );
  }, [
    clearable,
    disabled,
    hasInput,
    hidden,
    readOnly,
    isTouched,
    defaultValue,
  ]);

  const classes = classNames(styles.base, className, {
    [styles.clearable]: clearable,
  });

  const {
    ref: formRef,
    handleChange,
    handleBlur,
    ...inputProps
  } = useRHFInputProps({
    onChange,
    onBlur,
    registerOutput: register(name, validationOptions),
  });

  return (
    <>
      <div className={styles.wrapper}>
        <input
          aria-invalid={hasError}
          aria-describedby={errorId}
          ref={(e) => {
            formRef(e);
            if (typeof ref === 'function') {
              ref(e);
            } else if (ref) {
              ref.current = e;
            }
          }}
          className={classes}
          disabled={disabled}
          readOnly={readOnly}
          defaultValue={defaultValue}
          type={type}
          onChange={handleChange}
          onBlur={handleBlur}
          {...rest}
          {...inputProps}
        />
        {showButton && (
          <button
            className={styles.button}
            onClick={handleButtonClick}
            aria-label={`Clear ${name} input`}
            type="button"
          >
            <CloseSVG aria-hidden />
          </button>
        )}
      </div>

      {showErrors && <ErrorComponent />}
    </>
  );
};

export default forwardRef(Input);
