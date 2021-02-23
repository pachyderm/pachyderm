import classNames from 'classnames';
import React from 'react';

import styles from './Circle.module.css';

const colors = {
  green: 'green',
  red: 'red',
  yellow: 'yellow',
  gray: 'gray',
};

export type CircleColor = keyof typeof colors;

type Props = React.HTMLAttributes<HTMLDivElement> & {
  color?: CircleColor;
};

const Circle: React.FC<Props> = ({
  children,
  className,
  color = 'gray',
  ...props
}) => {
  const classes = classNames(styles.base, className, {
    [styles.green]: color === colors.green,
    [styles.red]: color === colors.red,
    [styles.yellow]: color === colors.yellow,
    [styles.gray]: color === colors.gray,
  });
  return (
    <div className={classes} {...props}>
      {children}
    </div>
  );
};

export default Circle;
