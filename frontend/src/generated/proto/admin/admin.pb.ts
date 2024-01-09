/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb";
import * as Pfs_v2Pfs from "../pfs/pfs.pb";
import * as Versionpb_v2Version from "../version/versionpb/version.pb";
export type ClusterInfo = {
    __typename?: "ClusterInfo";
    id?: string;
    deploymentId?: string;
    warningsOk?: boolean;
    warnings?: string[];
    proxyHost?: string;
    proxyTls?: boolean;
    paused?: boolean;
    webResources?: WebResource;
};

export type InspectClusterRequest = {
    __typename?: "InspectClusterRequest";
    clientVersion?: Versionpb_v2Version.Version;
    currentProject?: Pfs_v2Pfs.Project;
};

export type WebResource = {
    __typename?: "WebResource";
    archiveDownloadBaseUrl?: string;
    createPipelineRequestJsonSchemaUrl?: string;
};

export class API {
    static InspectCluster(req: InspectClusterRequest, initReq?: fm.InitReq): Promise<ClusterInfo> {
        return fm.fetchReq<InspectClusterRequest, ClusterInfo>(`/admin_v2.API/InspectCluster`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
}
