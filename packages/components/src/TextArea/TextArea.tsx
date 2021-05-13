import classNames from 'classnames';
import React, {useCallback, TextareaHTMLAttributes, useMemo} from 'react';
import {RegisterOptions} from 'react-hook-form';

import useRHFInputProps from 'hooks/useRHFInputProps';

import useClearableInput from '../hooks/useClearableInput';
import useFormField from '../hooks/useFormField';
import {ExitSVG} from '../Svg';

import styles from './TextArea.module.css';

interface TextAreaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  validationOptions?: RegisterOptions;
  autoExpand?: boolean;
  clearable?: boolean;
  name: string;
}

const expandInput = (element: HTMLTextAreaElement) => {
  element.style.height = 'inherit';

  const computed = window.getComputedStyle(element);

  const calculatedHeight =
    element.scrollHeight +
    parseInt(computed.getPropertyValue('border-bottom-width'), 10) * 2;

  const height = calculatedHeight > 46 ? calculatedHeight : 46; // 46 is the height of the Input component

  element.style.height = `${height}px`;
};

const TextArea: React.FC<TextAreaProps> = ({
  validationOptions = {},
  autoExpand = false,
  clearable = false,
  className,
  name,
  readOnly,
  disabled,
  onChange,
  onBlur,
  ...rest
}) => {
  const {
    register,
    error: ErrorComponent,
    hasError,
    errorId,
    setValue,
    watch,
  } = useFormField(name);

  const currentValue = watch(name);

  const [hasInput, handleButtonClick] = useClearableInput(
    name,
    setValue,
    currentValue,
  );

  const handleInput = useCallback(
    (e: React.FormEvent<HTMLTextAreaElement>) => {
      if (autoExpand) {
        expandInput(e.target as HTMLTextAreaElement);
      }
    },
    [autoExpand],
  );

  const rows = autoExpand ? 1 : undefined;

  const showButton = useMemo(() => {
    return clearable && !disabled && !readOnly && hasInput;
  }, [clearable, disabled, hasInput, readOnly]);

  const classes = classNames(styles.base, className, {
    [styles.clearable]: clearable,
    [styles.autoExpand]: autoExpand,
  });

  const {handleChange, handleBlur, ...inputProps} = useRHFInputProps({
    onChange,
    onBlur,
    registerOutput: register(name, validationOptions),
  });

  return (
    <>
      <div className={styles.wrapper}>
        <textarea
          aria-invalid={hasError}
          aria-describedby={errorId}
          className={classes}
          onInput={handleInput}
          rows={rows}
          readOnly={readOnly}
          disabled={disabled}
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
            <ExitSVG aria-hidden width={10} height={10} />
          </button>
        )}
      </div>

      <ErrorComponent />
    </>
  );
};

export default TextArea;
