import {type JSONSchema7} from 'json-schema';
import {useCallback, useEffect, useState} from 'react';

import {useFetch} from '@dash-frontend/hooks/useFetch';

import {fixSchema} from './schemas';

export const useJsonSchema = (
  schemaUrl: string,
  fallback: JSONSchema7,
  skip: boolean,
) => {
  const formatResponse = useCallback(async (res: Response) => {
    return await res.text();
  }, []);
  const {data, loading, error} = useFetch({
    url: `/jsonschema/${schemaUrl}`,
    formatResponse,
    skip,
  });
  const [schema, setSchema] = useState<JSONSchema7>();

  useEffect(() => {
    if (!loading && !error && data) {
      try {
        const json = fixSchema(JSON.parse(data));

        if (
          Object.keys(json).length === 0 ||
          (!Object.hasOwn(json, 'properties') &&
            !Object.hasOwn(json, 'definitions'))
        ) {
          throw new Error('JSON Schema is empty.');
        }

        setSchema(json);
      } catch {
        setSchema(fallback);
      }
    }

    if (error) {
      setSchema(fallback);
    }
  }, [loading, error, data, fallback]);

  return {schema, loading, error};
};
