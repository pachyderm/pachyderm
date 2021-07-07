/* eslint-disable @typescript-eslint/naming-convention */

import {gql} from '@apollo/client';
import * as Apollo from '@apollo/client';

import * as Types from '@graphqlTypes';
const defaultOptions = {};
export const JobOverviewFragmentDoc = gql`
  fragment JobOverview on Job {
    id
    state
    createdAt
    finishedAt
    pipelineName
  }
`;
export const JobSetFieldsFragmentDoc = gql`
  fragment JobSetFields on JobSet {
    id
    state
    createdAt
    jobs {
      ...JobOverview
      inputString
      inputBranch
      transform {
        cmdList
        image
      }
    }
  }
  ${JobOverviewFragmentDoc}
`;
export const LogFieldsFragmentDoc = gql`
  fragment LogFields on Log {
    timestamp {
      seconds
      nanos
    }
    user
    message
  }
`;
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
export type ExchangeCodeMutationResult =
  Apollo.MutationResult<Types.ExchangeCodeMutation>;
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
      id
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
  subscription getDags($args: DagQueryArgs!) {
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
        transferring
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
      id
    }
  }
`;

/**
 * __useGetDagsSubscription__
 *
 * To run a query within a React component, call `useGetDagsSubscription` and pass it any options that fit your needs.
 * When your component renders, `useGetDagsSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetDagsSubscription({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetDagsSubscription(
  baseOptions: Apollo.SubscriptionHookOptions<
    Types.GetDagsSubscription,
    Types.GetDagsSubscriptionVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useSubscription<
    Types.GetDagsSubscription,
    Types.GetDagsSubscriptionVariables
  >(GetDagsDocument, options);
}
export type GetDagsSubscriptionHookResult = ReturnType<
  typeof useGetDagsSubscription
>;
export type GetDagsSubscriptionResult =
  Apollo.SubscriptionResult<Types.GetDagsSubscription>;
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
      sizeDisplay
      downloadDisabled
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
export const JobDocument = gql`
  query job($args: JobQueryArgs!) {
    job(args: $args) {
      ...JobOverview
      inputString
      inputBranch
      transform {
        cmdList
        image
      }
    }
  }
  ${JobOverviewFragmentDoc}
`;

/**
 * __useJobQuery__
 *
 * To run a query within a React component, call `useJobQuery` and pass it any options that fit your needs.
 * When your component renders, `useJobQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useJobQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useJobQuery(
  baseOptions: Apollo.QueryHookOptions<Types.JobQuery, Types.JobQueryVariables>,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.JobQuery, Types.JobQueryVariables>(
    JobDocument,
    options,
  );
}
export function useJobLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.JobQuery,
    Types.JobQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.JobQuery, Types.JobQueryVariables>(
    JobDocument,
    options,
  );
}
export type JobQueryHookResult = ReturnType<typeof useJobQuery>;
export type JobLazyQueryHookResult = ReturnType<typeof useJobLazyQuery>;
export type JobQueryResult = Apollo.QueryResult<
  Types.JobQuery,
  Types.JobQueryVariables
>;
export const JobsDocument = gql`
  query jobs($args: JobsQueryArgs!) {
    jobs(args: $args) {
      ...JobOverview
      inputString
      inputBranch
      transform {
        cmdList
        image
      }
    }
  }
  ${JobOverviewFragmentDoc}
`;

/**
 * __useJobsQuery__
 *
 * To run a query within a React component, call `useJobsQuery` and pass it any options that fit your needs.
 * When your component renders, `useJobsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useJobsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useJobsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.JobsQuery,
    Types.JobsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.JobsQuery, Types.JobsQueryVariables>(
    JobsDocument,
    options,
  );
}
export function useJobsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.JobsQuery,
    Types.JobsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.JobsQuery, Types.JobsQueryVariables>(
    JobsDocument,
    options,
  );
}
export type JobsQueryHookResult = ReturnType<typeof useJobsQuery>;
export type JobsLazyQueryHookResult = ReturnType<typeof useJobsLazyQuery>;
export type JobsQueryResult = Apollo.QueryResult<
  Types.JobsQuery,
  Types.JobsQueryVariables
>;
export const JobSetDocument = gql`
  query jobSet($args: JobSetQueryArgs!) {
    jobSet(args: $args) {
      ...JobSetFields
    }
  }
  ${JobSetFieldsFragmentDoc}
`;

/**
 * __useJobSetQuery__
 *
 * To run a query within a React component, call `useJobSetQuery` and pass it any options that fit your needs.
 * When your component renders, `useJobSetQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useJobSetQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useJobSetQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.JobSetQuery,
    Types.JobSetQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.JobSetQuery, Types.JobSetQueryVariables>(
    JobSetDocument,
    options,
  );
}
export function useJobSetLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.JobSetQuery,
    Types.JobSetQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.JobSetQuery, Types.JobSetQueryVariables>(
    JobSetDocument,
    options,
  );
}
export type JobSetQueryHookResult = ReturnType<typeof useJobSetQuery>;
export type JobSetLazyQueryHookResult = ReturnType<typeof useJobSetLazyQuery>;
export type JobSetQueryResult = Apollo.QueryResult<
  Types.JobSetQuery,
  Types.JobSetQueryVariables
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
export const GetWorkspaceLogsDocument = gql`
  query getWorkspaceLogs($args: WorkspaceLogsArgs!) {
    workspaceLogs(args: $args) {
      ...LogFields
    }
  }
  ${LogFieldsFragmentDoc}
`;

/**
 * __useGetWorkspaceLogsQuery__
 *
 * To run a query within a React component, call `useGetWorkspaceLogsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetWorkspaceLogsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetWorkspaceLogsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetWorkspaceLogsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetWorkspaceLogsQuery,
    Types.GetWorkspaceLogsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.GetWorkspaceLogsQuery,
    Types.GetWorkspaceLogsQueryVariables
  >(GetWorkspaceLogsDocument, options);
}
export function useGetWorkspaceLogsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetWorkspaceLogsQuery,
    Types.GetWorkspaceLogsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetWorkspaceLogsQuery,
    Types.GetWorkspaceLogsQueryVariables
  >(GetWorkspaceLogsDocument, options);
}
export type GetWorkspaceLogsQueryHookResult = ReturnType<
  typeof useGetWorkspaceLogsQuery
>;
export type GetWorkspaceLogsLazyQueryHookResult = ReturnType<
  typeof useGetWorkspaceLogsLazyQuery
>;
export type GetWorkspaceLogsQueryResult = Apollo.QueryResult<
  Types.GetWorkspaceLogsQuery,
  Types.GetWorkspaceLogsQueryVariables
>;
export const GetLogsDocument = gql`
  query getLogs($args: LogsArgs!) {
    logs(args: $args) {
      ...LogFields
    }
  }
  ${LogFieldsFragmentDoc}
`;

/**
 * __useGetLogsQuery__
 *
 * To run a query within a React component, call `useGetLogsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetLogsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetLogsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetLogsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetLogsQuery,
    Types.GetLogsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.GetLogsQuery, Types.GetLogsQueryVariables>(
    GetLogsDocument,
    options,
  );
}
export function useGetLogsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetLogsQuery,
    Types.GetLogsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.GetLogsQuery, Types.GetLogsQueryVariables>(
    GetLogsDocument,
    options,
  );
}
export type GetLogsQueryHookResult = ReturnType<typeof useGetLogsQuery>;
export type GetLogsLazyQueryHookResult = ReturnType<typeof useGetLogsLazyQuery>;
export type GetLogsQueryResult = Apollo.QueryResult<
  Types.GetLogsQuery,
  Types.GetLogsQueryVariables
>;
export const GetWorkspaceLogStreamDocument = gql`
  subscription getWorkspaceLogStream($args: WorkspaceLogsArgs!) {
    workspaceLogs(args: $args) {
      ...LogFields
    }
  }
  ${LogFieldsFragmentDoc}
`;

/**
 * __useGetWorkspaceLogStreamSubscription__
 *
 * To run a query within a React component, call `useGetWorkspaceLogStreamSubscription` and pass it any options that fit your needs.
 * When your component renders, `useGetWorkspaceLogStreamSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetWorkspaceLogStreamSubscription({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetWorkspaceLogStreamSubscription(
  baseOptions: Apollo.SubscriptionHookOptions<
    Types.GetWorkspaceLogStreamSubscription,
    Types.GetWorkspaceLogStreamSubscriptionVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useSubscription<
    Types.GetWorkspaceLogStreamSubscription,
    Types.GetWorkspaceLogStreamSubscriptionVariables
  >(GetWorkspaceLogStreamDocument, options);
}
export type GetWorkspaceLogStreamSubscriptionHookResult = ReturnType<
  typeof useGetWorkspaceLogStreamSubscription
>;
export type GetWorkspaceLogStreamSubscriptionResult =
  Apollo.SubscriptionResult<Types.GetWorkspaceLogStreamSubscription>;
export const GetLogsStreamDocument = gql`
  subscription getLogsStream($args: LogsArgs!) {
    logs(args: $args) {
      ...LogFields
    }
  }
  ${LogFieldsFragmentDoc}
`;

/**
 * __useGetLogsStreamSubscription__
 *
 * To run a query within a React component, call `useGetLogsStreamSubscription` and pass it any options that fit your needs.
 * When your component renders, `useGetLogsStreamSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetLogsStreamSubscription({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetLogsStreamSubscription(
  baseOptions: Apollo.SubscriptionHookOptions<
    Types.GetLogsStreamSubscription,
    Types.GetLogsStreamSubscriptionVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useSubscription<
    Types.GetLogsStreamSubscription,
    Types.GetLogsStreamSubscriptionVariables
  >(GetLogsStreamDocument, options);
}
export type GetLogsStreamSubscriptionHookResult = ReturnType<
  typeof useGetLogsStreamSubscription
>;
export type GetLogsStreamSubscriptionResult =
  Apollo.SubscriptionResult<Types.GetLogsStreamSubscription>;
export const PipelineDocument = gql`
  query pipeline($args: PipelineQueryArgs!) {
    pipeline(args: $args) {
      id
      name
      state
      type
      description
      transform {
        cmdList
        image
      }
      inputString
      cacheSize
      datumTimeoutS
      datumTries
      jobTimeoutS
      outputBranch
      s3OutputRepo
      egress
      schedulingSpec {
        nodeSelectorMap {
          key
          value
        }
        priorityClassName
      }
    }
  }
`;

/**
 * __usePipelineQuery__
 *
 * To run a query within a React component, call `usePipelineQuery` and pass it any options that fit your needs.
 * When your component renders, `usePipelineQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = usePipelineQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function usePipelineQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.PipelineQuery,
    Types.PipelineQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.PipelineQuery, Types.PipelineQueryVariables>(
    PipelineDocument,
    options,
  );
}
export function usePipelineLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.PipelineQuery,
    Types.PipelineQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.PipelineQuery, Types.PipelineQueryVariables>(
    PipelineDocument,
    options,
  );
}
export type PipelineQueryHookResult = ReturnType<typeof usePipelineQuery>;
export type PipelineLazyQueryHookResult = ReturnType<
  typeof usePipelineLazyQuery
>;
export type PipelineQueryResult = Apollo.QueryResult<
  Types.PipelineQuery,
  Types.PipelineQueryVariables
>;
export const ProjectDetailsDocument = gql`
  query projectDetails($args: ProjectDetailsQueryArgs!) {
    projectDetails(args: $args) {
      sizeDisplay
      repoCount
      pipelineCount
      jobSets {
        ...JobSetFields
      }
    }
  }
  ${JobSetFieldsFragmentDoc}
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
export const RepoDocument = gql`
  query repo($args: RepoQueryArgs!) {
    repo(args: $args) {
      branches {
        id
        name
      }
      commits {
        repoName
        branch {
          id
          name
        }
        description
        id
        started
        finished
        sizeDisplay
      }
      createdAt
      description
      id
      linkedPipeline {
        id
        name
      }
      name
      sizeDisplay
    }
  }
`;

/**
 * __useRepoQuery__
 *
 * To run a query within a React component, call `useRepoQuery` and pass it any options that fit your needs.
 * When your component renders, `useRepoQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useRepoQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useRepoQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.RepoQuery,
    Types.RepoQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.RepoQuery, Types.RepoQueryVariables>(
    RepoDocument,
    options,
  );
}
export function useRepoLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.RepoQuery,
    Types.RepoQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.RepoQuery, Types.RepoQueryVariables>(
    RepoDocument,
    options,
  );
}
export type RepoQueryHookResult = ReturnType<typeof useRepoQuery>;
export type RepoLazyQueryHookResult = ReturnType<typeof useRepoLazyQuery>;
export type RepoQueryResult = Apollo.QueryResult<
  Types.RepoQuery,
  Types.RepoQueryVariables
>;
export const SearchResultsDocument = gql`
  query searchResults($args: SearchResultQueryArgs!) {
    searchResults(args: $args) {
      pipelines {
        name
        id
      }
      repos {
        name
        id
      }
      jobSet {
        id
      }
    }
  }
`;

/**
 * __useSearchResultsQuery__
 *
 * To run a query within a React component, call `useSearchResultsQuery` and pass it any options that fit your needs.
 * When your component renders, `useSearchResultsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSearchResultsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useSearchResultsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.SearchResultsQuery,
    Types.SearchResultsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.SearchResultsQuery,
    Types.SearchResultsQueryVariables
  >(SearchResultsDocument, options);
}
export function useSearchResultsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.SearchResultsQuery,
    Types.SearchResultsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.SearchResultsQuery,
    Types.SearchResultsQueryVariables
  >(SearchResultsDocument, options);
}
export type SearchResultsQueryHookResult = ReturnType<
  typeof useSearchResultsQuery
>;
export type SearchResultsLazyQueryHookResult = ReturnType<
  typeof useSearchResultsLazyQuery
>;
export type SearchResultsQueryResult = Apollo.QueryResult<
  Types.SearchResultsQuery,
  Types.SearchResultsQueryVariables
>;
