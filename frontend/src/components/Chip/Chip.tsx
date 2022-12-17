import classNames from 'classnames';
import React, {ButtonHTMLAttributes} from 'react';

import {Group, Icon} from '@pachyderm/components';

import styles from './Chip.module.css';

export interface ChipProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  LeftIconSVG?: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  RightIconSVG?: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
}

export const Chip = ({
  LeftIconSVG,
  RightIconSVG,
  className,
  children,
  ...rest
}: React.PropsWithChildren<ChipProps>) => {
  const classes = classNames(styles.base, className);

  return (
    <button className={classes} {...rest}>
      <Group spacing={8} align="center" justify="center">
        {LeftIconSVG && (
          <Icon small>
            <LeftIconSVG />
          </Icon>
        )}
        {children}
        {RightIconSVG && (
          <Icon small>
            <RightIconSVG />
          </Icon>
        )}
      </Group>
    </button>
  );
};

export const ChipGroup: React.FC = ({children}) => {
  return <div className={styles.group}>{children}</div>;
};
