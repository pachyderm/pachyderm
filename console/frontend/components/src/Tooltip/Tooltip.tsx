import {
  useFloating,
  useHover,
  useInteractions,
  arrow,
  FloatingArrow,
  offset,
  shift,
  autoPlacement,
  autoUpdate,
  Placement,
} from '@floating-ui/react';
import classNames from 'classnames';
import React, {useState, useRef} from 'react';

import styles from './Tooltip.module.css';

const ARROW_HEIGHT = 7;
const GAP = 2;

type TooltipProps = {
  disabled?: boolean;
  tooltipText: string | React.ReactNode;
  className?: string;
  noWrap?: boolean;
  children?: React.ReactNode;
  allowedPlacements?: Array<Placement>;
  noSpanWrapper?: boolean;
};

const Tooltip: React.FC<TooltipProps> = ({
  disabled = false,
  tooltipText,
  children,
  noWrap = false,
  allowedPlacements,
  noSpanWrapper,
  className,
}) => {
  const arrowRef = useRef(null);
  const [isOpen, setIsOpen] = useState(false);

  const {refs, floatingStyles, context} = useFloating({
    open: isOpen,
    onOpenChange: setIsOpen,
    // Ordering of middleware is important
    middleware: [
      offset(ARROW_HEIGHT + GAP),
      autoPlacement({allowedPlacements}),
      shift(),
      arrow({
        element: arrowRef,
      }),
    ],
    whileElementsMounted: autoUpdate,
  });

  const hover = useHover(context, {
    enabled: !disabled,
  });

  const {getReferenceProps, getFloatingProps} = useInteractions([hover]);

  if (disabled && noSpanWrapper) {
    return <>{children}</>;
  }

  if (disabled) {
    return <span>{children}</span>;
  }

  return (
    <>
      <span
        ref={refs.setReference}
        aria-describedby="tooltip"
        {...getReferenceProps()}
        className={styles.element}
      >
        {children}
      </span>
      {isOpen && (
        <div
          role="tooltip"
          id="tooltip"
          ref={refs.setFloating}
          style={floatingStyles}
          className={classNames(
            styles.base,
            {[styles.noWrap]: noWrap},
            className,
          )}
          {...getFloatingProps()}
        >
          {tooltipText}
          <FloatingArrow ref={arrowRef} context={context} />
        </div>
      )}
    </>
  );
};

export default Tooltip;
