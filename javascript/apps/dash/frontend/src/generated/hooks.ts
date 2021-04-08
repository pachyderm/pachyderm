/* eslint-disable @typescript-eslint/naming-convention */

import {gql} from '@apollo/client';
import * as Apollo from '@apollo/client';

import * as Types from '@graphqlTypes';
const defaultOptions = {};

export const ExchangeCodeDocument = gql`
  mutation exchangeCode($code: String!) {
    exchangeCode(code: $code) {
      pachToken
      idToken
    }
  }
`;
export type ExchangeCodeMutationFn = Apollo.MutationFunction<
  Types.ExchangeCodeMutation,
  Types.ExchangeCodeMutationVariables
>;

/**
 * __useExchangeCodeMutation__
 *
 * To run a mutation, you first call `useExchangeCodeMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useExchangeCodeMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [exchangeCodeMutation, { data, loading, error }] = useExchangeCodeMutation({
 *   variables: {
 *      code: // value for 'code'
 *   },
 * });
 */
export function useExchangeCodeMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.ExchangeCodeMutation,
    Types.ExchangeCodeMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.ExchangeCodeMutation,
    Types.ExchangeCodeMutationVariables
  >(ExchangeCodeDocument, options);
}
export type ExchangeCodeMutationHookResult = ReturnType<
  typeof useExchangeCodeMutation
>;
export type ExchangeCodeMutationResult = Apollo.MutationResult<Types.ExchangeCodeMutation>;
export type ExchangeCodeMutationOptions = Apollo.BaseMutationOptions<
  Types.ExchangeCodeMutation,
  Types.ExchangeCodeMutationVariables
>;
export const GetAccountDocument = gql`
  query getAccount {
    account {
      id
      email
      name
    }
  }
`;

/**
 * __useGetAccountQuery__
 *
 * To run a query within a React component, call `useGetAccountQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetAccountQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetAccountQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetAccountQuery(
  baseOptions?: Apollo.QueryHookOptions<
    Types.GetAccountQuery,
    Types.GetAccountQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.GetAccountQuery, Types.GetAccountQueryVariables>(
    GetAccountDocument,
    options,
  );
}
export function useGetAccountLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetAccountQuery,
    Types.GetAccountQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetAccountQuery,
    Types.GetAccountQueryVariables
  >(GetAccountDocument, options);
}
export type GetAccountQueryHookResult = ReturnType<typeof useGetAccountQuery>;
export type GetAccountLazyQueryHookResult = ReturnType<
  typeof useGetAccountLazyQuery
>;
export type GetAccountQueryResult = Apollo.QueryResult<
  Types.GetAccountQuery,
  Types.GetAccountQueryVariables
>;
export const AuthConfigDocument = gql`
  query authConfig {
    authConfig {
      authUrl
      clientId
      pachdClientId
    }
  }
`;

/**
 * __useAuthConfigQuery__
 *
 * To run a query within a React component, call `useAuthConfigQuery` and pass it any options that fit your needs.
 * When your component renders, `useAuthConfigQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAuthConfigQuery({
 *   variables: {
 *   },
 * });
 */
export function useAuthConfigQuery(
  baseOptions?: Apollo.QueryHookOptions<
    Types.AuthConfigQuery,
    Types.AuthConfigQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.AuthConfigQuery, Types.AuthConfigQueryVariables>(
    AuthConfigDocument,
    options,
  );
}
export function useAuthConfigLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.AuthConfigQuery,
    Types.AuthConfigQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.AuthConfigQuery,
    Types.AuthConfigQueryVariables
  >(AuthConfigDocument, options);
}
export type AuthConfigQueryHookResult = ReturnType<typeof useAuthConfigQuery>;
export type AuthConfigLazyQueryHookResult = ReturnType<
  typeof useAuthConfigLazyQuery
>;
export type AuthConfigQueryResult = Apollo.QueryResult<
  Types.AuthConfigQuery,
  Types.AuthConfigQueryVariables
>;
export const GetDagDocument = gql`
  query getDag($args: DagQueryArgs!) {
    dag(args: $args) {
      nodes {
        id
        name
        type
        access
        state
        x
        y
      }
      links {
        id
        source
        target
        sourceState
        targetState
        state
        bendPoints {
          x
          y
        }
        startPoint {
          x
          y
        }
        endPoint {
          x
          y
        }
      }
    }
  }
`;

/**
 * __useGetDagQuery__
 *
 * To run a query within a React component, call `useGetDagQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetDagQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetDagQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetDagQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetDagQuery,
    Types.GetDagQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.GetDagQuery, Types.GetDagQueryVariables>(
    GetDagDocument,
    options,
  );
}
export function useGetDagLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetDagQuery,
    Types.GetDagQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.GetDagQuery, Types.GetDagQueryVariables>(
    GetDagDocument,
    options,
  );
}
export type GetDagQueryHookResult = ReturnType<typeof useGetDagQuery>;
export type GetDagLazyQueryHookResult = ReturnType<typeof useGetDagLazyQuery>;
export type GetDagQueryResult = Apollo.QueryResult<
  Types.GetDagQuery,
  Types.GetDagQueryVariables
>;
export const GetDagsDocument = gql`
  query getDags($args: DagQueryArgs!) {
    dags(args: $args) {
      nodes {
        id
        name
        type
        access
        state
        x
        y
      }
      links {
        id
        source
        target
        sourceState
        targetState
        state
        bendPoints {
          x
          y
        }
        startPoint {
          x
          y
        }
        endPoint {
          x
          y
        }
      }
    }
  }
`;

/**
 * __useGetDagsQuery__
 *
 * To run a query within a React component, call `useGetDagsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetDagsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetDagsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetDagsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetDagsQuery,
    Types.GetDagsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.GetDagsQuery, Types.GetDagsQueryVariables>(
    GetDagsDocument,
    options,
  );
}
export function useGetDagsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetDagsQuery,
    Types.GetDagsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.GetDagsQuery, Types.GetDagsQueryVariables>(
    GetDagsDocument,
    options,
  );
}
export type GetDagsQueryHookResult = ReturnType<typeof useGetDagsQuery>;
export type GetDagsLazyQueryHookResult = ReturnType<typeof useGetDagsLazyQuery>;
export type GetDagsQueryResult = Apollo.QueryResult<
  Types.GetDagsQuery,
  Types.GetDagsQueryVariables
>;
export const GetFilesDocument = gql`
  query getFiles($args: FileQueryArgs!) {
    files(args: $args) {
      committed {
        nanos
        seconds
      }
      commitId
      download
      hash
      path
      repoName
      sizeBytes
      type
    }
  }
`;

/**
 * __useGetFilesQuery__
 *
 * To run a query within a React component, call `useGetFilesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetFilesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetFilesQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetFilesQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetFilesQuery,
    Types.GetFilesQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.GetFilesQuery, Types.GetFilesQueryVariables>(
    GetFilesDocument,
    options,
  );
}
export function useGetFilesLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetFilesQuery,
    Types.GetFilesQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.GetFilesQuery, Types.GetFilesQueryVariables>(
    GetFilesDocument,
    options,
  );
}
export type GetFilesQueryHookResult = ReturnType<typeof useGetFilesQuery>;
export type GetFilesLazyQueryHookResult = ReturnType<
  typeof useGetFilesLazyQuery
>;
export type GetFilesQueryResult = Apollo.QueryResult<
  Types.GetFilesQuery,
  Types.GetFilesQueryVariables
>;
export const GetJobsDocument = gql`
  query getJobs($args: JobQueryArgs!) {
    jobs(args: $args) {
      id
      state
      createdAt
    }
  }
`;

/**
 * __useGetJobsQuery__
 *
 * To run a query within a React component, call `useGetJobsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetJobsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetJobsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetJobsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetJobsQuery,
    Types.GetJobsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.GetJobsQuery, Types.GetJobsQueryVariables>(
    GetJobsDocument,
    options,
  );
}
export function useGetJobsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetJobsQuery,
    Types.GetJobsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.GetJobsQuery, Types.GetJobsQueryVariables>(
    GetJobsDocument,
    options,
  );
}
export type GetJobsQueryHookResult = ReturnType<typeof useGetJobsQuery>;
export type GetJobsLazyQueryHookResult = ReturnType<typeof useGetJobsLazyQuery>;
export type GetJobsQueryResult = Apollo.QueryResult<
  Types.GetJobsQuery,
  Types.GetJobsQueryVariables
>;
export const LoggedInDocument = gql`
  query loggedIn {
    loggedIn @client
  }
`;

/**
 * __useLoggedInQuery__
 *
 * To run a query within a React component, call `useLoggedInQuery` and pass it any options that fit your needs.
 * When your component renders, `useLoggedInQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useLoggedInQuery({
 *   variables: {
 *   },
 * });
 */
export function useLoggedInQuery(
  baseOptions?: Apollo.QueryHookOptions<
    Types.LoggedInQuery,
    Types.LoggedInQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.LoggedInQuery, Types.LoggedInQueryVariables>(
    LoggedInDocument,
    options,
  );
}
export function useLoggedInLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.LoggedInQuery,
    Types.LoggedInQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.LoggedInQuery, Types.LoggedInQueryVariables>(
    LoggedInDocument,
    options,
  );
}
export type LoggedInQueryHookResult = ReturnType<typeof useLoggedInQuery>;
export type LoggedInLazyQueryHookResult = ReturnType<
  typeof useLoggedInLazyQuery
>;
export type LoggedInQueryResult = Apollo.QueryResult<
  Types.LoggedInQuery,
  Types.LoggedInQueryVariables
>;
export const ProjectDetailsDocument = gql`
  query projectDetails($args: ProjectDetailsQueryArgs!) {
    projectDetails(args: $args) {
      sizeGBytes
      repoCount
      pipelineCount
      jobs {
        id
        state
        createdAt
      }
    }
  }
`;

/**
 * __useProjectDetailsQuery__
 *
 * To run a query within a React component, call `useProjectDetailsQuery` and pass it any options that fit your needs.
 * When your component renders, `useProjectDetailsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProjectDetailsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useProjectDetailsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.ProjectDetailsQuery,
    Types.ProjectDetailsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.ProjectDetailsQuery,
    Types.ProjectDetailsQueryVariables
  >(ProjectDetailsDocument, options);
}
export function useProjectDetailsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.ProjectDetailsQuery,
    Types.ProjectDetailsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.ProjectDetailsQuery,
    Types.ProjectDetailsQueryVariables
  >(ProjectDetailsDocument, options);
}
export type ProjectDetailsQueryHookResult = ReturnType<
  typeof useProjectDetailsQuery
>;
export type ProjectDetailsLazyQueryHookResult = ReturnType<
  typeof useProjectDetailsLazyQuery
>;
export type ProjectDetailsQueryResult = Apollo.QueryResult<
  Types.ProjectDetailsQuery,
  Types.ProjectDetailsQueryVariables
>;
export const ProjectDocument = gql`
  query project($id: ID!) {
    project(id: $id) {
      id
      name
      description
      createdAt
      status
    }
  }
`;

/**
 * __useProjectQuery__
 *
 * To run a query within a React component, call `useProjectQuery` and pass it any options that fit your needs.
 * When your component renders, `useProjectQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProjectQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useProjectQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.ProjectQuery,
    Types.ProjectQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.ProjectQuery, Types.ProjectQueryVariables>(
    ProjectDocument,
    options,
  );
}
export function useProjectLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.ProjectQuery,
    Types.ProjectQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.ProjectQuery, Types.ProjectQueryVariables>(
    ProjectDocument,
    options,
  );
}
export type ProjectQueryHookResult = ReturnType<typeof useProjectQuery>;
export type ProjectLazyQueryHookResult = ReturnType<typeof useProjectLazyQuery>;
export type ProjectQueryResult = Apollo.QueryResult<
  Types.ProjectQuery,
  Types.ProjectQueryVariables
>;
export const ProjectsDocument = gql`
  query projects {
    projects {
      id
      name
      description
      createdAt
      status
    }
  }
`;

/**
 * __useProjectsQuery__
 *
 * To run a query within a React component, call `useProjectsQuery` and pass it any options that fit your needs.
 * When your component renders, `useProjectsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProjectsQuery({
 *   variables: {
 *   },
 * });
 */
export function useProjectsQuery(
  baseOptions?: Apollo.QueryHookOptions<
    Types.ProjectsQuery,
    Types.ProjectsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.ProjectsQuery, Types.ProjectsQueryVariables>(
    ProjectsDocument,
    options,
  );
}
export function useProjectsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.ProjectsQuery,
    Types.ProjectsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.ProjectsQuery, Types.ProjectsQueryVariables>(
    ProjectsDocument,
    options,
  );
}
export type ProjectsQueryHookResult = ReturnType<typeof useProjectsQuery>;
export type ProjectsLazyQueryHookResult = ReturnType<
  typeof useProjectsLazyQuery
>;
export type ProjectsQueryResult = Apollo.QueryResult<
  Types.ProjectsQuery,
  Types.ProjectsQueryVariables
>;
