import React, {useEffect} from 'react';
import {VariableSizeList, VariableSizeListProps} from 'react-window';

interface LogsListProps extends VariableSizeListProps {
  forwardRef: React.RefObject<VariableSizeList>;
}

const LogsList: React.FC<LogsListProps> = ({
  children,
  forwardRef,
  itemCount,
  ...props
}) => {
  useEffect(() => {
    if (forwardRef?.current && itemCount) {
      forwardRef.current.scrollToItem(itemCount, 'end');
    }
  }, [forwardRef, itemCount]);
  return (
    <VariableSizeList {...props} itemCount={itemCount} ref={forwardRef}>
      {children}
    </VariableSizeList>
  );
};

export default LogsList;
