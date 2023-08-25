import classNames from 'classnames';
import React, {ButtonHTMLAttributes} from 'react';

import {Group, Icon} from '@pachyderm/components';

import styles from './Chip.module.css';

export interface ChipProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  LeftIconSVG?: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  LeftIconSmall?: boolean;
  RightIconSVG?: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  isButton?: boolean;
}

type ChipGroupProps = {
  children?: React.ReactNode;
  className?: string;
};

export const Chip = ({
  LeftIconSVG,
  LeftIconSmall = true,
  RightIconSVG,
  className,
  children,
  isButton = true,
  ...rest
}: ChipProps) => {
  const classes = classNames(styles.base, className);

  const ChipWrapper = ({children}: {children?: React.ReactNode}) => {
    if (isButton) {
      return (
        <button className={classes} {...rest}>
          {children}
        </button>
      );
    } else {
      return <div className={classes}>{children}</div>;
    }
  };

  return (
    <ChipWrapper>
      <Group spacing={8} align="center" justify="center">
        {LeftIconSVG && (
          <Icon small={LeftIconSmall}>
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
    </ChipWrapper>
  );
};

export const ChipGroup: React.FC<ChipGroupProps> = ({className, children}) => {
  return <div className={classNames(styles.group, className)}>{children}</div>;
};
