import classNames from 'classnames';
import React, {FunctionComponent, isValidElement, Children} from 'react';

import styles from './Group.module.css';

export type GroupProps = React.HTMLAttributes<HTMLDivElement> & {
  inline?: boolean;
  vertical?: boolean;
  spacing?: 8 | 16 | 24 | 32 | 40 | 64 | 128;
  justify?: string;
  align?: string;
  reverse?: boolean;
};

// Ensure that each Group child is wrapped in an element (i.e. if there is raw
// text used inside a Group).
const ensureElement = (x: React.ReactNode): React.ReactElement | null => {
  if (!x) {
    return null;
  }

  return isValidElement(x) ? x : <div>{x}</div>;
};

// // TODO: if we end up needing more of these, just make a generic component
const NoShrink: FunctionComponent<GroupProps> = ({
  children,
  className,
  ...props
}) => {
  const classes = classNames(className, styles.noshrink);
  return (
    <Group className={classes} {...props}>
      {children}
    </Group>
  );
};

const Divider: FunctionComponent<GroupProps> = ({className, ...props}) => {
  const classes = classNames(className, styles.divider);
  return <NoShrink className={classes} {...props} />;
};

const Group: FunctionComponent<GroupProps> = ({
  children,
  className,
  inline,
  vertical,
  spacing,
  justify,
  align,
  reverse,
  ...props
}) => {
  const classes = classNames(className, styles.base, {
    [styles.inline]: inline,
    [styles.vertical]: vertical,
    [styles.horizontal]: !vertical,
    [styles.reverse]: reverse,

    [styles.justifyEnd]: justify === 'end',
    [styles.justifyCenter]: justify === 'center',
    [styles.justifyStretch]: justify === 'stretch',
    [styles.justifyBetween]: justify === 'between',

    [styles.alignCenter]: align === 'center',
    [styles.alignStart]: align === 'start',
    [styles.alignEnd]: align === 'end',
    [styles.alignBaseline]: align === 'baseline',

    [styles.spacing8]: spacing === 8,
    [styles.spacing16]: spacing === 16,
    [styles.spacing24]: spacing === 24,
    [styles.spacing32]: spacing === 32,
    [styles.spacing40]: spacing === 40,
    [styles.spacing64]: spacing === 64,
    [styles.spacing128]: spacing === 128,
  });

  return (
    <div className={classes} {...props}>
      {Children.map(children, ensureElement)}
    </div>
  );
};

export default Object.assign(Group, {NoShrink, Divider});
