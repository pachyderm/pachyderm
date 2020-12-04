import grpc from "k6/net/grpc";
import { check } from "k6";

const GOPATH = __ENV.GOPATH ? __ENV.GOPATH : __ENV.HOME + "/go";

const client = new grpc.Client();
client.load(["../../..", GOPATH + "/src/github.com/gogo/protobuf"], "client/pfs/pfs.proto");

export function getClient() {
    client.connect("localhost:30650", {
        plaintext: true,
    });
    return client;
}

export function grpcOK(op, response) {
    const msg = op + " ok?";
    const c = {};
    c[msg] = (r) => r && r.status === grpc.StatusOK;
    check(response, c);
}
