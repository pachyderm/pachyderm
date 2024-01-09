import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import {
  DefaultDropdown,
  DropdownItem,
  PureCheckbox,
  RadioButton,
  OverflowSVG,
  Icon,
  LockSVG,
  Tooltip,
} from '@pachyderm/components';

import DataCell from '../DataCell';

import styles from './Row.module.css';
export interface RowProps extends HTMLAttributes<HTMLTableRowElement> {
  active?: boolean;
  sticky?: boolean;
  isSelected?: boolean;
  hasCheckbox?: boolean;
  hasRadio?: boolean;
  hasLock?: boolean;
  lockedTooltipText?: string;
  onClick?: () => void;
  overflowMenuItems?: DropdownItem[];
  dropdownOnSelect?: (id: string) => void;
  openOnClick?: () => void;
}

const Row: React.FC<RowProps> = ({
  children,
  className,
  isSelected = false,
  hasCheckbox,
  hasRadio,
  hasLock,
  lockedTooltipText,
  overflowMenuItems,
  dropdownOnSelect,
  onClick,
  openOnClick,
  ...rest
}) => {
  return (
    <>
      <tr className={styles.checkboxRow}>
        <td>
          {hasCheckbox && onClick && (
            <PureCheckbox
              className={styles.checkbox}
              selected={isSelected}
              onChange={onClick}
            />
          )}
          {hasRadio && onClick && (
            <RadioButton.Pure
              className={styles.checkbox}
              selected={isSelected}
              onChange={onClick}
            />
          )}
          {hasLock && (
            <Tooltip tooltipText={lockedTooltipText}>
              <Icon color="grey" className={styles.checkbox}>
                <LockSVG />
              </Icon>
            </Tooltip>
          )}
        </td>
      </tr>
      <tr
        className={classnames(styles.base, className, {
          [styles.active]: isSelected,
          [styles.button]: Boolean(onClick),
          [styles.hasPrefixIcon]: Boolean(hasCheckbox || hasRadio || hasLock),
          [styles.selected]: isSelected,
          [styles.hasOverFlowMenu]: overflowMenuItems,
        })}
        role="button"
        onClick={onClick}
        {...rest}
      >
        {children}
        {overflowMenuItems && (
          <DataCell
            sticky
            isSelected={isSelected}
            className={styles.dropdownCell}
          >
            <DefaultDropdown
              openOnClick={openOnClick}
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
