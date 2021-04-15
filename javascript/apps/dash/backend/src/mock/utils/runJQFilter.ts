import {run} from 'node-jq';

interface RunJQFilterArgs<T> {
  jqFilter: string;
  object: Record<string, unknown>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  objectMapper: (json: any) => T;
}

const runJQFilter = <T>({
  jqFilter,
  object,
  objectMapper,
}: RunJQFilterArgs<T>): Promise<T[]> => {
  return new Promise((res) => {
    run(jqFilter, object, {
      input: 'json',
      output: 'string',
    }).then((filteredResponse) => {
      if (filteredResponse && typeof filteredResponse === 'string') {
        const parsedRecords = filteredResponse.split('\n').map((record) => {
          return objectMapper(JSON.parse(record));
        });

        res(parsedRecords);
      } else {
        res([]);
      }
    });
  });
};

export default runJQFilter;
