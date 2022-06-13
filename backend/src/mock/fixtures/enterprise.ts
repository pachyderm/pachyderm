import {GetStateResponse, State} from '@pachyderm/node-pachyderm';
import {timestampFromObject} from '@pachyderm/node-pachyderm/dist/builders/protobuf';
import {TokenInfo} from '@pachyderm/node-pachyderm/dist/proto/enterprise/enterprise_pb';

const active = new GetStateResponse()
  .setActivationCode('foo')
  .setState(State.ACTIVE)
  .setInfo(
    new TokenInfo().setExpires(
      timestampFromObject({seconds: 50596369, nanos: 0}),
    ),
  );
const inactive = new GetStateResponse().setState(State.NONE);
const expired = new GetStateResponse()
  .setState(State.EXPIRED)
  .setInfo(
    new TokenInfo().setExpires(
      timestampFromObject({seconds: 50596369, nanos: 0}),
    ),
  );
const expiring = new GetStateResponse()
  .setActivationCode('foo')
  .setState(State.ACTIVE)
  .setInfo(
    new TokenInfo().setExpires(
      timestampFromObject({
        seconds: Math.floor(Date.now() / 1000) + 24 * 60 * 60,
        nanos: 0,
      }),
    ),
  );

const enterpriseStates = {
  active,
  inactive,
  expired,
  expiring,
};

export default enterpriseStates;
