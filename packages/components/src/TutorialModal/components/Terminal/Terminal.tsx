import React from 'react';

import {TerminalSVG} from 'Svg';

import styles from './Terminal.module.css';

const Terminal: React.FC = ({children}) => {
  return (
    <div className={styles.terminal}>
      <span className={styles.terminalCommand}>
        <TerminalSVG className={styles.terminalSVG} />
        Terminal Command
      </span>
      <pre>{children}</pre>
    </div>
  );
};

export default Terminal;
