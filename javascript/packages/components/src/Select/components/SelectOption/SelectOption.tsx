import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import useSelectOption from '../../hooks/useSelectOption';

import styles from './SelectOption.module.css';

export interface SelectOptionProps extends HTMLAttributes<HTMLDivElement> {
  value: string;
}

const SelectOption: React.FC<SelectOptionProps> = ({value, children}) => {
  const {
    activeOptionRef,
    htmlId,
    isActive,
    isSelected,
    onClick,
    onMouseDown,
  } = useSelectOption(value, children);

  return (
    <div
      role="option"
      id={htmlId}
      aria-selected={isSelected}
      className={classnames(styles.base, {[styles.active]: isActive})}
      onClick={onClick}
      onMouseDown={onMouseDown}
      ref={(el) => {
        if (isActive && activeOptionRef) activeOptionRef.current = el;
      }}
    >
      {children}
    </div>
  );
};

export default SelectOption;
