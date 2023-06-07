import {dsvFormat} from 'd3-dsv';
import {text} from 'd3-fetch';
import range from 'lodash/range';
import zipObject from 'lodash/zipObject';
import {useState, useEffect, useMemo} from 'react';

const DELIMITERS = [',', '\t', ';', '|'];
export const DELIMITER_ITEMS = [
  {id: ',', content: 'comma'},
  {id: '\t', content: 'tab'},
  {id: ';', content: 'semicolon'},
  {id: '|', content: 'pipe'},
];
interface Occurrences {
  [delimiter: string]: number[];
}
const guessBestDelimiter = (text: string) => {
  const lines = text.split('\n');

  // Handle a file without newlines
  if (lines.length === 1) {
    lines.push('');
  }

  const occurrencesByLine = lines
    .slice(0, Math.min(lines.length - 1, 10))
    .reduce((memo: Occurrences, line, index) => {
      DELIMITERS.forEach((delimiter) => {
        memo[delimiter] = memo[delimiter] || [];
        memo[delimiter][index] = line
          .split('')
          .filter((c) => c === delimiter).length;
      });
      return memo;
    }, {});
  return DELIMITERS.sort((a, b) => {
    const minA = Math.min(...occurrencesByLine[a]);
    const minB = Math.min(...occurrencesByLine[b]);
    if (minA === 0 && minB > 0) return 1;
    if (minB === 0 && minA > 0) return -1;
    return (
      Math.max(...occurrencesByLine[a]) -
      minA -
      (Math.max(...occurrencesByLine[b]) - minB)
    );
  })[0];
};

export type FilePreviewProps = {
  downloadLink: string | undefined | null;
};

const useCSVPreview = ({downloadLink}: FilePreviewProps) => {
  const [data, setData] = useState<Record<string, unknown>[]>([]);
  const [rawData, setRawData] = useState('');
  const [loading, setLoading] = useState(false);
  const [delimiter, setDelimiter] = useState(DELIMITERS[0]);
  const delimiterLabel = useMemo(() => {
    return DELIMITER_ITEMS.find(({id}) => id === delimiter)?.content;
  }, [delimiter]);

  useEffect(() => {
    const getData = async () => {
      if (downloadLink) {
        setLoading(true);
        const responseData = await text(downloadLink, {
          credentials: 'include',
          mode: 'cors',
        }).catch((e) => console.error(e));
        if (responseData) {
          setRawData(responseData);
          const delimiterGuess = guessBestDelimiter(responseData);
          setDelimiter(delimiterGuess);
        }
        setLoading(false);
      }
    };
    getData();
  }, [downloadLink]);

  useEffect(() => {
    const format = dsvFormat(delimiter);
    const parsedData = format.parse(rawData);

    // Handle a single line csv file
    if (
      parsedData &&
      parsedData.length === 0 &&
      parsedData.columns.length > 0
    ) {
      parsedData.push(
        // Use the array index + 1 as the column header and object key
        zipObject(range(1, parsedData.columns.length + 1), parsedData.columns),
      );
    }

    parsedData && setData(parsedData);
  }, [rawData, delimiter]);

  const headers = useMemo(() => Object.keys(data[0] || {}), [data]);

  return {
    headers,
    data,
    loading,
    delimiter,
    setDelimiter,
    delimiterLabel,
  };
};

export default useCSVPreview;
