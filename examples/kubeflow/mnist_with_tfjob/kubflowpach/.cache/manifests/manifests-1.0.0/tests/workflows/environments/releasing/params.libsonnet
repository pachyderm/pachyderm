local params = import '../../components/params.libsonnet';

params + {
  components+: {
    // Insert component parameter overrides here. Ex:
    // guestbook +: {
    // name: "guestbook-dev",
    // replicas: params.global.replicas,
    // },
    workflows+: {
      bucket: 'kubeflow-releasing-artifacts',
      gcpCredentialsSecretName: 'gcp-credentials',
      name: 'jlewi-kubeflow-manifests-release-403-2f58',
      namespace: 'kubeflow-releasing',
      project: 'kubeflow-releasing',
      prow_env: 'JOB_NAME=kubeflow-manifests-release,JOB_TYPE=presubmit,PULL_NUMBER=403,REPO_NAME=kubeflow-manifests,REPO_OWNER=kubeflow,BUILD_NUMBER=2f58',
      versionTag: 'v20180226-403',
      zone: 'us-central1-a',
    },
  },
}
