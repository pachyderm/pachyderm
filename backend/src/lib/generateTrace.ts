import {v4 as uuid} from 'uuid';

/*
  We will implement distributed tracing by logging standard messages for incoming and outgoing RPCs
  and HTTP requests in all Pachyderm products.  Typically, “traces” will be started by the
  pachyderm-proxy.  Spans are identified by a UUID4 (RFC4122 part 4.4), represented as an ASCII
  string with lowercase hex digits (RFC4122 part 3) with the 15th nibble of the string
  representation (“6a6ce55e-8146-420c-95ad-8df5dfd147f6”) set to:
    1: pachd-initiated interactive trace
    2: pachd-initiated background trace
    3: console-initiated trace
    4: pachyderm-proxy initiated trace (https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/request_id/uuid/v3/uuid.proto)
    9: pachyderm-proxy initiated trace
    a: pachyderm-proxy initiated trace (forced by server)
    b: pachyderm-proxy initiated trace (forced by client)
 */
export const generateConsoleTraceUuid = () => {
  const temp = uuid().split('');
  temp[14] = '3';
  const traceUuid = temp.join('');
  return traceUuid;
};
