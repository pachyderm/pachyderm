import * as pulumi from "@pulumi/pulumi";
import * as k8s from "@pulumi/kubernetes";

const cfg = new pulumi.Config();
const slug = "pachyderm/ci-cluster/dev";
const stackRef = new pulumi.StackReference(slug);
const currBranch = cfg.require("branch").toLowerCase();
const sha = cfg.require("sha");

const kubeConfig = stackRef.getOutput("kubeconfig");

const clusterProvider = new k8s.Provider("k8sprovider", {
  kubeconfig: kubeConfig,
});

const namespace = new k8s.core.v1.Namespace(
  "console-ns",
  {
    metadata: { name: "console-" + currBranch },
  },
  { provider: clusterProvider }
);

const pachyderm = new k8s.helm.v3.Release(
  "console-release",
  {
    chart: "pachyderm",
    namespace: namespace.metadata.name,
    repositoryOpts: {
      repo: "https://pachyderm.github.io/helmchart",
    },
    values: {
      pachd: {
        activateEnterprise: true, // not needed on the low...
        enterpriseLicenseKey: process.env.PACHYDERM_ENTERPRISE_KEY,
        oauthClientSecret: "i9mRbLujCvi8j3NPKOFPklXai71oqz3y",
        rootToken: "1WgTXSc2MccsxunEzXvSAejsKNyT4Lsy",
        enterpriseSecret: "SBgvzhmVtMxiVbzSIzpWqi3fKCfsup3o",
        annotations: {
          "cluster-autoscaler.kubernetes.io/safe-to-evict": "true",
        },
      },
      deployTarget: "LOCAL",
      console: {
        enabled: true,
        image: {
          tag: sha,
        },
        config: {
          oauthClientSecret: "RW5WuOanil1nGPc6e5h8iNMJplt4tfjX",
        }
      },
      ingress: {
        enabled: true,
        host: "console-" + currBranch + ".clusters-ci.pachyderm.io",
        annotations: {
          "kubernetes.io/ingress.class":              "traefik"
          //"traefik.ingress.kubernetes.io/router.tls": "true",
        },
      },
      global: {
        postgresql: {
          postgresqlPassword: "Vq90lGAZBA",
          postgresqlPostgresPassword: "6pz2ykqaIv",
        },
      },
    },
  },
  { provider: clusterProvider }
);
