import cloneDeep from 'lodash/cloneDeep';

/**
 * This constructs a special object.
 *
 * This object will have the same nested key structure as a userSpec.
 * The values in the nested object represent what should be show in
 * the effective spec as decorations.
 *
 * If the value is a primitive, that represents the previous value that should
 * be shown in strike through.
 * If the value is null, that represents that no previous value should be shown.
 */
export const createEffectiveSpecDecorationMap = (
  userSpec?: Record<string, unknown>,
  createPipelineDefaults?: Record<string, unknown>,
): Record<string, unknown> => {
  if (typeof userSpec !== 'object') return {};

  if (typeof createPipelineDefaults === 'undefined')
    createPipelineDefaults = {};

  const result =
    typeof structuredClone === 'function'
      ? structuredClone(userSpec)
      : cloneDeep(userSpec);

  /**
   * Recursively mutates nested objects in the given object.
   *
   * Looks for leaves in an object and assigns the values respective to the given
   * reference object.
   */
  const setEffectiveSpecDecorationMapLeaves = (
    obj: Record<string, unknown>,
    refObj: Record<string, unknown>,
  ): Record<string, unknown> => {
    for (const key of Object.keys(obj)) {
      // If both the objects contain a property with the same key
      if (key in refObj) {
        // Arrays are an edge case where we set the value to null
        if (Array.isArray(obj[key]) || Array.isArray(refObj[key])) {
          obj[key] = null;
        }

        // If this is an object we need to go deeper
        else if (
          typeof obj[key] === 'object' &&
          typeof refObj[key] === 'object' &&
          obj[key] !== null &&
          refObj[key] !== null
        ) {
          setEffectiveSpecDecorationMapLeaves(
            obj[key] as Record<string, unknown>,
            refObj[key] as Record<string, unknown>,
          );
        }

        // The value at this key is a primitive.
        else {
          // If the values haven't changed, set the value to null.
          // Otherwise set it to the previous value from the reference object.
          obj[key] = refObj[key] === obj[key] ? null : refObj[key];
        }
      } else {
        // Since the value was not found in the reference object we have a net-new value.

        // Array edge case
        if (Array.isArray(obj[key])) {
          obj[key] = null;
        }

        // Go deeper until this is a leaf then set to null
        if (typeof obj[key] === 'object' && obj[key] !== null) {
          setEffectiveSpecDecorationMapLeaves(
            obj[key] as Record<string, unknown>,
            {},
          );
        } else {
          // We have a primitive so we can set it to null.
          obj[key] = null;
        }
      }
    }
    return obj;
  };

  return setEffectiveSpecDecorationMapLeaves(result, createPipelineDefaults);
};
