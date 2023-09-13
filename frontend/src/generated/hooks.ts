import {gql} from '@apollo/client';
import * as Apollo from '@apollo/client';
import * as Types from '@graphqlTypes';
const defaultOptions = {} as const;
export const BranchFragmentFragmentDoc = gql`
  fragment BranchFragment on Branch {
    name
    repo {
      name
      type
    }
  }
`;
export const CommitFragmentFragmentDoc = gql`
  fragment CommitFragment on Commit {
    repoName
    branch {
      name
      repo {
        name
        type
      }
    }
    description
    originKind
    id
    started
    finished
    sizeBytes
    sizeDisplay
  }
`;
export const DatumFragmentDoc = gql`
  fragment Datum on Datum {
    id
    jobId
    requestedJobId
    state
    downloadTimestamp {
      seconds
      nanos
    }
    uploadTimestamp {
      seconds
      nanos
    }
    processTimestamp {
      seconds
      nanos
    }
    downloadBytes
    uploadBytes
  }
`;
export const DiffFragmentFragmentDoc = gql`
  fragment DiffFragment on Diff {
    size
    sizeDisplay
    filesUpdated {
      count
      sizeDelta
    }
    filesAdded {
      count
      sizeDelta
    }
    filesDeleted {
      count
      sizeDelta
    }
  }
`;
export const JobOverviewFragmentDoc = gql`
  fragment JobOverview on Job {
    id
    state
    nodeState
    createdAt
    startedAt
    finishedAt
    restarts
    pipelineName
    pipelineVersion
    reason
    dataProcessed
    dataSkipped
    dataFailed
    dataTotal
    dataRecovered
    downloadBytesDisplay
    uploadBytesDisplay
    outputCommit
  }
`;
export const JobSetFieldsFragmentDoc = gql`
  fragment JobSetFields on JobSet {
    id
    state
    createdAt
    startedAt
    finishedAt
    inProgress
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
export const RepoFragmentFragmentDoc = gql`
  fragment RepoFragment on Repo {
    branches {
      name
    }
    createdAt
    description
    id
    name
    sizeDisplay
    sizeBytes
    access
    projectId
    authInfo {
      rolesList
    }
  }
`;
export const RepoWithLinkedPipelineFragmentFragmentDoc = gql`
  fragment RepoWithLinkedPipelineFragment on Repo {
    branches {
      name
    }
    createdAt
    description
    id
    name
    sizeDisplay
    sizeBytes
    access
    projectId
    linkedPipeline {
      id
      name
    }
    authInfo {
      rolesList
    }
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
export const CreatePipelineV2Document = gql`
  mutation createPipelineV2($args: CreatePipelineV2Args!) {
    createPipelineV2(args: $args) {
      effectiveCreatePipelineRequestJson
    }
  }
`;
export type CreatePipelineV2MutationFn = Apollo.MutationFunction<
  Types.CreatePipelineV2Mutation,
  Types.CreatePipelineV2MutationVariables
>;

/**
 * __useCreatePipelineV2Mutation__
 *
 * To run a mutation, you first call `useCreatePipelineV2Mutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreatePipelineV2Mutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createPipelineV2Mutation, { data, loading, error }] = useCreatePipelineV2Mutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useCreatePipelineV2Mutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.CreatePipelineV2Mutation,
    Types.CreatePipelineV2MutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.CreatePipelineV2Mutation,
    Types.CreatePipelineV2MutationVariables
  >(CreatePipelineV2Document, options);
}
export type CreatePipelineV2MutationHookResult = ReturnType<
  typeof useCreatePipelineV2Mutation
>;
export type CreatePipelineV2MutationResult =
  Apollo.MutationResult<Types.CreatePipelineV2Mutation>;
export type CreatePipelineV2MutationOptions = Apollo.BaseMutationOptions<
  Types.CreatePipelineV2Mutation,
  Types.CreatePipelineV2MutationVariables
>;
export const CreateProjectDocument = gql`
  mutation createProject($args: CreateProjectArgs!) {
    createProject(args: $args) {
      id
      description
      status
      createdAt {
        seconds
        nanos
      }
    }
  }
`;
export type CreateProjectMutationFn = Apollo.MutationFunction<
  Types.CreateProjectMutation,
  Types.CreateProjectMutationVariables
>;

/**
 * __useCreateProjectMutation__
 *
 * To run a mutation, you first call `useCreateProjectMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateProjectMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createProjectMutation, { data, loading, error }] = useCreateProjectMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useCreateProjectMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.CreateProjectMutation,
    Types.CreateProjectMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.CreateProjectMutation,
    Types.CreateProjectMutationVariables
  >(CreateProjectDocument, options);
}
export type CreateProjectMutationHookResult = ReturnType<
  typeof useCreateProjectMutation
>;
export type CreateProjectMutationResult =
  Apollo.MutationResult<Types.CreateProjectMutation>;
export type CreateProjectMutationOptions = Apollo.BaseMutationOptions<
  Types.CreateProjectMutation,
  Types.CreateProjectMutationVariables
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
export const DeleteFilesDocument = gql`
  mutation deleteFiles($args: DeleteFilesArgs!) {
    deleteFiles(args: $args)
  }
`;
export type DeleteFilesMutationFn = Apollo.MutationFunction<
  Types.DeleteFilesMutation,
  Types.DeleteFilesMutationVariables
>;

/**
 * __useDeleteFilesMutation__
 *
 * To run a mutation, you first call `useDeleteFilesMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteFilesMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteFilesMutation, { data, loading, error }] = useDeleteFilesMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useDeleteFilesMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.DeleteFilesMutation,
    Types.DeleteFilesMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.DeleteFilesMutation,
    Types.DeleteFilesMutationVariables
  >(DeleteFilesDocument, options);
}
export type DeleteFilesMutationHookResult = ReturnType<
  typeof useDeleteFilesMutation
>;
export type DeleteFilesMutationResult =
  Apollo.MutationResult<Types.DeleteFilesMutation>;
export type DeleteFilesMutationOptions = Apollo.BaseMutationOptions<
  Types.DeleteFilesMutation,
  Types.DeleteFilesMutationVariables
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
export const DeleteProjectAndResourcesDocument = gql`
  mutation deleteProjectAndResources($args: DeleteProjectAndResourcesArgs!) {
    deleteProjectAndResources(args: $args)
  }
`;
export type DeleteProjectAndResourcesMutationFn = Apollo.MutationFunction<
  Types.DeleteProjectAndResourcesMutation,
  Types.DeleteProjectAndResourcesMutationVariables
>;

/**
 * __useDeleteProjectAndResourcesMutation__
 *
 * To run a mutation, you first call `useDeleteProjectAndResourcesMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteProjectAndResourcesMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteProjectAndResourcesMutation, { data, loading, error }] = useDeleteProjectAndResourcesMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useDeleteProjectAndResourcesMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.DeleteProjectAndResourcesMutation,
    Types.DeleteProjectAndResourcesMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.DeleteProjectAndResourcesMutation,
    Types.DeleteProjectAndResourcesMutationVariables
  >(DeleteProjectAndResourcesDocument, options);
}
export type DeleteProjectAndResourcesMutationHookResult = ReturnType<
  typeof useDeleteProjectAndResourcesMutation
>;
export type DeleteProjectAndResourcesMutationResult =
  Apollo.MutationResult<Types.DeleteProjectAndResourcesMutation>;
export type DeleteProjectAndResourcesMutationOptions =
  Apollo.BaseMutationOptions<
    Types.DeleteProjectAndResourcesMutation,
    Types.DeleteProjectAndResourcesMutationVariables
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
export const ModifyRolesDocument = gql`
  mutation modifyRoles($args: ModifyRolesArgs!) {
    modifyRoles(args: $args)
  }
`;
export type ModifyRolesMutationFn = Apollo.MutationFunction<
  Types.ModifyRolesMutation,
  Types.ModifyRolesMutationVariables
>;

/**
 * __useModifyRolesMutation__
 *
 * To run a mutation, you first call `useModifyRolesMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useModifyRolesMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [modifyRolesMutation, { data, loading, error }] = useModifyRolesMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useModifyRolesMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.ModifyRolesMutation,
    Types.ModifyRolesMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.ModifyRolesMutation,
    Types.ModifyRolesMutationVariables
  >(ModifyRolesDocument, options);
}
export type ModifyRolesMutationHookResult = ReturnType<
  typeof useModifyRolesMutation
>;
export type ModifyRolesMutationResult =
  Apollo.MutationResult<Types.ModifyRolesMutation>;
export type ModifyRolesMutationOptions = Apollo.BaseMutationOptions<
  Types.ModifyRolesMutation,
  Types.ModifyRolesMutationVariables
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
export const MutationDocument = gql`
  mutation Mutation($args: SetClusterDefaultsArgs!) {
    setClusterDefaults(args: $args) {
      affectedPipelinesList {
        name
        project {
          name
        }
      }
    }
  }
`;
export type MutationMutationFn = Apollo.MutationFunction<
  Types.MutationMutation,
  Types.MutationMutationVariables
>;

/**
 * __useMutationMutation__
 *
 * To run a mutation, you first call `useMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [mutationMutation, { data, loading, error }] = useMutationMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useMutationMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.MutationMutation,
    Types.MutationMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.MutationMutation,
    Types.MutationMutationVariables
  >(MutationDocument, options);
}
export type MutationMutationHookResult = ReturnType<typeof useMutationMutation>;
export type MutationMutationResult =
  Apollo.MutationResult<Types.MutationMutation>;
export type MutationMutationOptions = Apollo.BaseMutationOptions<
  Types.MutationMutation,
  Types.MutationMutationVariables
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
export const UpdateProjectDocument = gql`
  mutation updateProject($args: UpdateProjectArgs!) {
    updateProject(args: $args) {
      id
      description
    }
  }
`;
export type UpdateProjectMutationFn = Apollo.MutationFunction<
  Types.UpdateProjectMutation,
  Types.UpdateProjectMutationVariables
>;

/**
 * __useUpdateProjectMutation__
 *
 * To run a mutation, you first call `useUpdateProjectMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateProjectMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateProjectMutation, { data, loading, error }] = useUpdateProjectMutation({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useUpdateProjectMutation(
  baseOptions?: Apollo.MutationHookOptions<
    Types.UpdateProjectMutation,
    Types.UpdateProjectMutationVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useMutation<
    Types.UpdateProjectMutation,
    Types.UpdateProjectMutationVariables
  >(UpdateProjectDocument, options);
}
export type UpdateProjectMutationHookResult = ReturnType<
  typeof useUpdateProjectMutation
>;
export type UpdateProjectMutationResult =
  Apollo.MutationResult<Types.UpdateProjectMutation>;
export type UpdateProjectMutationOptions = Apollo.BaseMutationOptions<
  Types.UpdateProjectMutation,
  Types.UpdateProjectMutationVariables
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
export const GetAdminInfoDocument = gql`
  query getAdminInfo {
    adminInfo {
      clusterId
    }
  }
`;

/**
 * __useGetAdminInfoQuery__
 *
 * To run a query within a React component, call `useGetAdminInfoQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetAdminInfoQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetAdminInfoQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetAdminInfoQuery(
  baseOptions?: Apollo.QueryHookOptions<
    Types.GetAdminInfoQuery,
    Types.GetAdminInfoQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.GetAdminInfoQuery,
    Types.GetAdminInfoQueryVariables
  >(GetAdminInfoDocument, options);
}
export function useGetAdminInfoLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetAdminInfoQuery,
    Types.GetAdminInfoQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetAdminInfoQuery,
    Types.GetAdminInfoQueryVariables
  >(GetAdminInfoDocument, options);
}
export type GetAdminInfoQueryHookResult = ReturnType<
  typeof useGetAdminInfoQuery
>;
export type GetAdminInfoLazyQueryHookResult = ReturnType<
  typeof useGetAdminInfoLazyQuery
>;
export type GetAdminInfoQueryResult = Apollo.QueryResult<
  Types.GetAdminInfoQuery,
  Types.GetAdminInfoQueryVariables
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
export const GetAuthorizeDocument = gql`
  query getAuthorize($args: GetAuthorizeArgs!) {
    getAuthorize(args: $args) {
      satisfiedList
      missingList
      authorized
      principal
    }
  }
`;

/**
 * __useGetAuthorizeQuery__
 *
 * To run a query within a React component, call `useGetAuthorizeQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetAuthorizeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetAuthorizeQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetAuthorizeQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetAuthorizeQuery,
    Types.GetAuthorizeQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.GetAuthorizeQuery,
    Types.GetAuthorizeQueryVariables
  >(GetAuthorizeDocument, options);
}
export function useGetAuthorizeLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetAuthorizeQuery,
    Types.GetAuthorizeQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetAuthorizeQuery,
    Types.GetAuthorizeQueryVariables
  >(GetAuthorizeDocument, options);
}
export type GetAuthorizeQueryHookResult = ReturnType<
  typeof useGetAuthorizeQuery
>;
export type GetAuthorizeLazyQueryHookResult = ReturnType<
  typeof useGetAuthorizeLazyQuery
>;
export type GetAuthorizeQueryResult = Apollo.QueryResult<
  Types.GetAuthorizeQuery,
  Types.GetAuthorizeQueryVariables
>;
export const GetBranchesDocument = gql`
  query getBranches($args: BranchesQueryArgs!) {
    branches(args: $args) {
      ...BranchFragment
    }
  }
  ${BranchFragmentFragmentDoc}
`;

/**
 * __useGetBranchesQuery__
 *
 * To run a query within a React component, call `useGetBranchesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetBranchesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetBranchesQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetBranchesQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetBranchesQuery,
    Types.GetBranchesQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.GetBranchesQuery,
    Types.GetBranchesQueryVariables
  >(GetBranchesDocument, options);
}
export function useGetBranchesLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetBranchesQuery,
    Types.GetBranchesQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetBranchesQuery,
    Types.GetBranchesQueryVariables
  >(GetBranchesDocument, options);
}
export type GetBranchesQueryHookResult = ReturnType<typeof useGetBranchesQuery>;
export type GetBranchesLazyQueryHookResult = ReturnType<
  typeof useGetBranchesLazyQuery
>;
export type GetBranchesQueryResult = Apollo.QueryResult<
  Types.GetBranchesQuery,
  Types.GetBranchesQueryVariables
>;
export const GetClusterDefaultsDocument = gql`
  query GetClusterDefaults {
    getClusterDefaults {
      clusterDefaultsJson
    }
  }
`;

/**
 * __useGetClusterDefaultsQuery__
 *
 * To run a query within a React component, call `useGetClusterDefaultsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetClusterDefaultsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetClusterDefaultsQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetClusterDefaultsQuery(
  baseOptions?: Apollo.QueryHookOptions<
    Types.GetClusterDefaultsQuery,
    Types.GetClusterDefaultsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.GetClusterDefaultsQuery,
    Types.GetClusterDefaultsQueryVariables
  >(GetClusterDefaultsDocument, options);
}
export function useGetClusterDefaultsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetClusterDefaultsQuery,
    Types.GetClusterDefaultsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetClusterDefaultsQuery,
    Types.GetClusterDefaultsQueryVariables
  >(GetClusterDefaultsDocument, options);
}
export type GetClusterDefaultsQueryHookResult = ReturnType<
  typeof useGetClusterDefaultsQuery
>;
export type GetClusterDefaultsLazyQueryHookResult = ReturnType<
  typeof useGetClusterDefaultsLazyQuery
>;
export type GetClusterDefaultsQueryResult = Apollo.QueryResult<
  Types.GetClusterDefaultsQuery,
  Types.GetClusterDefaultsQueryVariables
>;
export const CommitDiffDocument = gql`
  query commitDiff($args: CommitDiffQueryArgs!) {
    commitDiff(args: $args) {
      ...DiffFragment
    }
  }
  ${DiffFragmentFragmentDoc}
`;

/**
 * __useCommitDiffQuery__
 *
 * To run a query within a React component, call `useCommitDiffQuery` and pass it any options that fit your needs.
 * When your component renders, `useCommitDiffQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCommitDiffQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useCommitDiffQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.CommitDiffQuery,
    Types.CommitDiffQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.CommitDiffQuery, Types.CommitDiffQueryVariables>(
    CommitDiffDocument,
    options,
  );
}
export function useCommitDiffLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.CommitDiffQuery,
    Types.CommitDiffQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.CommitDiffQuery,
    Types.CommitDiffQueryVariables
  >(CommitDiffDocument, options);
}
export type CommitDiffQueryHookResult = ReturnType<typeof useCommitDiffQuery>;
export type CommitDiffLazyQueryHookResult = ReturnType<
  typeof useCommitDiffLazyQuery
>;
export type CommitDiffQueryResult = Apollo.QueryResult<
  Types.CommitDiffQuery,
  Types.CommitDiffQueryVariables
>;
export const CommitDocument = gql`
  query commit($args: CommitQueryArgs!) {
    commit(args: $args) {
      ...CommitFragment
    }
  }
  ${CommitFragmentFragmentDoc}
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
export const CommitSearchDocument = gql`
  query commitSearch($args: CommitSearchQueryArgs!) {
    commitSearch(args: $args) {
      ...CommitFragment
    }
  }
  ${CommitFragmentFragmentDoc}
`;

/**
 * __useCommitSearchQuery__
 *
 * To run a query within a React component, call `useCommitSearchQuery` and pass it any options that fit your needs.
 * When your component renders, `useCommitSearchQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCommitSearchQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useCommitSearchQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.CommitSearchQuery,
    Types.CommitSearchQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.CommitSearchQuery,
    Types.CommitSearchQueryVariables
  >(CommitSearchDocument, options);
}
export function useCommitSearchLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.CommitSearchQuery,
    Types.CommitSearchQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.CommitSearchQuery,
    Types.CommitSearchQueryVariables
  >(CommitSearchDocument, options);
}
export type CommitSearchQueryHookResult = ReturnType<
  typeof useCommitSearchQuery
>;
export type CommitSearchLazyQueryHookResult = ReturnType<
  typeof useCommitSearchLazyQuery
>;
export type CommitSearchQueryResult = Apollo.QueryResult<
  Types.CommitSearchQuery,
  Types.CommitSearchQueryVariables
>;
export const GetCommitsDocument = gql`
  query getCommits($args: CommitsQueryArgs!) {
    commits(args: $args) {
      items {
        ...CommitFragment
      }
      cursor {
        seconds
        nanos
      }
      parentCommit
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
export const DatumSearchDocument = gql`
  query datumSearch($args: DatumQueryArgs!) {
    datumSearch(args: $args) {
      ...Datum
    }
  }
  ${DatumFragmentDoc}
`;

/**
 * __useDatumSearchQuery__
 *
 * To run a query within a React component, call `useDatumSearchQuery` and pass it any options that fit your needs.
 * When your component renders, `useDatumSearchQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDatumSearchQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useDatumSearchQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.DatumSearchQuery,
    Types.DatumSearchQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.DatumSearchQuery,
    Types.DatumSearchQueryVariables
  >(DatumSearchDocument, options);
}
export function useDatumSearchLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.DatumSearchQuery,
    Types.DatumSearchQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.DatumSearchQuery,
    Types.DatumSearchQueryVariables
  >(DatumSearchDocument, options);
}
export type DatumSearchQueryHookResult = ReturnType<typeof useDatumSearchQuery>;
export type DatumSearchLazyQueryHookResult = ReturnType<
  typeof useDatumSearchLazyQuery
>;
export type DatumSearchQueryResult = Apollo.QueryResult<
  Types.DatumSearchQuery,
  Types.DatumSearchQueryVariables
>;
export const DatumsDocument = gql`
  query datums($args: DatumsQueryArgs!) {
    datums(args: $args) {
      items {
        ...Datum
      }
      cursor
      hasNextPage
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
export const FileDownloadDocument = gql`
  query fileDownload($args: FileDownloadArgs!) {
    fileDownload(args: $args)
  }
`;

/**
 * __useFileDownloadQuery__
 *
 * To run a query within a React component, call `useFileDownloadQuery` and pass it any options that fit your needs.
 * When your component renders, `useFileDownloadQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFileDownloadQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useFileDownloadQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.FileDownloadQuery,
    Types.FileDownloadQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.FileDownloadQuery,
    Types.FileDownloadQueryVariables
  >(FileDownloadDocument, options);
}
export function useFileDownloadLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.FileDownloadQuery,
    Types.FileDownloadQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.FileDownloadQuery,
    Types.FileDownloadQueryVariables
  >(FileDownloadDocument, options);
}
export type FileDownloadQueryHookResult = ReturnType<
  typeof useFileDownloadQuery
>;
export type FileDownloadLazyQueryHookResult = ReturnType<
  typeof useFileDownloadLazyQuery
>;
export type FileDownloadQueryResult = Apollo.QueryResult<
  Types.FileDownloadQuery,
  Types.FileDownloadQueryVariables
>;
export const GetFilesDocument = gql`
  query getFiles($args: FileQueryArgs!) {
    files(args: $args) {
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
        commitAction
      }
      cursor
      hasNextPage
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
export const FindCommitsDocument = gql`
  query findCommits($args: FindCommitsQueryArgs!) {
    findCommits(args: $args) {
      commits {
        id
        started
        description
        commitAction
      }
      cursor
      hasNextPage
    }
  }
`;

/**
 * __useFindCommitsQuery__
 *
 * To run a query within a React component, call `useFindCommitsQuery` and pass it any options that fit your needs.
 * When your component renders, `useFindCommitsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFindCommitsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useFindCommitsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.FindCommitsQuery,
    Types.FindCommitsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.FindCommitsQuery,
    Types.FindCommitsQueryVariables
  >(FindCommitsDocument, options);
}
export function useFindCommitsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.FindCommitsQuery,
    Types.FindCommitsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.FindCommitsQuery,
    Types.FindCommitsQueryVariables
  >(FindCommitsDocument, options);
}
export type FindCommitsQueryHookResult = ReturnType<typeof useFindCommitsQuery>;
export type FindCommitsLazyQueryHookResult = ReturnType<
  typeof useFindCommitsLazyQuery
>;
export type FindCommitsQueryResult = Apollo.QueryResult<
  Types.FindCommitsQuery,
  Types.FindCommitsQueryVariables
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
      items {
        ...JobSetFields
      }
      cursor {
        seconds
        nanos
      }
      hasNextPage
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
export const JobsByPipelineDocument = gql`
  query jobsByPipeline($args: JobsByPipelineQueryArgs!) {
    jobsByPipeline(args: $args) {
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
 * __useJobsByPipelineQuery__
 *
 * To run a query within a React component, call `useJobsByPipelineQuery` and pass it any options that fit your needs.
 * When your component renders, `useJobsByPipelineQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useJobsByPipelineQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useJobsByPipelineQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.JobsByPipelineQuery,
    Types.JobsByPipelineQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.JobsByPipelineQuery,
    Types.JobsByPipelineQueryVariables
  >(JobsByPipelineDocument, options);
}
export function useJobsByPipelineLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.JobsByPipelineQuery,
    Types.JobsByPipelineQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.JobsByPipelineQuery,
    Types.JobsByPipelineQueryVariables
  >(JobsByPipelineDocument, options);
}
export type JobsByPipelineQueryHookResult = ReturnType<
  typeof useJobsByPipelineQuery
>;
export type JobsByPipelineLazyQueryHookResult = ReturnType<
  typeof useJobsByPipelineLazyQuery
>;
export type JobsByPipelineQueryResult = Apollo.QueryResult<
  Types.JobsByPipelineQuery,
  Types.JobsByPipelineQueryVariables
>;
export const JobsDocument = gql`
  query jobs($args: JobsQueryArgs!) {
    jobs(args: $args) {
      items {
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
      cursor {
        seconds
        nanos
      }
      hasNextPage
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
      items {
        ...LogFields
      }
      cursor {
        timestamp {
          seconds
          nanos
        }
        message
      }
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
export const GetPermissionsDocument = gql`
  query getPermissions($args: GetPermissionsArgs!) {
    getPermissions(args: $args) {
      rolesList
    }
  }
`;

/**
 * __useGetPermissionsQuery__
 *
 * To run a query within a React component, call `useGetPermissionsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetPermissionsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetPermissionsQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetPermissionsQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetPermissionsQuery,
    Types.GetPermissionsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.GetPermissionsQuery,
    Types.GetPermissionsQueryVariables
  >(GetPermissionsDocument, options);
}
export function useGetPermissionsLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetPermissionsQuery,
    Types.GetPermissionsQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetPermissionsQuery,
    Types.GetPermissionsQueryVariables
  >(GetPermissionsDocument, options);
}
export type GetPermissionsQueryHookResult = ReturnType<
  typeof useGetPermissionsQuery
>;
export type GetPermissionsLazyQueryHookResult = ReturnType<
  typeof useGetPermissionsLazyQuery
>;
export type GetPermissionsQueryResult = Apollo.QueryResult<
  Types.GetPermissionsQuery,
  Types.GetPermissionsQueryVariables
>;
export const PipelineDocument = gql`
  query pipeline($args: PipelineQueryArgs!) {
    pipeline(args: $args) {
      id
      name
      description
      version
      createdAt
      state
      nodeState
      stopped
      recentError
      lastJobState
      lastJobNodeState
      type
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
      description
      version
      createdAt
      state
      nodeState
      stopped
      recentError
      lastJobState
      lastJobNodeState
      type
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
      description
      createdAt {
        seconds
        nanos
      }
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
export const ProjectStatusDocument = gql`
  query projectStatus($id: ID!) {
    projectStatus(id: $id) {
      id
      status
    }
  }
`;

/**
 * __useProjectStatusQuery__
 *
 * To run a query within a React component, call `useProjectStatusQuery` and pass it any options that fit your needs.
 * When your component renders, `useProjectStatusQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProjectStatusQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useProjectStatusQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.ProjectStatusQuery,
    Types.ProjectStatusQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.ProjectStatusQuery,
    Types.ProjectStatusQueryVariables
  >(ProjectStatusDocument, options);
}
export function useProjectStatusLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.ProjectStatusQuery,
    Types.ProjectStatusQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.ProjectStatusQuery,
    Types.ProjectStatusQueryVariables
  >(ProjectStatusDocument, options);
}
export type ProjectStatusQueryHookResult = ReturnType<
  typeof useProjectStatusQuery
>;
export type ProjectStatusLazyQueryHookResult = ReturnType<
  typeof useProjectStatusLazyQuery
>;
export type ProjectStatusQueryResult = Apollo.QueryResult<
  Types.ProjectStatusQuery,
  Types.ProjectStatusQueryVariables
>;
export const ProjectsDocument = gql`
  query projects {
    projects {
      id
      description
      status
      createdAt {
        seconds
        nanos
      }
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
      ...RepoFragment
    }
  }
  ${RepoFragmentFragmentDoc}
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
export const RepoWithCommitDocument = gql`
  query repoWithCommit($args: RepoQueryArgs!) {
    repo(args: $args) {
      ...RepoFragment
      lastCommit {
        ...CommitFragment
      }
    }
  }
  ${RepoFragmentFragmentDoc}
  ${CommitFragmentFragmentDoc}
`;

/**
 * __useRepoWithCommitQuery__
 *
 * To run a query within a React component, call `useRepoWithCommitQuery` and pass it any options that fit your needs.
 * When your component renders, `useRepoWithCommitQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useRepoWithCommitQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useRepoWithCommitQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.RepoWithCommitQuery,
    Types.RepoWithCommitQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.RepoWithCommitQuery,
    Types.RepoWithCommitQueryVariables
  >(RepoWithCommitDocument, options);
}
export function useRepoWithCommitLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.RepoWithCommitQuery,
    Types.RepoWithCommitQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.RepoWithCommitQuery,
    Types.RepoWithCommitQueryVariables
  >(RepoWithCommitDocument, options);
}
export type RepoWithCommitQueryHookResult = ReturnType<
  typeof useRepoWithCommitQuery
>;
export type RepoWithCommitLazyQueryHookResult = ReturnType<
  typeof useRepoWithCommitLazyQuery
>;
export type RepoWithCommitQueryResult = Apollo.QueryResult<
  Types.RepoWithCommitQuery,
  Types.RepoWithCommitQueryVariables
>;
export const RepoWithLinkedPipelineDocument = gql`
  query repoWithLinkedPipeline($args: RepoQueryArgs!) {
    repo(args: $args) {
      ...RepoWithLinkedPipelineFragment
    }
  }
  ${RepoWithLinkedPipelineFragmentFragmentDoc}
`;

/**
 * __useRepoWithLinkedPipelineQuery__
 *
 * To run a query within a React component, call `useRepoWithLinkedPipelineQuery` and pass it any options that fit your needs.
 * When your component renders, `useRepoWithLinkedPipelineQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useRepoWithLinkedPipelineQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useRepoWithLinkedPipelineQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.RepoWithLinkedPipelineQuery,
    Types.RepoWithLinkedPipelineQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.RepoWithLinkedPipelineQuery,
    Types.RepoWithLinkedPipelineQueryVariables
  >(RepoWithLinkedPipelineDocument, options);
}
export function useRepoWithLinkedPipelineLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.RepoWithLinkedPipelineQuery,
    Types.RepoWithLinkedPipelineQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.RepoWithLinkedPipelineQuery,
    Types.RepoWithLinkedPipelineQueryVariables
  >(RepoWithLinkedPipelineDocument, options);
}
export type RepoWithLinkedPipelineQueryHookResult = ReturnType<
  typeof useRepoWithLinkedPipelineQuery
>;
export type RepoWithLinkedPipelineLazyQueryHookResult = ReturnType<
  typeof useRepoWithLinkedPipelineLazyQuery
>;
export type RepoWithLinkedPipelineQueryResult = Apollo.QueryResult<
  Types.RepoWithLinkedPipelineQuery,
  Types.RepoWithLinkedPipelineQueryVariables
>;
export const ReposDocument = gql`
  query repos($args: ReposQueryArgs!) {
    repos(args: $args) {
      ...RepoFragment
    }
  }
  ${RepoFragmentFragmentDoc}
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
export const ReposWithCommitDocument = gql`
  query reposWithCommit($args: ReposQueryArgs!) {
    repos(args: $args) {
      ...RepoFragment
      lastCommit {
        ...CommitFragment
      }
    }
  }
  ${RepoFragmentFragmentDoc}
  ${CommitFragmentFragmentDoc}
`;

/**
 * __useReposWithCommitQuery__
 *
 * To run a query within a React component, call `useReposWithCommitQuery` and pass it any options that fit your needs.
 * When your component renders, `useReposWithCommitQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useReposWithCommitQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useReposWithCommitQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.ReposWithCommitQuery,
    Types.ReposWithCommitQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.ReposWithCommitQuery,
    Types.ReposWithCommitQueryVariables
  >(ReposWithCommitDocument, options);
}
export function useReposWithCommitLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.ReposWithCommitQuery,
    Types.ReposWithCommitQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.ReposWithCommitQuery,
    Types.ReposWithCommitQueryVariables
  >(ReposWithCommitDocument, options);
}
export type ReposWithCommitQueryHookResult = ReturnType<
  typeof useReposWithCommitQuery
>;
export type ReposWithCommitLazyQueryHookResult = ReturnType<
  typeof useReposWithCommitLazyQuery
>;
export type ReposWithCommitQueryResult = Apollo.QueryResult<
  Types.ReposWithCommitQuery,
  Types.ReposWithCommitQueryVariables
>;
export const GetRolesDocument = gql`
  query getRoles($args: GetRolesArgs!) {
    getRoles(args: $args) {
      roleBindings {
        principal
        roles
      }
    }
  }
`;

/**
 * __useGetRolesQuery__
 *
 * To run a query within a React component, call `useGetRolesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetRolesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetRolesQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetRolesQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetRolesQuery,
    Types.GetRolesQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<Types.GetRolesQuery, Types.GetRolesQueryVariables>(
    GetRolesDocument,
    options,
  );
}
export function useGetRolesLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetRolesQuery,
    Types.GetRolesQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<Types.GetRolesQuery, Types.GetRolesQueryVariables>(
    GetRolesDocument,
    options,
  );
}
export type GetRolesQueryHookResult = ReturnType<typeof useGetRolesQuery>;
export type GetRolesLazyQueryHookResult = ReturnType<
  typeof useGetRolesLazyQuery
>;
export type GetRolesQueryResult = Apollo.QueryResult<
  Types.GetRolesQuery,
  Types.GetRolesQueryVariables
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
export const GetVersionInfoDocument = gql`
  query getVersionInfo {
    versionInfo {
      pachdVersion {
        major
        minor
        micro
        additional
        gitCommit
        gitTreeModified
        buildDate
        goVersion
        platform
      }
      consoleVersion
    }
  }
`;

/**
 * __useGetVersionInfoQuery__
 *
 * To run a query within a React component, call `useGetVersionInfoQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetVersionInfoQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetVersionInfoQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetVersionInfoQuery(
  baseOptions?: Apollo.QueryHookOptions<
    Types.GetVersionInfoQuery,
    Types.GetVersionInfoQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.GetVersionInfoQuery,
    Types.GetVersionInfoQueryVariables
  >(GetVersionInfoDocument, options);
}
export function useGetVersionInfoLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetVersionInfoQuery,
    Types.GetVersionInfoQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetVersionInfoQuery,
    Types.GetVersionInfoQueryVariables
  >(GetVersionInfoDocument, options);
}
export type GetVersionInfoQueryHookResult = ReturnType<
  typeof useGetVersionInfoQuery
>;
export type GetVersionInfoLazyQueryHookResult = ReturnType<
  typeof useGetVersionInfoLazyQuery
>;
export type GetVersionInfoQueryResult = Apollo.QueryResult<
  Types.GetVersionInfoQuery,
  Types.GetVersionInfoQueryVariables
>;
export const GetVerticesDocument = gql`
  query getVertices($args: VerticesQueryArgs!) {
    vertices(args: $args) {
      id
      project
      name
      parents {
        id
        project
        name
      }
      state
      nodeState
      access
      type
      jobState
      jobNodeState
      createdAt
    }
  }
`;

/**
 * __useGetVerticesQuery__
 *
 * To run a query within a React component, call `useGetVerticesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetVerticesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetVerticesQuery({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useGetVerticesQuery(
  baseOptions: Apollo.QueryHookOptions<
    Types.GetVerticesQuery,
    Types.GetVerticesQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useQuery<
    Types.GetVerticesQuery,
    Types.GetVerticesQueryVariables
  >(GetVerticesDocument, options);
}
export function useGetVerticesLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    Types.GetVerticesQuery,
    Types.GetVerticesQueryVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useLazyQuery<
    Types.GetVerticesQuery,
    Types.GetVerticesQueryVariables
  >(GetVerticesDocument, options);
}
export type GetVerticesQueryHookResult = ReturnType<typeof useGetVerticesQuery>;
export type GetVerticesLazyQueryHookResult = ReturnType<
  typeof useGetVerticesLazyQuery
>;
export type GetVerticesQueryResult = Apollo.QueryResult<
  Types.GetVerticesQuery,
  Types.GetVerticesQueryVariables
>;
export const VerticesDocument = gql`
  subscription vertices($args: VerticesQueryArgs!) {
    vertices(args: $args) {
      id
      project
      name
      parents {
        id
        project
        name
      }
      state
      nodeState
      access
      type
      jobState
      jobNodeState
      createdAt
    }
  }
`;

/**
 * __useVerticesSubscription__
 *
 * To run a query within a React component, call `useVerticesSubscription` and pass it any options that fit your needs.
 * When your component renders, `useVerticesSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useVerticesSubscription({
 *   variables: {
 *      args: // value for 'args'
 *   },
 * });
 */
export function useVerticesSubscription(
  baseOptions: Apollo.SubscriptionHookOptions<
    Types.VerticesSubscription,
    Types.VerticesSubscriptionVariables
  >,
) {
  const options = {...defaultOptions, ...baseOptions};
  return Apollo.useSubscription<
    Types.VerticesSubscription,
    Types.VerticesSubscriptionVariables
  >(VerticesDocument, options);
}
export type VerticesSubscriptionHookResult = ReturnType<
  typeof useVerticesSubscription
>;
export type VerticesSubscriptionResult =
  Apollo.SubscriptionResult<Types.VerticesSubscription>;
