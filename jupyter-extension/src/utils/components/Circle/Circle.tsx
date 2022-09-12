import classNames from 'classnames';
import React from 'react';

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

export const Circle: React.FC<Props> = ({
  children,
  className,
  color = 'gray',
  ...props
}) => {
  const classes = classNames('jp-circle-base', className, {
    ['jp-circle-green']: color === colors.green,
    ['jp-circle-red']: color === colors.red,
    ['jp-circle-yellow']: color === colors.yellow,
    ['jp-circle-gray']: color === colors.gray,
  });
  return (
    <div className={classes} {...props}>
      {children}
    </div>
  );
};

export default Circle;
