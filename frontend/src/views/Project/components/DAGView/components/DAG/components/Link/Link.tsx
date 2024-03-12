import classNames from 'classnames';
import React from 'react';

import {DagDirection, Link as LinkType} from '@dash-frontend/lib/types';

import useLink from './hooks/useLink';
import styles from './Link.module.css';

type LinkProps = {
  link: LinkType;
  dagDirection: DagDirection;
  preorder?: string[];
  reversePreorder?: string[];
  hideDetails: boolean;
};

const Link: React.FC<LinkProps> = ({
  link,
  dagDirection,
  preorder,
  reversePreorder,
  hideDetails,
}) => {
  const {transferring, isCrossProject, id, speed, d, clipPath, highlightLink} =
    useLink(link, dagDirection, preorder, reversePreorder);

  const classes = classNames(styles.link, {
    [styles.transferring]: transferring,
    [styles.highlightLink]: highlightLink,
    [styles.crossProject]: isCrossProject,
  });

  // The path with strokeWidth 15 is being used to detecting if a user is hovering over our line.
  // This gives a larger area for selection instead of just allowing the hover behavior on the line.

  // The code inside of the “transferring” if statement is responsible for the animation.
  // We are animating a circle along the same path that we draw for the link. We then apply
  // the clipPath that was calculated in the hook to clip away part of the circle only leaving the
  // area that overlaps with the line creating the desired effect.
  return (
    <>
      <clipPath id={`${id}_clipPath`}>
        <path d={clipPath} />
      </clipPath>
      <g className={styles.group} id={`${link.id}_link`}>
        <path d={d} strokeWidth="15" pointerEvents="stroke" fill="none" />
        <path d={d} id={id} className={classes} fill="none" />

        {transferring && !hideDetails && (
          <g style={{clipPath: `url(#${id}_clipPath)`}}>
            <circle r={10} className={styles.circle}>
              <animateMotion dur={speed} repeatCount="indefinite" path={d} />
            </circle>
          </g>
        )}
      </g>
    </>
  );
};

export default Link;
