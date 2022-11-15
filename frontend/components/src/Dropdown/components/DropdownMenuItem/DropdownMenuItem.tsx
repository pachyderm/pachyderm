import classnames from 'classnames';
import noop from 'lodash/noop';
import React, {ButtonHTMLAttributes, useRef} from 'react';

import {Group, Icon} from '@pachyderm/components';

import useDropdownMenuItem from '../../hooks/useDropdownMenuItem';

import styles from './DropdownMenuItem.module.css';

export interface DropdownMenuItemProps
  extends ButtonHTMLAttributes<HTMLButtonElement> {
  id: string;
  important?: boolean;
  closeOnClick?: boolean;
  value?: string;
  buttonStyle?: 'default' | 'tertiary';
  IconSVG?: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
}

export const DropdownMenuItem: React.FC<DropdownMenuItemProps> = ({
  children,
  important = false,
  id,
  className,
  onClick = noop,
  closeOnClick = false,
  value = '',
  buttonStyle = 'default',
  IconSVG,
  ...rest
}) => {
  const ref = useRef<HTMLButtonElement>(null);
  const {isSelected, handleClick, handleKeyDown, shown} = useDropdownMenuItem({
    id,
    onClick,
    closeOnClick,
    ref,
    value,
  });
  const classes = classnames(styles.base, className, {
    [styles.tertiary]: buttonStyle === 'tertiary',
    [styles.important]: important,
    [styles.selected]: isSelected,
  });

  if (!shown) {
    return null;
  }

  return (
    <button
      ref={ref}
      data-testid="DropdownMenuItem__button"
      role="menuitem"
      type="button"
      className={classes}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      tabIndex={-1}
      {...rest}
    >
      <Group spacing={IconSVG && 8} align="center">
        {IconSVG && (
          <Icon small color={buttonStyle !== 'tertiary' ? 'black' : 'white'}>
            <IconSVG />
          </Icon>
        )}
        {children}
      </Group>
    </button>
  );
};

export default DropdownMenuItem;
