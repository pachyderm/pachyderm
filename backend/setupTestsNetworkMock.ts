import mockServer from './src/mock/networkMock';
import {graphqlServer} from './src/testHelpers';

beforeAll(async () => {
  const [grpcPort, authPort] = await mockServer.start();

  process.env.GRPC_PORT = String(grpcPort);
  process.env.AUTH_PORT = String(authPort);

  // set this, so we can reference it from testHelpers
  process.env.ISSUER_URI = `http://localhost:${authPort}`;
  process.env.PACHD_ADDRESS = `localhost:${process.env.GRPC_PORT}`;

  const graphqlPort = await graphqlServer.start();
  process.env.GRAPHQL_PORT = graphqlPort;
});

beforeEach(() => {
  mockServer.removeAllServices();
});

afterAll(async () => {
  await Promise.all([mockServer.stop(), graphqlServer.stop()]);
});
