import classnames from 'classnames';
import React, {forwardRef} from 'react';

import {Icon} from '../../../Icon';
import {ChevronDownSVG} from '../../../Svg';
import useComboBox from '../../hooks/useComboBox';

import styles from './ComboBox.module.css';

const ComboBox: React.ForwardRefRenderFunction<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
> = (props, ref) => {
  const {
    displayValue,
    activeValue,
    id,
    isOpen,
    onBlur,
    onClick,
    onKeyDown,
  } = useComboBox();

  return (
    <div
      role="combobox"
      aria-activedescendant={`${id}-${activeValue}`}
      aria-autocomplete="none"
      aria-haspopup="listbox"
      aria-expanded={isOpen}
      aria-labelledby={`${id} ${id}-value`}
      aria-controls={`${id}-listbox`}
      id={`${id}-value`}
      tabIndex={0}
      ref={ref}
      className={classnames(styles.base, {[styles.open]: isOpen})}
      onClick={onClick}
      onKeyDown={onKeyDown}
      onBlur={onBlur}
      {...props}
    >
      {displayValue}

      <Icon className={styles.icon} color="black" aria-hidden={true}>
        <ChevronDownSVG />
      </Icon>
    </div>
  );
};

export default forwardRef(ComboBox);
