import {useState, useMemo, useCallback} from 'react';

export type SortableItem<T> = {
  name: string;
  accessor: (a: T) => string | number | T;
  func: (a: any, b: any) => number;
  reverse?: boolean;
};
type useSortProps<T> = {
  data: T[];
  initialSort?: SortableItem<T>;
  initialDirection?: number;
};
export const stringComparator = (a: string, b: string) => {
  if (a > b) {
    return 1;
  } else if (a < b) {
    return -1;
  }
  return 0;
};

export const numberComparator = (a: number, b: number) => a - b;

export const useSort = <T>({
  data,
  initialSort = {
    func: () => 1,
    name: '',
    accessor: (a: T) => a,
  },
  initialDirection = 1,
}: useSortProps<T>) => {
  const [direction, setDirection] = useState(initialDirection);
  const [comparator, setComparator] = useState(initialSort);

  const handleSetComparator = useCallback(
    (newComparator: SortableItem<T>) => {
      setComparator((oldComparator) => {
        if (
          newComparator.name === oldComparator.name ||
          newComparator.reverse
        ) {
          setDirection(-1 * direction);
        } else {
          setDirection(1);
        }
        return newComparator;
      });
    },
    [direction],
  );

  const sortedData = useMemo(() => {
    return [...data].sort((a, b) => {
      return (
        direction *
        comparator.func(comparator.accessor(a), comparator.accessor(b))
      );
    });
  }, [comparator, data, direction]);

  return {
    sortedData,
    reversed: direction === -1,
    setComparator: handleSetComparator,
    comparatorName: comparator.name,
    numberComparator,
    stringComparator,
  };
};
