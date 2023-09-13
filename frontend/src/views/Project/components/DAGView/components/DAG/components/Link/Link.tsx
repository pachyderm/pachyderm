import classNames from 'classnames';
import React from 'react';

import {DagDirection, Link as LinkType} from '@dash-frontend/lib/types';

import useLink from './hooks/useLink';
import styles from './Link.module.css';

type LinkProps = {
  link: LinkType;
  dagDirection: DagDirection;
};

const Link: React.FC<LinkProps> = ({link, dagDirection}) => {
  const {d, transferring, isCrossProject, pathRef} = useLink(
    link,
    dagDirection,
  );

  const classes = classNames(styles.link, {
    [styles.transferring]: transferring,
    [styles.crossProject]: isCrossProject,
  });

  return (
    <g>
      <path ref={pathRef} className={classes} d={d} id={link.id} fill="none" />
      {transferring && (
        <circle r={6} className={styles.circle}>
          <animateMotion dur="0.8s" repeatCount="indefinite" path={d} />
        </circle>
      )}
    </g>
  );
};

export default Link;
