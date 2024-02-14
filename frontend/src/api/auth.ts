import {
  AuthorizeRequest,
  GetRoleBindingRequest,
  ModifyRoleBindingRequest,
} from '@dash-frontend/generated/proto/auth/auth.pb';

import {auth} from './base/api';
import {isAuthDisabled, RequestError} from './utils/error';
import {getRequestOptions, getHeaders} from './utils/requestHeaders';

export type Account = {
  email: string;
  name?: string;
  id: string;
};

export type Exchange = {
  idToken?: string;
};

export type AuthConfig = {
  authEndpoint: string;
  clientId: string;
  pachdClientId: string;
};

export const authorize = async (req: AuthorizeRequest) => {
  try {
    const response = await auth.Authorize(req, getRequestOptions());

    return response;
  } catch (error) {
    const authDisabled = isAuthDisabled(error);

    if (authDisabled) {
      return {
        satisfied: [],
        missing: [],
        principal: '',
        authorized: null,
      };
    }

    throw error;
  }
};

export const getRoleBinding = async (req: GetRoleBindingRequest) => {
  try {
    const response = await auth.GetRoleBinding(req, getRequestOptions());

    return response;
  } catch (error) {
    const authDisabled = isAuthDisabled(error);

    if (authDisabled) {
      return;
    }

    throw error;
  }
};

export const modifyRoleBinding = async (req: ModifyRoleBindingRequest) => {
  try {
    const response = await auth.ModifyRoleBinding(req, getRequestOptions());

    return response;
  } catch (error) {
    const authDisabled = isAuthDisabled(error);

    if (authDisabled) {
      return;
    }

    throw error;
  }
};

export const authenticate = async (code: string) => {
  const {idToken, message} = (await fetch('/auth/exchange', {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({code}),
  }).then((res) => res.json())) as Exchange & RequestError;

  if (!idToken) {
    throw new Error(message);
  }

  try {
    const {pachToken} = await auth.Authenticate({idToken}, getRequestOptions());

    if (!pachToken) {
      throw new Error('Authenticate Error: Did not receive a pachToken.');
    }

    return {
      idToken,
      pachToken,
    };
  } catch (error) {
    const authDisabled = isAuthDisabled(error);

    if (authDisabled) {
      return {
        idToken: '',
        pachToken: '',
      };
    }

    throw error;
  }
};

export const config = async () => {
  // Make sure auth is enabled before asking for the config
  try {
    await auth.WhoAmI({}, getRequestOptions());
  } catch (error) {
    const authDisabled = isAuthDisabled(error);

    if (authDisabled) {
      return {
        authEndpoint: '',
        clientId: '',
        pachdClientId: '',
      } as AuthConfig;
    }
  }

  const res = await fetch('/auth/config', {
    method: 'GET',
    headers: getHeaders(),
  });

  const {message, authEndpoint, clientId, pachdClientId} =
    (await res.json()) as AuthConfig & RequestError;

  if (!res.ok || message) {
    throw new Error(message);
  }

  return {
    authEndpoint,
    clientId,
    pachdClientId,
  };
};

export const account = async () => {
  const idToken = localStorage.getItem('id-token');

  if (idToken) {
    return (await fetch('/auth/account', {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({idToken}),
    }).then((res) => res.json())) as Account;
  } else {
    try {
      await auth.WhoAmI({}, getRequestOptions());
    } catch (error) {
      const authDisabled = isAuthDisabled(error);

      if (authDisabled) {
        return {
          id: 'unauthenticated',
          email: '',
          name: 'User',
        } as Account;
      }

      throw new Error('Authentication Error: Could not retrieve an account.');
    }
  }
};

// Export all of the auth types
export * from '@dash-frontend/generated/proto/auth/auth.pb';
