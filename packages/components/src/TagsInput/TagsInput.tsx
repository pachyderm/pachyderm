import classNames from 'classnames';
import capitalize from 'lodash/capitalize';
import React, {InputHTMLAttributes} from 'react';

import {CloseSVG} from '../Svg';
import {FieldError} from '../Text';

import {useTagsInput} from './hooks/useTagsInput';
import styles from './TagsInput.module.css';

export interface TagsInputProps extends InputHTMLAttributes<HTMLInputElement> {
  'data-testid'?: string;
  defaultValues?: string[];
  name?: string;
  placeholder?: string;
  errorMessage?: string;
  validate?: (value: string) => boolean;
}

export const TagsInput: React.FC<TagsInputProps> = ({
  'data-testid': dataTestId = 'TagsInput__input',
  defaultValues = [],
  name = 'tagsInput',
  validate = () => true,
  placeholder,
  errorMessage,
  ...rest
}) => {
  const {
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
  } = useTagsInput({defaultValues, name, validate});

  return (
    <>
      <div
        className={classNames(
          styles.base,
          {
            [styles.errors]: hasVisibleErrors,
            [styles.focused]: focused,
          },
          {...rest},
        )}
        data-testid="TagsInput__container"
      >
        {values.map((value, i) => (
          <span
            className={classNames(styles.tag, {
              [styles.tagError]: errors[i],
            })}
            key={i}
            id={ariaDescribedByIds[i]}
          >
            {value}
            <CloseSVG
              aria-label={`Remove ${value}`}
              className={styles.close}
              onClick={() => handleDelete(i)}
            />
          </span>
        ))}
        <input
          aria-describedby={ariaDescribedBy}
          className={styles.input}
          data-testid={dataTestId}
          onBlur={handleBlur}
          onFocus={handleFocus}
          placeholder={values.length === 0 ? placeholder : undefined}
          onKeyDown={handleKeyDown}
          ref={inputRef}
        />
        <input type="hidden" name={name} ref={register()} />
      </div>
      <span className={styles.inline}>
        {values.length > 1 && (
          <button
            className={styles.clearButton}
            type="button"
            onClick={() => handleReset()}
          >
            {`Clear ${capitalize(name)}`}
          </button>
        )}
        {hasVisibleErrors && errorMessage && (
          <FieldError className={styles.inlineError}>{errorMessage}</FieldError>
        )}
      </span>
    </>
  );
};
