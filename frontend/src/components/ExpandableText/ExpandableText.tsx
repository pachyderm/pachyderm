import classNames from 'classnames';
import React, {useState, useRef, useEffect} from 'react';

import {
  ButtonLink,
  ChevronUpSVG,
  ChevronDownSVG,
  Icon,
} from '@pachyderm/components';

import styles from './ExpandableText.module.css';

const DEFAULT_LINE_HEIGHT = 1.437 * 16;

type ExpandableTextProps = {
  text: string;
  lineClamp?: string;
  lineHeight?: number;
};

const ExpandableText: React.FC<ExpandableTextProps> = ({
  text,
  lineClamp = '2',
  lineHeight = DEFAULT_LINE_HEIGHT,
}) => {
  const [expanded, setExpanded] = useState(false);
  const [canExpand, setCanExpand] = useState(false);
  const textRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (
      textRef.current &&
      textRef.current.clientHeight > Number(lineClamp) * lineHeight
    ) {
      setCanExpand(true);
    } else {
      setCanExpand(false);
    }
  }, [lineClamp, lineHeight]);

  return (
    <>
      <div
        className={classNames(styles.base, {
          [styles.minimized]: canExpand && !expanded,
        })}
        style={{WebkitLineClamp: lineClamp}}
        ref={textRef}
      >
        {text}
      </div>
      {canExpand && (
        <ButtonLink
          onClick={() => setExpanded(!expanded)}
          className={styles.inlineButton}
        >
          {expanded ? 'Read Less' : 'Read More'}
          <Icon smaller color="plum">
            {expanded ? <ChevronUpSVG /> : <ChevronDownSVG />}
          </Icon>
        </ButtonLink>
      )}
    </>
  );
};

export default ExpandableText;
