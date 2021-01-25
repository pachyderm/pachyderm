import {mockServer, graphqlServer} from './testHelpers';

beforeAll(async () => {
  await Promise.all([mockServer.start(), graphqlServer.start()]);
});

afterAll(async () => {
  await Promise.all([mockServer.stop(), graphqlServer.stop()]);
});
