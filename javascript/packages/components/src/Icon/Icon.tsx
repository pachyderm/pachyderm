import classNames from 'classnames';
import React, {FunctionComponent} from 'react';

import styles from './Icon.module.css';

const Colors = {
  black: 'black',
  white: 'white',
  purple: 'purple',
  silver: 'silver',
};

type IconColor = keyof typeof Colors;

export type Props = React.HTMLAttributes<HTMLDivElement> & {
  color?: IconColor;
  small?: boolean;
};

const defaultColor = Colors.purple;

export const Icon: FunctionComponent<Props> = ({
  children,
  color,
  small,
  style,
  className,
  ...rest
}) => {
  const actualColor = (color && Colors[color]) || defaultColor;
  const classes = classNames(styles.base, className, {
    [styles.black]: actualColor === Colors.black,
    [styles.white]: actualColor === Colors.white,
    [styles.purple]: actualColor === Colors.purple,
    [styles.silver]: actualColor === Colors.silver,
    [styles.small]: small,
  });
  return (
    <div className={classes} style={style} {...rest}>
      {children}
    </div>
  );
};

export default Icon;
