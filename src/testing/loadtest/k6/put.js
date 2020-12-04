import grpc from "k6/net/grpc";
import { sleep, check } from "k6";

const NUM_COMMITS = 1;
const NUM_FILES_PER_COMMIT = 1;

const client = new grpc.Client();
client.load(["../../..", "../../../client"], "client/pfs/pfs.proto");

function grpcOK(op, response) {
    const msg = op + " ok?";
    const c = {};
    (c[msg] = (r) => r && r.status === grpc.StatusOK), check(response, c);
}

export default () => {
    client.connect("localhost:30650", {
        plaintext: true,
    });
    const repo = {
        name: "load-test-repo-" + Math.ceil(100000 * Math.random()),
    };
    const create = client.invoke("pfs.API/CreateRepo", {
        repo: repo,
        description: "created by load tests",
    });
    grpcOK("create", create);

    let commit = {
        repo: repo,
    };
    for (let i = 0; i < NUM_COMMITS; i++) {
        const start = client.invoke("pfs.API/StartCommit", {
            parent: commit,
            description: "test",
            branch: "foo",
        });
        grpcOK("start commit", start);
        commit = start.message;

        for (let j = 0; j < NUM_FILES_PER_COMMIT; j++) {
            const put = client.invoke("pfs.API/PutFile", {
                file: {
                    commit: commit,
                    path: "load-test-" + i,
                },
                value: "dGhpcyBpcyBhIHRlc3QgZmlsZQo=",
            });
            grpcOK("put file", put);
        }
        const finish = client.invoke("pfs.API/FinishCommit", {
            commit: commit,
        });
        grpcOK("finish commit", finish);
    }

    const del = client.invoke("pfs.API/DeleteRepo", {
        repo: repo,
    });
    grpcOK("delete repo", del);

    client.close();
    sleep(1);
};
