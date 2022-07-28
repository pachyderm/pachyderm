import flattenDeep from 'lodash/flattenDeep';
import isEmpty from 'lodash/isEmpty';
import {useState, useMemo} from 'react';

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
  showBrackets: boolean,
  childOf: string[] = [],
): ListItemArray[] => {
  if (typeof obj !== 'object')
    return String(obj)
      .split('\n')
      .map((line, index) => ({
        keyString: '',
        valueString: line,
        depth: index > 0 ? 1 : 0,
        isObject: false,
        isOpen: true,
        childOf: [],
        id: line,
      }));

  const values = !Array.isArray(obj) ? Object.values(obj || {}) : obj;

  return (!Array.isArray(obj) ? Object.keys(obj || {}) : obj).map(
    (key, index) => {
      const value = values[index];
      const dataIsArray = Array.isArray(obj);
      const valueIsArray = Array.isArray(value);
      const valueIsObject = typeof value === 'object';

      const identifier = `${childOf.join('')}${depth}${index}`;
      const isOpen = valueIsObject && !minimizedRows.includes(identifier);

      let valueString = '';

      if (showBrackets) {
        valueString = valueIsObject
          ? `${!dataIsArray ? ': ' : ''}${valueIsArray ? '[' : '{'}${
              !isOpen ? `...${valueIsArray ? ']' : '}'}` : ''
            }`
          : `${key && !dataIsArray ? ': ' : ''}${JSON.stringify(value)}`;
      } else {
        valueString = valueIsObject
          ? !dataIsArray
            ? `: ${isEmpty(value) ? '[]' : ''}`
            : ''
          : `${key && !dataIsArray ? ': ' : ''}${JSON.stringify(value)}`;
      }

      const openingItem = {
        keyString: dataIsArray ? '' : key,
        valueString: `${
          !showBrackets && dataIsArray && valueString ? '- ' : ''
        }${valueString}`,
        isOpen,
        isObject: valueIsObject && !isEmpty(value),
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
        ...(showBrackets || openingItem.valueString ? [openingItem] : []),
        ...(isOpen
          ? normalizeJSON(value, depth + 1, minimizedRows, showBrackets, [
              ...childOf,
              identifier,
            ])
          : []),
        ...(isOpen && showBrackets ? [closingItem] : []),
      ];
    },
  );
};

const useDataPreview = (inputData: unknown, showBrackets: boolean) => {
  const [minimizedRows, setMinimizedRows] = useState<string[]>([]);

  const flatData = useMemo(() => {
    let flattenedData = flattenDeep(
      normalizeJSON(inputData, 1, minimizedRows, showBrackets),
    );
    const outsideBracketMetadata = {
      keyString: '',
      isOpen: false,
      isObject: true,
      childOf: [],
      depth: 0,
      id: '',
    };
    if (showBrackets) {
      flattenedData = [
        {
          ...outsideBracketMetadata,
          valueString: Array.isArray(inputData) ? `[` : '{',
        },
        ...flattenedData,
        {
          ...outsideBracketMetadata,
          valueString: Array.isArray(inputData) ? `]` : '}',
        },
      ];
    }
    return flattenedData;
  }, [inputData, minimizedRows, showBrackets]);

  return {
    flatData,
    minimizedRows,
    setMinimizedRows,
  };
};

export default useDataPreview;
