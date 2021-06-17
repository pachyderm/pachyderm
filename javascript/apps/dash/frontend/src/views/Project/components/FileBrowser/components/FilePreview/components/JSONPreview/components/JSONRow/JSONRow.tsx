import classNames from 'classnames';
import times from 'lodash/times';
import React, {CSSProperties} from 'react';

import {ListItem} from '../../hooks/useJSONPreview';
import {PADDING_SIZE} from '../../JSONPreview';

import styles from './JSONRow.module.css';

type JSONRowProps = {
  itemData: ListItem;
  style: CSSProperties;
  handleExpand: () => void;
  handleMinimize: () => void;
};

const JSONRow: React.FC<JSONRowProps> = ({
  itemData,
  style,
  handleMinimize,
  handleExpand,
}) => {
  const {keyString, valueString, depth, isOpen, isObject} = itemData;

  return (
    <li
      className={classNames(styles.base, styles[`depth${depth % 6}`], {
        [styles.clickable]: isObject,
      })}
      style={{
        ...style,
        top: `${Number(style.top) + PADDING_SIZE}px`,
      }}
      onClick={isOpen ? handleMinimize : handleExpand}
    >
      {times(depth + (isObject ? 0 : 1), (index) => (
        <span key={index} className={styles.depthTab} />
      ))}
      {isObject && (
        <span
          className={classNames(
            styles.value,
            styles[isOpen ? 'arrowDown' : 'arrowRight'],
          )}
        >
          ▼
        </span>
      )}
      {keyString}
      <span className={styles.value}>{valueString}</span>
    </li>
  );
};

export default JSONRow;
