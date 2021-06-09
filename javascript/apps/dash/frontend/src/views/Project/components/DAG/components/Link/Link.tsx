import classNames from 'classnames';
import React from 'react';

import {Link as LinkType} from '@graphqlTypes';

import convertNodeStateToDagState from '../../utils/convertNodeStateToDagState';

import useLink from './hooks/useLink';
import styles from './Link.module.css';

type LinkProps = {
  link: LinkType;
  isInteractive: boolean;
  offset: {x: number; y: number};
};

const Link: React.FC<LinkProps> = ({link, isInteractive, offset}) => {
  const {d, hoveredNode, selectedNode, transferring} = useLink(link, offset);
  const classes = classNames(
    styles.link,
    styles[`${convertNodeStateToDagState(link.sourceState)}Source`],
    styles[`${convertNodeStateToDagState(link.targetState)}Target`],
    {
      [styles.transferring]: transferring,
      [styles.selected]:
        (isInteractive && [selectedNode, hoveredNode].includes(link.source)) ||
        [selectedNode, hoveredNode].includes(link.target),
    },
  );

  return (
    <>
      <path className={classes} d={d} id={link.id} fill="none" />
      {transferring && (
        <circle r={6} className={styles.circle}>
          <animateMotion dur="0.8s" repeatCount="indefinite" path={d} />
        </circle>
      )}
    </>
  );
};

export default Link;
