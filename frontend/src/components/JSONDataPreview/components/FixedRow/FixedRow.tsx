import classNames from 'classnames';
import times from 'lodash/times';
import React, {CSSProperties} from 'react';

import {PADDING_SIZE} from '@dash-frontend/views/FileBrowser/constants/FileBrowser';
import {
  Icon,
  ChevronDownSVG,
  ChevronRightSVG,
  CodePreview,
} from '@pachyderm/components';

import {ListItem} from '../../hooks/useDataPreview';

import styles from './FixedRow.module.css';

type FixedRowProps = {
  itemData: ListItem;
  style: CSSProperties;
  handleExpand: () => void;
  handleMinimize: () => void;
};

const FixedRow: React.FC<FixedRowProps> = ({
  itemData,
  style,
  handleMinimize,
  handleExpand,
}) => {
  const {keyString, valueString, depth, isOpen, isObject, id} = itemData;

  return (
    <li
      className={classNames(
        styles.base,
        styles[`depth${depth === 0 ? '0' : depth % 6}`],
        {
          [styles.clickable]: isObject,
        },
      )}
      style={{
        ...style,
        top: `${Number(style.top) + PADDING_SIZE}px`,
      }}
      onClick={isOpen ? handleMinimize : handleExpand}
    >
      {times(depth, (index) => (
        <span key={index} className={styles.depthTab} />
      ))}
      {isObject && id && (
        <Icon small className={styles.chevron}>
          {isOpen ? <ChevronDownSVG /> : <ChevronRightSVG />}
        </Icon>
      )}
      {keyString}
      <CodePreview className={classNames(styles.value, 'language-json')}>
        {valueString}
      </CodePreview>
    </li>
  );
};

export default FixedRow;
