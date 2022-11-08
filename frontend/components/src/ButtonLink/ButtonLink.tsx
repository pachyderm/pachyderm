import classNames from 'classnames';
import React, {ButtonHTMLAttributes} from 'react';

import styles from './ButtonLink.module.css';

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  small?: boolean;
}

export const ButtonLink: React.FC<Props> = ({
  children,
  className,
  small = false,
  ...props
}) => {
  const classes = classNames(styles.base, className, {
    [styles.small]: small,
  });

  return (
    <button className={classes} {...props}>
      {children}
    </button>
  );
};

export default ButtonLink;
