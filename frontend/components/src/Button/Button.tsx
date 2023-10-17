import classNames from 'classnames';
import React, {Children, FunctionComponent, useMemo} from 'react';
import {Link, LinkProps} from 'react-router-dom';

import {Group} from '../Group';
import {Icon} from '../Icon';

import styles from './Button.module.css';

export type ButtonProps = React.ButtonHTMLAttributes<
  HTMLButtonElement & HTMLAnchorElement
> & {
  buttonType?:
    | 'primary'
    | 'secondary'
    | 'ghost'
    | 'tertiary'
    | 'quaternary'
    | 'dropdown';
  color?: string;
  href?: string;
  to?: LinkProps['to'];
  'data-testid'?: string;
  download?: boolean;
  IconSVG?: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  iconPosition?: 'start' | 'end' | 'both';
  buttonRef?: React.RefObject<HTMLButtonElement>;
};

export const ButtonGroup: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...rest
}) => {
  return (
    <div className={`${styles.buttonGroup} ${className}`} {...rest}>
      {children}
    </div>
  );
};

export const Button: FunctionComponent<ButtonProps> = ({
  color = '',
  href,
  IconSVG,
  className,
  buttonType = 'primary',
  children,
  disabled,
  to,
  iconPosition = 'start',
  download,
  buttonRef,
  ...props
}) => {
  const iconOnly = useMemo(
    () => IconSVG && Children.count(children) === 0,
    [IconSVG, children],
  );

  const classes = classNames(styles.base, className, {
    [styles.dropdown]: buttonType === 'dropdown',
    [styles.primary]: buttonType === 'primary',
    [styles.secondary]: buttonType === 'secondary',
    [styles.ghost]: buttonType === 'ghost',
    [styles.tertiary]: buttonType === 'tertiary',
    [styles.quaternary]: buttonType === 'quaternary',
    [styles.iconOnly]: iconOnly,
    [styles.link]: Boolean(href || to),
    [styles.black]: color === 'black',
  });

  const buttonChildren = useMemo(() => {
    let iconColor: 'plum' | 'black' | 'white' = 'white';

    switch (buttonType) {
      case 'secondary':
        iconColor = 'plum';
        break;
      case 'ghost':
        iconColor = color === 'black' ? 'black' : 'plum';
        break;
      case 'dropdown':
      case 'quaternary':
        iconColor = 'black';
        break;
      default:
        iconColor = 'white';
    }

    if (iconOnly && IconSVG) {
      return (
        <Icon small color={iconColor} disabled={disabled}>
          <IconSVG />
        </Icon>
      );
    }

    return (
      <Group spacing={8} align="center" justify="center">
        {IconSVG && (iconPosition === 'start' || iconPosition === 'both') && (
          <Icon small color={iconColor}>
            <IconSVG />
          </Icon>
        )}
        {children}
        {IconSVG && (iconPosition === 'end' || iconPosition === 'both') && (
          <Icon small color={iconColor}>
            <IconSVG />
          </Icon>
        )}
      </Group>
    );
  }, [IconSVG, children, disabled, buttonType, color, iconPosition, iconOnly]);

  if (!disabled && href) {
    return (
      <a
        className={classes}
        href={href}
        target={download ? undefined : '_blank'}
        rel="noopener noreferrer"
        download={download}
        {...props}
      >
        {buttonChildren}
      </a>
    );
  }

  if (!disabled && to) {
    return (
      <Link className={classes} to={to} {...props}>
        {buttonChildren}
      </Link>
    );
  }

  return (
    <button disabled={disabled} className={classes} ref={buttonRef} {...props}>
      {buttonChildren}
    </button>
  );
};
