import express from 'express';

import gqlServer from 'gqlServer';

const PORT = process.env.PORT || '3000';
const app = express();

gqlServer.applyMiddleware({app, path: '/graphql'});

app.listen({port: PORT}, () =>
  console.log(
    `Server ready at http://localhost:${PORT}${gqlServer.graphqlPath}`,
  ),
);
