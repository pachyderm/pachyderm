import {ReadonlyJSONObject} from '@lumino/coreutils';

// Normally it is bad practice to use an any type, however given that this a function for mocking API responses used
// in tests this seems fine. Any type issues will be found in the test failures.
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
export function mockedRequestAPI(mockedResponses?: any): any {
  const mockedImplementation = (
    endPoint?: string,
    method?: string,
    body?: ReadonlyJSONObject | null,
    namespace?: string,
  ) => {
    mockedResponses = mockedResponses ?? {};
    return mockedResponses;
  };
  return mockedImplementation;
}
