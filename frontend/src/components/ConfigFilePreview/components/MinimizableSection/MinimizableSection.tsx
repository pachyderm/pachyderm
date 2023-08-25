import classnames from 'classnames';
import React, {useState} from 'react';

import {Icon, ChevronDownSVG, ChevronRightSVG} from '@pachyderm/components';

import styles from './MinimizableSection.module.css';

interface MinimizableSectionProps {
  children?: React.ReactNode;
  header: React.ReactElement;
}

const MinimizableSection: React.FC<MinimizableSectionProps> = ({
  header,
  children,
}) => {
  const [minimized, setMinimized] = useState(false);
  return (
    <>
      <div
        className={styles.header}
        onClick={() => setMinimized((prevValue) => !prevValue)}
      >
        <Icon small className={styles.chevron}>
          {minimized ? <ChevronRightSVG /> : <ChevronDownSVG />}
        </Icon>
        {header}
      </div>
      <div className={classnames({[styles.minimized]: minimized})}>
        {children}
      </div>
    </>
  );
};

export default MinimizableSection;
