import classNames from 'classnames';
import React, {FunctionComponent} from 'react';
import {Link, LinkProps} from 'react-router-dom';

import styles from './Button.module.css';

export type ButtonProps = React.ButtonHTMLAttributes<
  HTMLButtonElement & HTMLAnchorElement
> & {
  autoHeight?: boolean;
  autoSize?: boolean;
  autoWidth?: boolean;
  buttonType?: 'primary' | 'secondary' | 'tertiary';
  href?: string;
  to?: LinkProps['to'];
  popIn?: boolean;
  'data-testid'?: string;
  download?: boolean;
};

export const Button: FunctionComponent<ButtonProps> = ({
  autoHeight = false,
  autoSize = false,
  autoWidth = false,
  href,
  popIn = false,
  className,
  buttonType = 'primary',
  children,
  disabled,
  to,
  ...props
}) => {
  const classes = classNames(styles.base, className, {
    [styles.primary]: buttonType === 'primary',
    [styles.secondary]: buttonType === 'secondary',
    [styles.tertiary]: buttonType === 'tertiary',
    [styles.autoHeight]: autoHeight,
    [styles.autoSize]: autoSize,
    [styles.autoWidth]: autoWidth,
    [styles.link]: Boolean(href || to),
    [styles.popIn]: popIn,
  });

  if (href) {
    return (
      <a className={classes} href={href} {...props}>
        {children}
      </a>
    );
  }

  if (to) {
    return (
      <Link className={classes} to={to} {...props}>
        {children}
      </Link>
    );
  }

  return (
    <button disabled={disabled} className={classes} {...props}>
      {children}
    </button>
  );
};
