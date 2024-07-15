import classnames from 'classnames';
import React, {HTMLAttributes, useState} from 'react';

import {Icon, ChevronDownSVG, ChevronRightSVG} from '@pachyderm/components';

import styles from './CollapsibleSection.module.css';

interface CollapsibleSectionProps extends HTMLAttributes<HTMLDivElement> {
  children?: React.ReactNode;
  header: React.ReactElement;
  sectionRef?: React.RefObject<HTMLDivElement>;
  className?: string;
  headerContent?: JSX.Element;
  maxHeight?: number;
  boxShadow?: boolean;
}

const CollapsibleSection: React.FC<CollapsibleSectionProps> = ({
  header,
  children,
  sectionRef,
  className,
  maxHeight,
  boxShadow = false,
}) => {
  const [minimized, setMinimized] = useState(false);
  return (
    <div className={classnames({[styles.boxShadow]: boxShadow})}>
      <div
        className={styles.header}
        onClick={() => setMinimized((prevValue) => !prevValue)}
        ref={sectionRef}
      >
        {header}
        <Icon small>
          {minimized ? <ChevronRightSVG /> : <ChevronDownSVG />}
        </Icon>
      </div>
      <div
        className={classnames(
          {[styles.minimized]: minimized},
          styles.content,
          className,
        )}
        style={{maxHeight: `${maxHeight}px`}}
      >
        {children}
      </div>
    </div>
  );
};

export default CollapsibleSection;
