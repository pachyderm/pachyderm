import {ReadonlyJSONObject} from '@lumino/coreutils';

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
