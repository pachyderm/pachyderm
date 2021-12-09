import {URLExt} from '@jupyterlab/coreutils';
import {ServerConnection} from '@jupyterlab/services';
import {ReadonlyJSONObject} from '@lumino/coreutils';

export async function requestAPI<T>(
  endPoint = '',
  method = 'GET',
  body: ReadonlyJSONObject | null = null,
  namespace = 'pachyderm/v1',
): Promise<T> {
  // Make request to Jupyter API
  const settings = ServerConnection.makeSettings();
  const requestUrl = URLExt.join(
    settings.baseUrl,
    namespace, // API Namespace
    endPoint,
  );

  const init: RequestInit = {
    method,
    body: body ? JSON.stringify(body) : undefined,
  };

  let response: Response;
  try {
    response = await ServerConnection.makeRequest(requestUrl, init, settings);
  } catch {
    throw new Error();
  }

  let data: any = await response.text();
  if (data.length > 0) {
    try {
      data = JSON.parse(data);
    } catch (error) {
      console.log('Not a JSON response body.');
    }
  }

  if (!response.ok) {
    throw new Error();
  }

  return data;
}
