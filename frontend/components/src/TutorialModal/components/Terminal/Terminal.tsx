import React from 'react';

import {TerminalSVG} from '@pachyderm/components';

import {CodeText} from './../../../Text';
import styles from './Terminal.module.css';

const Terminal: React.FC = ({children}) => {
  return (
    <div className={styles.terminal}>
      <CodeText className={styles.terminalCommand}>
        <TerminalSVG className={styles.terminalSVG} />
        Terminal Command
      </CodeText>
      <CodeText className={styles.terminalContent}>{children}</CodeText>
    </div>
  );
};

export default Terminal;
