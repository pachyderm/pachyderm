import classNames from 'classnames';
import React, {FunctionComponent} from 'react';

import styles from './Icon.module.css';

const Colors = {
  black: 'black',
  white: 'white',
  plum: 'plum',
  grey: 'grey',
  green: 'green',
  blue: 'blue',
  red: 'red',
  yellow: 'yellow',
  highlightGreen: 'highlightGreen',
  highlightOrange: 'highlightOrange',
};

type IconColor = keyof typeof Colors;

export type Props = React.HTMLAttributes<HTMLDivElement> & {
  color?: IconColor;
  small?: boolean;
  smaller?: boolean;
  disabled?: boolean;
};

const defaultColor = Colors.black;

export const Icon: FunctionComponent<Props> = ({
  children,
  color,
  small,
  smaller,
  style,
  className,
  disabled,
  ...rest
}) => {
  const actualColor = (color && Colors[color]) || defaultColor;
  const classes = classNames(styles.base, className, {
    [styles[actualColor]]: true,
    [styles.small]: small,
    [styles.smaller]: smaller,
    [styles.disabled]: disabled,
  });
  return (
    <div className={classes} style={style} {...rest}>
      {children}
    </div>
  );
};

export default Icon;
