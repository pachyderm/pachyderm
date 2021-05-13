import classnames from 'classnames';
import React, {useState, useMemo, useEffect, useRef, useCallback} from 'react';
import {
  useFormContext,
  RegisterOptions,
  CustomElement,
  FieldPath,
  FieldValues,
} from 'react-hook-form';

import ComboBox from './components/ComboBox';
import SelectOption from './components/SelectOption';
import SelectContext from './contexts/SelectContext';
import maintainScrollVisibility from './lib/maintainScrollVisibility';
import {OptionRef} from './lib/types';
import styles from './Select.module.css';

export interface SelectProps
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'placeholder'> {
  id: FieldPath<FieldValues>;
  validationOptions?: RegisterOptions;
  initialValue?: string;
  placeholder?: React.ReactNode;
}

const Select: React.FC<SelectProps> = ({
  id,
  placeholder = 'Select an option',
  initialValue = '',
  children,
  validationOptions = {},
  ...rest
}) => {
  const formContextValues = useFormContext();
  const [isOpen, setIsOpen] = useState(false);
  const [ignoreBlur, setIgnoreBlur] = useState(false);
  const [activeIndex, setActiveIndex] = useState<number>();
  const [options, setOptions] = useState<OptionRef[]>([]);
  const [value, setValue] = useState(initialValue);
  const [activeValue, setActiveValue] = useState(initialValue);
  const [displayValue, setDisplayValue] = useState(placeholder);

  const comboboxRef = useRef<HTMLDivElement | null>(null);
  const selectRef = useRef<CustomElement<Record<string, unknown>>>({
    name: id,
    value,
  });
  const listboxRef = useRef<HTMLDivElement | null>(null);
  const activeOptionRef = useRef<HTMLElement | null>(null);

  const {register, setValue: setFormValue} = formContextValues;

  const selectOptionAtIndex = useCallback(
    (index: number) => {
      const option = options[index];

      setValue(option.value);
      setDisplayValue(option.displayValue || null);
    },
    [options],
  );

  const ctxValue = useMemo(
    () => ({
      activeIndex,
      activeOptionRef,
      activeValue,
      displayValue,
      id,
      ignoreBlur,
      isOpen,
      options,
      selectOptionAtIndex,
      setActiveIndex,
      setActiveValue,
      setIgnoreBlur,
      setIsOpen,
      setOptions,
      value,
    }),
    [
      activeIndex,
      activeOptionRef,
      activeValue,
      displayValue,
      id,
      ignoreBlur,
      isOpen,
      options,
      selectOptionAtIndex,
      value,
    ],
  );

  const {ref: formRef} = register(id, validationOptions);

  useEffect(() => {
    setFormValue(id, value);
  }, [id, value, setFormValue]);

  useEffect(() => {
    selectRef.current = {
      ...selectRef.current,
      focus: () => comboboxRef.current?.focus(),
    };
  }, [comboboxRef]);

  useEffect(() => {
    // propogate initial value once option has registered
    if (value && options.length > 0 && displayValue === placeholder) {
      setDisplayValue(
        options.find((option) => option.value === value)?.displayValue || '',
      );
    }
  }, [value, options, displayValue, placeholder]);

  useEffect(() => {
    return formRef(selectRef.current);
  }, [validationOptions, formRef]);

  useEffect(() => {
    if (isOpen && activeValue) {
      maintainScrollVisibility(activeOptionRef.current, listboxRef.current);
    }
  }, [activeOptionRef, activeValue, isOpen]);

  return (
    <div className={classnames(styles.base, {[styles.open]: isOpen})}>
      <SelectContext.Provider value={ctxValue}>
        <ComboBox ref={comboboxRef} {...rest} />

        <div
          className={styles.listbox}
          role="listbox"
          id={`${id}-listbox`}
          ref={listboxRef}
        >
          {children}
        </div>
      </SelectContext.Provider>
    </div>
  );
};

export default Object.assign(Select, {Option: SelectOption});
