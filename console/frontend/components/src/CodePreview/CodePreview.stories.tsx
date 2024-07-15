import React from 'react';

import CodePreview from './CodePreview';

const openCvJson = `
{
  "pipeline": {
    "name": "edges"
  },
  "description": "A pipeline that performs image edge detection by using the OpenCV library.",
  "transform": {
    "cmd": [ "python3", "/edges.py" ],
    "image": "pachyderm/opencv"
  },
  "input": {
    "pfs": {
      "repo": "images",
      "glob": "/*"
    }
  }
}`;

const localDevYaml = `
# SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
# SPDX-License-Identifier: Apache-2.0
deployTarget: custom

pachd:
  image:
    tag: local
  resources:
    requests:
      cpu: 250m
      memory: 512M
  service:
    type: NodePort
  metrics:
    enabled: false
  clusterDeploymentID: dev
  # to enable enterprise features pass in pachd.activateEnterprise=true, and a valid pachd.enterpriseLicenseKey
  activateEnterprise: false
  # enterpriseLicenseKey: ""
  storage:
    backend: MINIO
    minio:
      bucket: "pachyderm-test"
      endpoint: "minio.default.svc.cluster.local:9000"
      id: "minioadmin"
      secret: "minioadmin"
      secure: "false"
      signature: ""

etcd:
  service:
    type: NodePort

postgresql:
  service:
    type: NodePort`;

export default {
  title: 'CodePreview',
  component: CodePreview,
  argTypes: {
    children: {
      defaultValue: openCvJson,
      control: {type: 'radio'},
      options: [openCvJson, localDevYaml],
    },
  },
};

export const Default = ({children}: {children?: React.ReactNode}) => {
  return <CodePreview>{children}</CodePreview>;
};

export const Yaml = () => {
  return <CodePreview>{localDevYaml}</CodePreview>;
};
