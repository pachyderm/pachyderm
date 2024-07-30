import classnames from 'classnames';
import noop from 'lodash/noop';
import React, {ButtonHTMLAttributes, useRef} from 'react';

import {Group, Icon, Tooltip} from '@pachyderm/components';

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
  topBorder?: boolean;
  tooltipText?: string;
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
  topBorder,
  tooltipText,
  ...rest
}) => {
  const ref = useRef<HTMLButtonElement>(null);
  const {handleClick, handleKeyDown, shown} = useDropdownMenuItem({
    id,
    onClick,
    closeOnClick,
    ref,
    value,
  });
  const classes = classnames(styles.base, className, {
    [styles.tertiary]: buttonStyle === 'tertiary',
    [styles.important]: important,
    [styles.topBorder]: topBorder,
  });

  if (!shown) {
    return null;
  }

  return (
    <Tooltip
      tooltipText={tooltipText}
      disabled={!tooltipText}
      allowedPlacements={['left-start', 'right-end']}
      noSpanWrapper
    >
      <button
        ref={ref}
        data-testid="DropdownMenuItem__button"
        role="menuitem"
        type="button"
        className={classes}
        onClick={(e) => handleClick(e)}
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
    </Tooltip>
  );
};

export default DropdownMenuItem;
