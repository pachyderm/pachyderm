import classNames from 'classnames';
import React from 'react';
import {
  Link as ReactRouterLink,
  LinkProps as ReactRouterLinkProps,
} from 'react-router-dom';

import styles from './Link.module.css';

export interface LinkProps extends Omit<ReactRouterLinkProps, 'to'> {
  externalLink?: boolean;
  small?: boolean;
  inline?: boolean;
  to?: ReactRouterLinkProps['to'];
}

const Link: React.FC<LinkProps> = ({
  children,
  className,
  externalLink = false,
  small = false,
  inline = false,
  to,
  ...rest
}) => {
  const linkClassName = classNames(styles.link, className, {
    [styles.small]: small,
    [styles.inline]: inline,
  });

  if (!to) {
    return (
      <a className={linkClassName} {...rest}>
        {children}
      </a>
    );
  }

  if (externalLink) {
    return (
      <a
        href={to as string}
        target="_blank"
        rel="noopener noreferrer"
        className={linkClassName}
        {...rest}
      >
        {children}
      </a>
    );
  }

  return (
    <ReactRouterLink to={to} className={linkClassName} {...rest}>
      {children}
    </ReactRouterLink>
  );
};

export default Link;
