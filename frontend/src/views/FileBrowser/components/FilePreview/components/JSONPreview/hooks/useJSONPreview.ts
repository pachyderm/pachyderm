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
  if (typeof obj !== 'object') return [];

  const values = !Array.isArray(obj) ? Object.values(obj || {}) : obj;

  return (!Array.isArray(obj) ? Object.keys(obj || {}) : obj).map(
    (key, index) => {
      const value = values[index];
      const dataIsArray = Array.isArray(obj);
      const valueIsArray = Array.isArray(value);
      const valueIsObject = typeof value === 'object';

      const identifier = `${childOf.join('')}${depth}${index}`;
      const isOpen = valueIsObject && !minimizedRows.includes(identifier);

      const openingItem = {
        keyString: dataIsArray ? '' : key,
        valueString: valueIsObject
          ? `${!dataIsArray ? ': ' : ''}${valueIsArray ? '[' : '{'}${
              !isOpen ? `...${valueIsArray ? ']' : '}'}` : ''
            }`
          : `${key && !dataIsArray ? ': ' : ''}${JSON.stringify(value)}`,
        isOpen,
        isObject: valueIsObject,
        childOf,
        depth,
        id: identifier,
      };

      const closingItem = {
        keyString: '',
        valueString: valueIsArray ? `]` : '}',
        isOpen: false,
        isObject: false,
        childOf,
        depth,
        id: identifier,
      };

      return [
        openingItem,
        ...(isOpen
          ? normalizeJSON(value, depth + 1, minimizedRows, [
              ...childOf,
              identifier,
            ])
          : []),
        ...(isOpen ? [closingItem] : []),
      ];
    },
  );
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
    let flattenedData = flattenDeep(normalizeJSON(data, 0, minimizedRows));
    const outsideBracketMetadata = {
      keyString: '',
      isOpen: false,
      isObject: true,
      childOf: [],
      depth: 0,
      id: '',
    };

    flattenedData = [
      {
        ...outsideBracketMetadata,
        valueString: Array.isArray(data) ? `[` : '{',
      },
      ...flattenedData,
      {
        ...outsideBracketMetadata,
        valueString: Array.isArray(data) ? `]` : '}',
      },
    ];
    return flattenedData;
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
