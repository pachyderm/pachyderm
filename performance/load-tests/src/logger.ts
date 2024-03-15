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
}

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
    this.successLog.map((el) => console.log(el));
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
      const urlTimings = this.timingLog[url].map((timing) => timing.responseEnd);

      return {
        url,
        numRequests: urlTimings.length,
        mean: mean(urlTimings),
        median: median(urlTimings),
        percentile90th: quantile(urlTimings, 0.90),
        max: max(urlTimings),
      };
    });
    const humanizeOpts: Options = {maxDecimalPoints: 3, round: true, units: ['s', 'ms']};

    // Sort by best to worst average performance
    timings.sort((a, b) => a.mean - b.mean);

    timings.forEach((timing) => {
      console.log(`URL: ${timing.url}`);
      console.log(`Num requests: ${timing.numRequests}`);
      console.log(`Mean: ${humanizeDuration(timing.mean, humanizeOpts)}`);
      console.log(`Median: ${humanizeDuration(timing.median, humanizeOpts)}`);
      console.log(`90th Percentile: ${humanizeDuration(timing.percentile90th, humanizeOpts)}`);
      console.log(`Worst: ${humanizeDuration(timing.max, humanizeOpts)}`);
      console.log('---');
    });
  }
}

export const statsLogger = new MemoryLogger();
