import React, {useRef, useState} from 'react';

import {Tooltip} from '../Tooltip';

import styles from './EllipsisMessage.module.css';

interface EllipsisMessageProps {
  message: string;
  'data-testid'?: string;
}

const EllipsisMessage: React.FC<EllipsisMessageProps> = ({
  message,
  ...rest
}) => {
  const [showTooltip, setShowTooltip] = useState(false);

  const ref = useRef<HTMLDivElement>(null);

  const shouldEnableTooltip = () => {
    if (ref.current) {
      // the parent element must have a static width for this to work correctly
      setShowTooltip(ref.current.clientWidth < ref.current.scrollWidth);
    }
  };

  return (
    <div {...rest} onMouseOver={shouldEnableTooltip}>
      <Tooltip tooltipText={message} disabled={!showTooltip}>
        <div className={styles.base} ref={ref}>
          {message}
        </div>
      </Tooltip>
    </div>
  );
};

export default EllipsisMessage;
