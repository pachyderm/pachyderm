import flattenDeep from 'lodash/flattenDeep';
import {useState, useMemo, useCallback} from 'react';

import {useFetch} from 'hooks/useFetch';

export type ListItem = {
  keyString: string;
  valueString: string;
  depth: number;
  isObject: boolean;
  isOpen: boolean;
  childOf: string[];
  id: string;
};

type ListItemArray = ListItem | ListItemArray[];

const normalizeJSON = (
  obj: unknown,
  depth: number,
  minimizedRows: string[],
  childOf: string[] = [],
): ListItemArray[] => {
  if (typeof obj !== 'object' || Array.isArray(obj)) return [];

  const values = Object.values(obj || {});

  return Object.keys(obj || {}).map((key, index) => {
    const value = values[index];
    const isObject = typeof value === 'object';
    const isArray = Array.isArray(value);
    const identifier = `${childOf.join('')}${depth}${index}`;
    const isOpen = isObject && !minimizedRows.includes(identifier);

    return [
      {
        keyString: key,
        valueString: isObject
          ? `: ${isArray ? '[' : '{'}${
              !isOpen ? `...${isArray ? ']' : '}'}` : ''
            }`
          : `: ${JSON.stringify(value)}`,
        isOpen,
        isObject,
        childOf,
        depth,
        id: identifier,
      },
      ...(isOpen
        ? normalizeJSON(
            isArray ? {...value} : value,
            depth + 1,
            minimizedRows,
            [...childOf, identifier],
          )
        : []),
      ...(isOpen
        ? [
            {
              keyString: '',
              valueString: isArray ? `]` : '}',
              isOpen: false,
              isObject: false,
              childOf,
              depth,
              id: identifier,
            },
          ]
        : []),
    ];
  });
};

const useJSONPreview = (downloadLink: string) => {
  const formatResponse = useCallback(
    async (res: Response) => await res.json(),
    [],
  );
  const {data, loading, error} = useFetch({
    url: downloadLink,
    formatResponse,
  });
  const [minimizedRows, setMinimizedRows] = useState<string[]>([]);

  const flatData = useMemo(() => {
    return flattenDeep(normalizeJSON(data, 0, minimizedRows));
  }, [data, minimizedRows]);

  return {
    flatData,
    loading,
    error,
    minimizedRows,
    setMinimizedRows,
  };
};

export default useJSONPreview;
