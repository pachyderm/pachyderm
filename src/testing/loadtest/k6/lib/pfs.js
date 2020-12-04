import grpc from "k6/net/grpc";
import { Counter } from "k6/metrics";
import { getClient, grpcOK } from "./grpc.js";

export function startCommit() {
    const client = getClient();
    const repo = {
        name: "load-test-repo-" + Math.ceil(100000 * Math.random()),
    };
    const create = client.invoke("pfs.API/CreateRepo", {
        repo: repo,
        description: "created by load tests",
    });
    grpcOK("create repo", create);

    let commit = {
        repo: repo,
    };
    const start = client.invoke("pfs.API/StartCommit", {
        parent: commit,
        description: "test",
        branch: "master",
    });
    grpcOK("start commit", start);
    commit = start.message;
    client.close();
    return { repo: repo, commit: commit };
}

export function finishCommit(data) {
    const client = getClient();
    const finish = client.invoke("pfs.API/FinishCommit", {
        commit: data.commit,
    });
    grpcOK("finish commit", finish);
    client.close();
}

export function deleteRepo(data) {
    const client = getClient();
    const del = client.invoke("pfs.API/DeleteRepo", {
        repo: data.repo,
    });
    grpcOK("delete repo", del);
    client.close();
}
