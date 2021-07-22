import {mockServer, graphqlServer} from './testHelpers';

beforeAll(async () => {
  const graphqlPort = await graphqlServer.start();
  const [grpcPort, authPort] = await mockServer.start();

  process.env.GRPC_PORT = String(grpcPort);
  process.env.AUTH_PORT = String(authPort);

  // set this, so we can reference it from testHelpers
  process.env.GRAPHQL_PORT = graphqlPort;
  process.env.ISSUER_URI = `http://localhost:${authPort}`;
  process.env.PACHD_ADDRESS = `localhost:${process.env.GRPC_PORT}`;
});

beforeEach(() => {
  mockServer.resetState();
});

afterAll(async () => {
  await Promise.all([mockServer.stop(), graphqlServer.stop()]);
});
