import {createLogger, LogLevel, stdSerializers} from 'bunyan';

const {LOG_LEVEL} = process.env;

let level: LogLevel;

// Check if log level is described in decimal format
if (LOG_LEVEL?.match(/^\d+$/)) {
  level = parseInt(LOG_LEVEL, 10);
} else if (LOG_LEVEL === 'none') {
  level = 1000;
} else {
  level = (process.env.LOG_LEVEL as LogLevel) || 'trace';
}

const logger = createLogger({
  name: 'dash-api',
  level,
  serializers: {err: stdSerializers.err},
});

export default logger;
