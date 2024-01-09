import {API as adminAPI} from '@dash-frontend/generated/proto/admin/admin.pb';
import {API as authAPI} from '@dash-frontend/generated/proto/auth/auth.pb';
import {API as enterpriseAPI} from '@dash-frontend/generated/proto/enterprise/enterprise.pb';
import {API as pfsAPI} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {API as ppsAPI} from '@dash-frontend/generated/proto/pps/pps.pb';
import {API as versionAPI} from '@dash-frontend/generated/proto/version/versionpb/version.pb';

import {isPermissionDenied} from '../utils/error';

const ignorePermissionDeniedProxy = <T extends {[key: string]: any}>(
  targetClass: T,
): T => {
  return new Proxy(targetClass, {
    get(target, propKey, _receiver) {
      const originalProperty = target[propKey as any];
      if (typeof originalProperty === 'function') {
        return (...args: any[]) => {
          try {
            const result = originalProperty.apply(this, args);
            if (result instanceof Promise) {
              return result.catch((error: any) => {
                // Asynchronous errors here
                if (isPermissionDenied(error)) return null;
                throw error;
              });
            }
            return result;
          } catch (error) {
            // Synchronous errors here
            if (isPermissionDenied(error)) return null;
            throw error;
          }
        };
      }
      return originalProperty;
    },
  });
};

export const admin = adminAPI;
export const auth = authAPI;
export const pps = ignorePermissionDeniedProxy(ppsAPI);
export const pfs = ignorePermissionDeniedProxy(pfsAPI);
export const enterprise = enterpriseAPI;
export const version = versionAPI;
