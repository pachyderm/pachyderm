{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
  },
  components: {
    "workflows": {
      bucket: "kubeflow-ci_temp",
      name: "some-very-very-very-very-very-long-name-manifests-presubmit-test-74-786a",
      namespace: "kubeflow-test-infra",
      prow_env: "JOB_NAME=manifests-presubmit-test,JOB_TYPE=presubmit,PULL_NUMBER=74,REPO_NAME=manifests,REPO_OWNER=kubeflow,BUILD_NUMBER=786a",
      versionTag: null,
    },
    kfctl_go_test: {
      bucket: "kubeflow-ci_temp",
      name: "somefakename",
      namespace: "kubeflow-test-infra",
      prow_env: "",
      deleteKubeflow: true,
      gkeApiVersion: "v1",
      workflowName: "kfctl-go",
      useBasicAuth: "false",
      useIstio: "true",
    },
  },
}
