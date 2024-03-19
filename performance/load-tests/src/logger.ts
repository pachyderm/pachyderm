import {max, mean, median, quantile} from 'simple-statistics';
import humanizeDuration, {Options} from 'humanize-duration';

interface Logger {
  writeSuccess(journey: string, index: number): void;
  writeError(journey: string, index: number, error: Error): void;
  writeTiming(url: string, timing: Timing): void;
  playback(): void;
}

type ErrorLog = {
  index: number;
  journey: string;
  error: Error;
  time: string;
};

type SuccessLog = {
  index: number;
  journey: string;
  time: string;
};

type Timing = {
  startTime: number;
  endTime?: number;
  domainLookupStart: number;
  domainLookupEnd: number;
  connectStart: number;
  secureConnectionStart: number;
  connectEnd: number;
  requestStart: number;
  responseStart: number;
  responseEnd: number;
};

type TimingLog = {
  [url: string]: Timing[];
};

class MemoryLogger implements Logger {
  private errorLog: ErrorLog[] = [];
  private successLog: SuccessLog[] = [];
  private timingLog: TimingLog = {};

  writeSuccess(journey: string, index: number): void {
    this.successLog.push({index, journey, time: new Date().toISOString()});
  }
  writeError(journey: string, index: number, error: Error): void {
    this.errorLog.push({index, journey, time: new Date().toISOString(), error});
  }
  writeTiming(url: string, timing: Timing) {
    this.timingLog[url] = this.timingLog[url] || [];
    this.timingLog[url].push(timing);
  }
  playback(): void {
    console.log('Logging successes:');
    console.table(this.successLog.map((el) => el));
    console.log('Logging failures:');
    this.errorLog.map((el) => console.dir(el, {depth: null}));

    const totalCount = this.successLog.length + this.errorLog.length;
    const errorCount = this.errorLog.length;

    console.log(`Total runs: ${totalCount}`);
    console.log(`Success count: ${this.successLog.length}`);
    console.log(`Failure count: ${errorCount}`);
    if (totalCount !== 0)
      console.log(
        `Failure %: ${((errorCount / totalCount) * 100).toFixed(2)}%`,
      );

    console.log('Timings:');

    const timings = Object.keys(this.timingLog).map((url) => {
      const urlTimings = this.timingLog[url].map(
        (timing) => timing.responseEnd,
      );

      return {
        url,
        numRequests: urlTimings.length,
        mean: mean(urlTimings),
        median: median(urlTimings),
        percentile90th: quantile(urlTimings, 0.9),
        max: max(urlTimings),
      };
    });
    const humanizeOpts: Options = {
      maxDecimalPoints: 3,
      round: true,
      units: ['s', 'ms'],
    };

    const formattedTimings = timings
      .sort((a, b) => (b.url > a.url ? -1 : 1))
      .map((timing) => ({
        url: timing.url,
        numRequests: timing.numRequests,
        mean: humanizeDuration(timing.mean, humanizeOpts),
        median: humanizeDuration(timing.median, humanizeOpts),
        percentile90th: humanizeDuration(timing.percentile90th, humanizeOpts),
        worst: humanizeDuration(timing.max, humanizeOpts),
      }));
    console.table(formattedTimings);
  }
}

export const statsLogger = new MemoryLogger();
