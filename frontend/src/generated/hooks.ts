import {gql} from '@apollo/client';
import * as Apollo from '@apollo/client';
import * as Types from '@graphqlTypes';
const defaultOptions = {} as const;
export const CommitFragmentFragmentDoc = gql`
  fragment CommitFragment on Commit {
    repoName
    branch {
      name
    }
    description
    originKind
    id
    started
    finished
    sizeBytes
    sizeDisplay
    hasLinkedJob
  }
`;
export const DatumFragmentDoc = gql`
  fragment Datum on Datum {
    id
    state
    downloadBytes
    uploadTime
    processTime
    downloadTime
  }
`;
export const DiffFragmentFragmentDoc = gql`
  fragment DiffFragment on Diff {
    size
    sizeDisplay
    filesUpdated
    filesAdded
    filesDeleted
  }
`;
export const JobOverviewFragmentDoc = gql`
  fragment JobOverview on Job {
    id
    state
    createdAt
    startedAt
    finishedAt
    pipelineName
    reason
    dataProcessed
    dataSkipped
    dataFailed
    dataTotal
    dataRecovered
    outputCommit
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
      transformString
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
export const CreateBranchDocument = gql`
  mutation createBranch($args: CreateBranchArgs!) {
    createBranch(args: $args) {
      name
      repo {
        name
      }
    }
  }
`;
export type CreateBranchMutationFn = Apollo.MutationFunction<
  Types.CreateBranchMutation,
  Types.CreateBranchMutationVariables
>;

/**
 * __useCreateBranchMutation__
 *
 * To run a mutation, you first call `useCreateBranchMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateBranchMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createBranchMutation, { data, loading, error }] = useCreateBranchMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useCreateBranchMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.CreateBranchMutation,
    Types.CreateBranchMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.CreateBranchMutation,
    Types.CreateBranchMutationVariables
  >(CreateBranchDocument, options);
}
export type CreateBranchMutationHookResult = ReturnType<
  typeof useCreateBranchMutation
>;
export type CreateBranchMutationResult =
  Apollo.MutationResult<Types.CreateBranchMutation>;
export type CreateBranchMutationOptions = Apollo.BaseMutationOptions<
  Types.CreateBranchMutation,
  Types.CreateBranchMutationVariables
>;
export const CreatePipelineDocument = gql`
  mutation createPipeline($args: CreatePipelineArgs!) {
    createPipeline(args: $args) {
      id
      name
      state
      type
      description
      datumTimeoutS
      datumTries
      jobTimeoutS
      outputBranch
      s3OutputRepo
      egress
      jsonSpec
      reason
    }
  }
`;
export type CreatePipelineMutationFn = Apollo.MutationFunction<
  Types.CreatePipelineMutation,
  Types.CreatePipelineMutationVariables
>;

/**
 * __useCreatePipelineMutation__
 *
 * To run a mutation, you first call `useCreatePipelineMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreatePipelineMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createPipelineMutation, { data, loading, error }] = useCreatePipelineMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useCreatePipelineMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.CreatePipelineMutation,
    Types.CreatePipelineMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.CreatePipelineMutation,
    Types.CreatePipelineMutationVariables
  >(CreatePipelineDocument, options);
}
export type CreatePipelineMutationHookResult = ReturnType<
  typeof useCreatePipelineMutation
>;
export type CreatePipelineMutationResult =
  Apollo.MutationResult<Types.CreatePipelineMutation>;
export type CreatePipelineMutationOptions = Apollo.BaseMutationOptions<
  Types.CreatePipelineMutation,
  Types.CreatePipelineMutationVariables
>;
export const CreateRepoDocument = gql`
  mutation createRepo($args: CreateRepoArgs!) {
    createRepo(args: $args) {
      createdAt
      description
      id
      name
      sizeDisplay
    }
  }
`;
export type CreateRepoMutationFn = Apollo.MutationFunction<
  Types.CreateRepoMutation,
  Types.CreateRepoMutationVariables
>;

/**
 * __useCreateRepoMutation__
 *
 * To run a mutation, you first call `useCreateRepoMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateRepoMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createRepoMutation, { data, loading, error }] = useCreateRepoMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useCreateRepoMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.CreateRepoMutation,
    Types.CreateRepoMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.CreateRepoMutation,
    Types.CreateRepoMutationVariables
  >(CreateRepoDocument, options);
}
export type CreateRepoMutationHookResult = ReturnType<
  typeof useCreateRepoMutation
>;
export type CreateRepoMutationResult =
  Apollo.MutationResult<Types.CreateRepoMutation>;
export type CreateRepoMutationOptions = Apollo.BaseMutationOptions<
  Types.CreateRepoMutation,
  Types.CreateRepoMutationVariables
>;
export const DeleteFileDocument = gql`
  mutation deleteFile($args: DeleteFileArgs!) {
    deleteFile(args: $args)
  }
`;
export type DeleteFileMutationFn = Apollo.MutationFunction<
  Types.DeleteFileMutation,
  Types.DeleteFileMutationVariables
>;

/**
 * __useDeleteFileMutation__
 *
 * To run a mutation, you first call `useDeleteFileMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteFileMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteFileMutation, { data, loading, error }] = useDeleteFileMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useDeleteFileMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.DeleteFileMutation,
    Types.DeleteFileMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.DeleteFileMutation,
    Types.DeleteFileMutationVariables
  >(DeleteFileDocument, options);
}
export type DeleteFileMutationHookResult = ReturnType<
  typeof useDeleteFileMutation
>;
export type DeleteFileMutationResult =
  Apollo.MutationResult<Types.DeleteFileMutation>;
export type DeleteFileMutationOptions = Apollo.BaseMutationOptions<
  Types.DeleteFileMutation,
  Types.DeleteFileMutationVariables
>;
export const DeletePipelineDocument = gql`
  mutation deletePipeline($args: DeletePipelineArgs!) {
    deletePipeline(args: $args)
  }
`;
export type DeletePipelineMutationFn = Apollo.MutationFunction<
  Types.DeletePipelineMutation,
  Types.DeletePipelineMutationVariables
>;

/**
 * __useDeletePipelineMutation__
 *
 * To run a mutation, you first call `useDeletePipelineMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeletePipelineMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deletePipelineMutation, { data, loading, error }] = useDeletePipelineMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useDeletePipelineMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.DeletePipelineMutation,
    Types.DeletePipelineMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.DeletePipelineMutation,
    Types.DeletePipelineMutationVariables
  >(DeletePipelineDocument, options);
}
export type DeletePipelineMutationHookResult = ReturnType<
  typeof useDeletePipelineMutation
>;
export type DeletePipelineMutationResult =
  Apollo.MutationResult<Types.DeletePipelineMutation>;
export type DeletePipelineMutationOptions = Apollo.BaseMutationOptions<
  Types.DeletePipelineMutation,
  Types.DeletePipelineMutationVariables
>;
export const DeleteRepoDocument = gql`
  mutation deleteRepo($args: DeleteRepoArgs!) {
    deleteRepo(args: $args)
  }
`;
export type DeleteRepoMutationFn = Apollo.MutationFunction<
  Types.DeleteRepoMutation,
  Types.DeleteRepoMutationVariables
>;

/**
 * __useDeleteRepoMutation__
 *
 * To run a mutation, you first call `useDeleteRepoMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteRepoMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteRepoMutation, { data, loading, error }] = useDeleteRepoMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useDeleteRepoMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.DeleteRepoMutation,
    Types.DeleteRepoMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.DeleteRepoMutation,
    Types.DeleteRepoMutationVariables
  >(DeleteRepoDocument, options);
}
export type DeleteRepoMutationHookResult = ReturnType<
  typeof useDeleteRepoMutation
>;
export type DeleteRepoMutationResult =
  Apollo.MutationResult<Types.DeleteRepoMutation>;
export type DeleteRepoMutationOptions = Apollo.BaseMutationOptions<
  Types.DeleteRepoMutation,
  Types.DeleteRepoMutationVariables
>;
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
export const FinishCommitDocument = gql`
  mutation finishCommit($args: FinishCommitArgs!) {
    finishCommit(args: $args)
  }
`;
export type FinishCommitMutationFn = Apollo.MutationFunction<
  Types.FinishCommitMutation,
  Types.FinishCommitMutationVariables
>;

/**
 * __useFinishCommitMutation__
 *
 * To run a mutation, you first call `useFinishCommitMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useFinishCommitMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [finishCommitMutation, { data, loading, error }] = useFinishCommitMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useFinishCommitMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.FinishCommitMutation,
    Types.FinishCommitMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.FinishCommitMutation,
    Types.FinishCommitMutationVariables
  >(FinishCommitDocument, options);
}
export type FinishCommitMutationHookResult = ReturnType<
  typeof useFinishCommitMutation
>;
export type FinishCommitMutationResult =
  Apollo.MutationResult<Types.FinishCommitMutation>;
export type FinishCommitMutationOptions = Apollo.BaseMutationOptions<
  Types.FinishCommitMutation,
  Types.FinishCommitMutationVariables
>;
export const PutFilesFromUrLsDocument = gql`
  mutation putFilesFromURLs($args: PutFilesFromURLsArgs!) {
    putFilesFromURLs(args: $args)
  }
`;
export type PutFilesFromUrLsMutationFn = Apollo.MutationFunction<
  Types.PutFilesFromUrLsMutation,
  Types.PutFilesFromUrLsMutationVariables
>;

/**
 * __usePutFilesFromUrLsMutation__
 *
 * To run a mutation, you first call `usePutFilesFromUrLsMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `usePutFilesFromUrLsMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [putFilesFromUrLsMutation, { data, loading, error }] = usePutFilesFromUrLsMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function usePutFilesFromUrLsMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.PutFilesFromUrLsMutation,
    Types.PutFilesFromUrLsMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.PutFilesFromUrLsMutation,
    Types.PutFilesFromUrLsMutationVariables
  >(PutFilesFromUrLsDocument, options);
}
export type PutFilesFromUrLsMutationHookResult = ReturnType<
  typeof usePutFilesFromUrLsMutation
>;
export type PutFilesFromUrLsMutationResult =
  Apollo.MutationResult<Types.PutFilesFromUrLsMutation>;
export type PutFilesFromUrLsMutationOptions = Apollo.BaseMutationOptions<
  Types.PutFilesFromUrLsMutation,
  Types.PutFilesFromUrLsMutationVariables
>;
export const StartCommitDocument = gql`
  mutation startCommit($args: StartCommitArgs!) {
    startCommit(args: $args) {
      branch {
        name
        repo {
          name
          type
        }
      }
      id
    }
  }
`;
export type StartCommitMutationFn = Apollo.MutationFunction<
  Types.StartCommitMutation,
  Types.StartCommitMutationVariables
>;

/**
 * __useStartCommitMutation__
 *
 * To run a mutation, you first call `useStartCommitMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useStartCommitMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [startCommitMutation, { data, loading, error }] = useStartCommitMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useStartCommitMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.StartCommitMutation,
    Types.StartCommitMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.StartCommitMutation,
    Types.StartCommitMutationVariables
  >(StartCommitDocument, options);
}
export type StartCommitMutationHookResult = ReturnType<
  typeof useStartCommitMutation
>;
export type StartCommitMutationResult =
  Apollo.MutationResult<Types.StartCommitMutation>;
export type StartCommitMutationOptions = Apollo.BaseMutationOptions<
  Types.StartCommitMutation,
  Types.StartCommitMutationVariables
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
      authEndpoint
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
export const CommitDocument = gql`
  query commit($args: CommitQueryArgs!) {
    commit(args: $args) {
      ...CommitFragment
      diff {
        ...DiffFragment
      }
    }
  }
  ${CommitFragmentFragmentDoc}
  ${DiffFragmentFragmentDoc}
`;

/**
 * __useCommitQuery__
 *
 * To run a query within a React component, call `useCommitQuery` and pass it any options that fit your needs.
 * When your component renders, `useCommitQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCommitQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useCommitQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.CommitQuery,
    Types.CommitQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.CommitQuery, Types.CommitQueryVariables>(
    CommitDocument,
    options,
  );
}
export function useCommitLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.CommitQuery,
    Types.CommitQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.CommitQuery, Types.CommitQueryVariables>(
    CommitDocument,
    options,
  );
}
export type CommitQueryHookResult = ReturnType<typeof useCommitQuery>;
export type CommitLazyQueryHookResult = ReturnType<typeof useCommitLazyQuery>;
export type CommitQueryResult = Apollo.QueryResult<
  Types.CommitQuery,
  Types.CommitQueryVariables
>;
export const GetCommitsDocument = gql`
  query getCommits($args: CommitsQueryArgs!) {
    commits(args: $args) {
      ...CommitFragment
    }
  }
  ${CommitFragmentFragmentDoc}
`;

/**
 * __useGetCommitsQuery__
 *
 * To run a query within a React component, call `useGetCommitsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetCommitsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetCommitsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetCommitsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetCommitsQuery,
    Types.GetCommitsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.GetCommitsQuery, Types.GetCommitsQueryVariables>(
    GetCommitsDocument,
    options,
  );
}
export function useGetCommitsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetCommitsQuery,
    Types.GetCommitsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetCommitsQuery,
    Types.GetCommitsQueryVariables
  >(GetCommitsDocument, options);
}
export type GetCommitsQueryHookResult = ReturnType<typeof useGetCommitsQuery>;
export type GetCommitsLazyQueryHookResult = ReturnType<
  typeof useGetCommitsLazyQuery
>;
export type GetCommitsQueryResult = Apollo.QueryResult<
  Types.GetCommitsQuery,
  Types.GetCommitsQueryVariables
>;
export const GetDagDocument = gql`
  query getDag($args: DagQueryArgs!) {
    dag(args: $args) {
      name
      state
      access
      parents
      type
      jobState
      createdAt
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
      name
      state
      access
      parents
      type
      jobState
      createdAt
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
export const DatumDocument = gql`
  query datum($args: DatumQueryArgs!) {
    datum(args: $args) {
      ...Datum
    }
  }
  ${DatumFragmentDoc}
`;

/**
 * __useDatumQuery__
 *
 * To run a query within a React component, call `useDatumQuery` and pass it any options that fit your needs.
 * When your component renders, `useDatumQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDatumQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useDatumQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.DatumQuery,
    Types.DatumQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.DatumQuery, Types.DatumQueryVariables>(
    DatumDocument,
    options,
  );
}
export function useDatumLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.DatumQuery,
    Types.DatumQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.DatumQuery, Types.DatumQueryVariables>(
    DatumDocument,
    options,
  );
}
export type DatumQueryHookResult = ReturnType<typeof useDatumQuery>;
export type DatumLazyQueryHookResult = ReturnType<typeof useDatumLazyQuery>;
export type DatumQueryResult = Apollo.QueryResult<
  Types.DatumQuery,
  Types.DatumQueryVariables
>;
export const DatumsDocument = gql`
  query datums($args: DatumsQueryArgs!) {
    datums(args: $args) {
      ...Datum
    }
  }
  ${DatumFragmentDoc}
`;

/**
 * __useDatumsQuery__
 *
 * To run a query within a React component, call `useDatumsQuery` and pass it any options that fit your needs.
 * When your component renders, `useDatumsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDatumsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useDatumsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.DatumsQuery,
    Types.DatumsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.DatumsQuery, Types.DatumsQueryVariables>(
    DatumsDocument,
    options,
  );
}
export function useDatumsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.DatumsQuery,
    Types.DatumsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.DatumsQuery, Types.DatumsQueryVariables>(
    DatumsDocument,
    options,
  );
}
export type DatumsQueryHookResult = ReturnType<typeof useDatumsQuery>;
export type DatumsLazyQueryHookResult = ReturnType<typeof useDatumsLazyQuery>;
export type DatumsQueryResult = Apollo.QueryResult<
  Types.DatumsQuery,
  Types.DatumsQueryVariables
>;
export const GetEnterpriseInfoDocument = gql`
  query getEnterpriseInfo {
    enterpriseInfo {
      state
      expiration
    }
  }
`;

/**
 * __useGetEnterpriseInfoQuery__
 *
 * To run a query within a React component, call `useGetEnterpriseInfoQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetEnterpriseInfoQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetEnterpriseInfoQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetEnterpriseInfoQuery(
  baseOptions?: Apollo.QueryHookOptions<
    Types.GetEnterpriseInfoQuery,
    Types.GetEnterpriseInfoQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.GetEnterpriseInfoQuery,
    Types.GetEnterpriseInfoQueryVariables
  >(GetEnterpriseInfoDocument, options);
}
export function useGetEnterpriseInfoLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetEnterpriseInfoQuery,
    Types.GetEnterpriseInfoQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetEnterpriseInfoQuery,
    Types.GetEnterpriseInfoQueryVariables
  >(GetEnterpriseInfoDocument, options);
}
export type GetEnterpriseInfoQueryHookResult = ReturnType<
  typeof useGetEnterpriseInfoQuery
>;
export type GetEnterpriseInfoLazyQueryHookResult = ReturnType<
  typeof useGetEnterpriseInfoLazyQuery
>;
export type GetEnterpriseInfoQueryResult = Apollo.QueryResult<
  Types.GetEnterpriseInfoQuery,
  Types.GetEnterpriseInfoQueryVariables
>;
export const GetFilesDocument = gql`
  query getFiles($args: FileQueryArgs!) {
    files(args: $args) {
      diff {
        ...DiffFragment
      }
      files {
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
        commitAction
      }
    }
  }
  ${DiffFragmentFragmentDoc}
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
      outputBranch
      outputCommit
      reason
      jsonDetails
      transformString
      transform {
        cmdList
        image
        debug
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
export const JobSetsDocument = gql`
  query jobSets($args: JobSetsQueryArgs!) {
    jobSets(args: $args) {
      ...JobSetFields
    }
  }
  ${JobSetFieldsFragmentDoc}
`;

/**
 * __useJobSetsQuery__
 *
 * To run a query within a React component, call `useJobSetsQuery` and pass it any options that fit your needs.
 * When your component renders, `useJobSetsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useJobSetsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useJobSetsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.JobSetsQuery,
    Types.JobSetsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.JobSetsQuery, Types.JobSetsQueryVariables>(
    JobSetsDocument,
    options,
  );
}
export function useJobSetsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.JobSetsQuery,
    Types.JobSetsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.JobSetsQuery, Types.JobSetsQueryVariables>(
    JobSetsDocument,
    options,
  );
}
export type JobSetsQueryHookResult = ReturnType<typeof useJobSetsQuery>;
export type JobSetsLazyQueryHookResult = ReturnType<typeof useJobSetsLazyQuery>;
export type JobSetsQueryResult = Apollo.QueryResult<
  Types.JobSetsQuery,
  Types.JobSetsQueryVariables
>;
export const JobsDocument = gql`
  query jobs($args: JobsQueryArgs!) {
    jobs(args: $args) {
      ...JobOverview
      inputString
      inputBranch
      transformString
      transform {
        cmdList
        image
        debug
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
      datumTimeoutS
      datumTries
      jobTimeoutS
      outputBranch
      s3OutputRepo
      egress
      jsonSpec
      reason
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
export const PipelinesDocument = gql`
  query pipelines($args: PipelinesQueryArgs!) {
    pipelines(args: $args) {
      id
      name
      state
      type
      description
      datumTimeoutS
      datumTries
      jobTimeoutS
      outputBranch
      s3OutputRepo
      egress
      jsonSpec
      reason
    }
  }
`;

/**
 * __usePipelinesQuery__
 *
 * To run a query within a React component, call `usePipelinesQuery` and pass it any options that fit your needs.
 * When your component renders, `usePipelinesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = usePipelinesQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function usePipelinesQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.PipelinesQuery,
    Types.PipelinesQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.PipelinesQuery, Types.PipelinesQueryVariables>(
    PipelinesDocument,
    options,
  );
}
export function usePipelinesLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.PipelinesQuery,
    Types.PipelinesQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.PipelinesQuery,
    Types.PipelinesQueryVariables
  >(PipelinesDocument, options);
}
export type PipelinesQueryHookResult = ReturnType<typeof usePipelinesQuery>;
export type PipelinesLazyQueryHookResult = ReturnType<
  typeof usePipelinesLazyQuery
>;
export type PipelinesQueryResult = Apollo.QueryResult<
  Types.PipelinesQuery,
  Types.PipelinesQueryVariables
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
        name
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
export const ReposDocument = gql`
  query repos($args: ReposQueryArgs!) {
    repos(args: $args) {
      branches {
        name
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
 * __useReposQuery__
 *
 * To run a query within a React component, call `useReposQuery` and pass it any options that fit your needs.
 * When your component renders, `useReposQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useReposQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useReposQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.ReposQuery,
    Types.ReposQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.ReposQuery, Types.ReposQueryVariables>(
    ReposDocument,
    options,
  );
}
export function useReposLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.ReposQuery,
    Types.ReposQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.ReposQuery, Types.ReposQueryVariables>(
    ReposDocument,
    options,
  );
}
export type ReposQueryHookResult = ReturnType<typeof useReposQuery>;
export type ReposLazyQueryHookResult = ReturnType<typeof useReposLazyQuery>;
export type ReposQueryResult = Apollo.QueryResult<
  Types.ReposQuery,
  Types.ReposQueryVariables
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
