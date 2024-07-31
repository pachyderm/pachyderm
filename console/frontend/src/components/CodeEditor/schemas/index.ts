import rawClusterDefaultsSchema from '@pachyderm/src/internal/jsonschema/pps_v2/ClusterDefaults.schema.json';
import rawCreatePipelineRequestSchema from '@pachyderm/src/internal/jsonschema/pps_v2/CreatePipelineRequest.schema.json';
import rawProjectDefaultsSchema from '@pachyderm/src/internal/jsonschema/pps_v2/ProjectDefaults.schema.json';
import {type JSONSchema7} from 'json-schema';
import cloneDeep from 'lodash/cloneDeep';

/** Type guard that narrows the type correctly. */
function hasKey<T extends object>(
  obj: T,
  key: string | number | symbol,
): obj is T & Record<typeof key, unknown> {
  return key in obj;
}

/**
 * This is needed because the CodeMirror JSON Schema plugin is not reading top level refs correctly.
 * This takes the content of the item referenced in the top level ref and puts it on the top level.
 * This function does a lot of type narrowing.
 */
export const fixSchema = (
  rawSchema: Record<string, unknown> | JSONSchema7,
): JSONSchema7 => {
  if (typeof rawSchema !== 'object' || rawSchema === null) return {};

  // Clone the rawSchema depending on the availability of structuredClone
  const schema =
    typeof structuredClone === 'function'
      ? structuredClone(rawSchema)
      : cloneDeep(rawSchema);

  // Extract value from $ref, e.g., #/definitions/ClusterDefaults
  const rawRef = schema['$ref'];

  // Remove top-level $ref
  delete schema['$ref'];

  if (typeof rawRef === 'string') {
    const key = rawRef.split('/')[2];
    const definitions = schema.definitions;

    if (
      definitions &&
      typeof definitions === 'object' &&
      hasKey(definitions, key)
    ) {
      const referencedDefinition = definitions[key];

      if (typeof referencedDefinition === 'object') {
        // Merge the top level with the referenced definition
        return {...schema, ...referencedDefinition};
      }
    }
  }

  return schema;
};

export const clusterDefaultsSchema: JSONSchema7 = fixSchema(
  rawClusterDefaultsSchema,
);

export const projectDefaultsSchema: JSONSchema7 = fixSchema(
  rawProjectDefaultsSchema,
);

export const createPipelineRequestSchema: JSONSchema7 = fixSchema(
  rawCreatePipelineRequestSchema,
);
