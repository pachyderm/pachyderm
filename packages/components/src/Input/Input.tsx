import classNames from 'classnames';
import React, {forwardRef, InputHTMLAttributes, useMemo} from 'react';
import {RegisterOptions} from 'react-hook-form';

import useClearableInput from '../hooks/useClearableInput';
import useFormField from '../hooks/useFormField';
import {ExitSVG} from '../Svg';

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

  return (
    <>
      <div className={styles.wrapper}>
        <input
          aria-invalid={hasError}
          aria-describedby={errorId}
          ref={(e: HTMLInputElement) => {
            register(e, validationOptions);
            if (typeof ref === 'function') {
              ref(e);
            } else if (ref) {
              ref.current = e;
            }
          }}
          className={classes}
          name={name}
          disabled={disabled}
          readOnly={readOnly}
          defaultValue={defaultValue}
          type={type}
          {...rest}
        />
        {showButton && (
          <button
            className={styles.button}
            onClick={handleButtonClick}
            aria-label={`Clear ${name} input`}
            type="button"
          >
            <ExitSVG aria-hidden width={10} height={10} />
          </button>
        )}
      </div>

      {showErrors && <ErrorComponent />}
    </>
  );
};

export default forwardRef(Input);
