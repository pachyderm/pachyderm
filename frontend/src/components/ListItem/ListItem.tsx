import classnames from 'classnames';
import React from 'react';

import {CaptionTextSmall, Icon, IconColor} from '@pachyderm/components';

import styles from './ListItem.module.css';

export type ListItemProps = {
  state?: 'default' | 'highlighted' | 'selected';
  LeftIconSVG?: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  leftIconColor?: IconColor;
  RightIconSVG?: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  rightIconColor?: IconColor;
  text?: string;
  captionText?: string;
  onClick?: () => void;
  'data-testid'?: string;
  role?: string;
};

const ListItem: React.FC<ListItemProps> = ({
  state = 'default',
  LeftIconSVG,
  leftIconColor,
  RightIconSVG,
  rightIconColor,
  text,
  captionText,
  onClick,
  ...rest
}) => {
  return (
    <div
      {...rest}
      onClick={onClick}
      className={classnames(styles.base, {
        [styles.highlighted]: state === 'highlighted',
        [styles.selected]: state === 'selected',
      })}
    >
      <div className={styles.left}>
        {LeftIconSVG && (
          <Icon small color={leftIconColor}>
            <LeftIconSVG />
          </Icon>
        )}
        <div
          className={classnames(styles.leftText, {
            [styles.leftIcon]: LeftIconSVG,
          })}
        >
          {text}
        </div>
      </div>

      <div className={styles.right}>
        {captionText && (
          <CaptionTextSmall className={styles.caption}>
            {captionText}
          </CaptionTextSmall>
        )}
        {RightIconSVG && (
          <Icon small color={rightIconColor}>
            <RightIconSVG />
          </Icon>
        )}
      </div>
    </div>
  );
};
export default ListItem;
