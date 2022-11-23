import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import {
  DefaultDropdown,
  DropdownItem,
  PureCheckbox,
  OverflowSVG,
} from '@pachyderm/components';

import DataCell from '../DataCell';

import styles from './Row.module.css';
export interface RowProps extends HTMLAttributes<HTMLTableRowElement> {
  active?: boolean;
  sticky?: boolean;
  isSelected?: boolean;
  hasCheckbox?: boolean;
  onClick?: () => void;
  overflowMenuItems?: DropdownItem[];
  dropdownOnSelect?: () => void;
}

const Row: React.FC<RowProps> = ({
  children,
  className,
  isSelected = false,
  hasCheckbox,
  overflowMenuItems,
  dropdownOnSelect,
  onClick,
  ...rest
}) => {
  return (
    <>
      {hasCheckbox && onClick && (
        <tr className={styles.checkboxRow}>
          <td>
            <PureCheckbox
              className={styles.checkbox}
              selected={isSelected}
              onChange={onClick}
            />
          </td>
        </tr>
      )}
      <tr
        className={classnames(styles.base, className, {
          [styles.active]: isSelected,
          [styles.button]: Boolean(onClick),
          [styles.hasCheckbox]: Boolean(hasCheckbox),
          [styles.selected]: isSelected,
          [styles.hasOverFlowMenu]: overflowMenuItems,
        })}
        role={onClick ? 'button' : undefined}
        onClick={onClick}
        {...rest}
      >
        {children}
        {overflowMenuItems && (
          <DataCell sticky isSelected={isSelected}>
            <DefaultDropdown
              items={overflowMenuItems}
              onSelect={dropdownOnSelect}
              storeSelected
              buttonOpts={{
                hideChevron: true,
                IconSVG: OverflowSVG,
                buttonType: 'ghost',
              }}
              menuOpts={{pin: 'right', className: styles.dropdown}}
            />
          </DataCell>
        )}
      </tr>
    </>
  );
};

export default Row;
