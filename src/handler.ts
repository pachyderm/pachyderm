import {URLExt} from '@jupyterlab/coreutils';
import {ServerConnection} from '@jupyterlab/services';
import {ReadonlyJSONObject} from '@lumino/coreutils';

export async function requestAPI<T>(
  endPoint = '',
  method = 'GET',
  body: ReadonlyJSONObject | null = null,
  namespace = 'pachyderm/v2',
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
  } catch (error) {
    throw new ServerConnection.NetworkError(error as TypeError);
  }

  let data: any = await response.text();
  if (data.length > 0) {
    try {
      data = JSON.parse(data);
    } catch (error) {
      console.log('Not a JSON response body.', response);
    }
  }

  if (!response.ok) {
    const {message, traceback} = data;
    throw new ServerConnection.ResponseError(
      response,
      message || `Invalid response: ${response.status} ${response.statusText}`,
      traceback || '',
    );
  }

  return data;
}
