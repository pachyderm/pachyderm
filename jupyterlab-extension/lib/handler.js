var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { URLExt } from '@jupyterlab/coreutils';
import { ServerConnection } from '@jupyterlab/services';
/**
 * Call the API extension
 *
 * @param endPoint API REST end point for the extension
 * @param init Initial values for the request
 * @returns The response body interpreted as JSON
 */
export function requestAPI(endPoint = '', init = {}) {
    return __awaiter(this, void 0, void 0, function* () {
        // Make request to Jupyter API
        const settings = ServerConnection.makeSettings();
        const requestUrl = URLExt.join(settings.baseUrl, 'pachyderm', // API Namespace
        endPoint);
        let response;
        try {
            response = yield ServerConnection.makeRequest(requestUrl, init, settings);
        }
        catch (error) {
            throw new ServerConnection.NetworkError(error);
        }
        let data = yield response.text();
        if (data.length > 0) {
            try {
                data = JSON.parse(data);
            }
            catch (error) {
                console.log('Not a JSON response body.', response);
            }
        }
        if (!response.ok) {
            throw new ServerConnection.ResponseError(response, data.message || data);
        }
        return data;
    });
}
