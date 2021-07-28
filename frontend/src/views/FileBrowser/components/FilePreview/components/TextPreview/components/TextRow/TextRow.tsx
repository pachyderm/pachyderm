import React, {CSSProperties} from 'react';

import styles from './TextRow.module.css';

type TextRowProps = {
  text: string;
  index: number;
  style: CSSProperties;
};

const TextRow: React.FC<TextRowProps> = ({text, style, index}) => {
  return (
    <li className={styles.base} style={style}>
      <span className={styles.index}>{index ? index : ''}</span>
      {text}
    </li>
  );
};

export default TextRow;
