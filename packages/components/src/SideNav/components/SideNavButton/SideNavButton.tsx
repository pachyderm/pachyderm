import classnames from 'classnames';
import React from 'react';

import {Button, ButtonProps} from '../../../Button';
import {Tooltip} from '../../../Tooltip';
import useSideNav from '../../hooks/useSideNav';

import styles from './SideNavButton.module.css';

export interface SideNavButtonProps extends ButtonProps {
  ['data-testid']?: string;
  IconSVG: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  disabled?: boolean;
  styleMode?: 'light' | 'dark';
  tooltipContent: string;
}

const SideNavButton: React.FC<SideNavButtonProps> = ({
  'data-testid': dataTestId,
  to,
  children,
  IconSVG,
  disabled,
  styleMode = 'dark',
  tooltipContent,
  ...rest
}) => {
  const {minimized} = useSideNav();

  return (
    <Tooltip
      tooltipKey={tooltipContent}
      placement="right"
      disabled={!minimized}
      tooltipText={tooltipContent}
    >
      <Button
        to={to}
        className={classnames(styles.base, {
          [styles.disabled]: disabled,
          [styles[styleMode]]: true,
          [styles.minimized]: minimized,
        })}
        data-testid={dataTestId}
        {...rest}
        IconSVG={IconSVG}
      >
        {!minimized ? children : null}
      </Button>
    </Tooltip>
  );
};

export default SideNavButton;
