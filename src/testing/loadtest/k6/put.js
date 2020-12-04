import { startCommit, finishCommit, deleteRepo } from "./lib/pfs.js";
import { getClient, grpcOK } from "./lib/grpc.js";

export function setup() {
    return startCommit();
}

export function teardown(data) {
    finishCommit(data);
    deleteRepo(data);
}

export default (data) => {
    const client = getClient();
    const put = client.invoke("pfs.API/PutFile", {
        file: {
            commit: data.commit,
            path: "load-test-" + Math.floor(100000 * Math.random()),
        },
        value: "dGhpcyBpcyBhIHRlc3QgZmlsZQo=",
    });
    grpcOK("put file", put);
};
