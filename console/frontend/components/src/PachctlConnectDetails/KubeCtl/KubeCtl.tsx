import React, {memo} from 'react';

import Description from '../Description';

interface KubeCtlProps {
  clusterName: string;
  GCPZone: string;
  GCPProjectID: string;
}

const KubeCtl: React.FC<KubeCtlProps> = ({
  clusterName,
  GCPZone,
  GCPProjectID,
}) => {
  const kubeCtlCommand =
    `gcloud container clusters get-credentials ${clusterName}` +
    ` --zone ${GCPZone} --project ${GCPProjectID}`;

  return (
    <Description
      id="kubeCtl"
      copyText={kubeCtlCommand}
      title="To connect via kubectl, enter:"
    >
      {kubeCtlCommand}
    </Description>
  );
};

export default memo(KubeCtl);
