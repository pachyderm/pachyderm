#!/usr/bin/python
import sys
import os
import yaml
import optparse

NANOSECONDS = int(1e9)

# Output stats in a format similar to siege
# see http://www.joedog.org/index/siege-home
OUTPUT_FORMAT = """Stats for type: [{type}]
Transactions:            {trans:>11} hits
Availability:            {avail:>11.2f} %
Elapsed time:            {elapsed:>11.2f} secs
Data transferred:        {data:>11.2f} MB
Response time:           {resp_time:>11.2f} secs
Transaction rate:        {trans_rate:>11.2f} trans/sec
Throughput:              {data_rate:>11.2f} MB/sec
Concurrency:             {conc:>11.2f}
Successful transactions: {trans_success:>11}
Failed transactions:     {trans_fail:>11}
Longest transaction:     {trans_long:>11.2f}
Shortest transaction:    {trans_short:>11.2f}
"""

def parse_options():
    usage = "usage: %prog [options]"
    parser = optparse.OptionParser(usage=usage)
    parser.add_option(
        "-f", "--file", dest="input", metavar="FILE",
        help="Name of input YAML file. Default uses sys.stdin")
    parser.add_option(
        "-v", "--verbose", dest="verbose", action="store_true",
        help="Enable verbose output")

    (options, args) = parser.parse_args()

    if not options.input and os.isatty(sys.stdin.fileno()):
        parser.error("option -f required if no data is provided "
                     "in stdin")

    return (options, args)

def main():
    (options, args) = parse_options()

    total     = {}
    durations = {}
    min_time  = {}
    max_time  = {}
    errors    = {}
    success   = {}

    calculate_stats(options, total, durations, min_time, max_time, errors,
                    success)
    print_results(total, durations, min_time, max_time, errors, success)

def calculate_stats(options, total, durations, min_time, max_time, errors,
                    success):
    print 'Calculating statistics...'
    
    f = sys.stdin
    if options.input:
        f = file(options.input, 'r')

    for item in yaml.safe_load_all(f):
        type_ = item.get('type')
        if type_ not in ('r', 'w'):
            continue # ignore any invalid items

        if 'error' in item:
            errors[type_] = errors.get(type_, 0) + 1
            continue # skip rest of analysis for this item
        else:
            success[type_] = success.get(type_, 0) + 1

        # parse the item
        data_size = item['chunks'][-1][0]
        duration = item['duration']
        start = item['start']
        end = start + duration / float(NANOSECONDS)

        if options.verbose:
            print "[{type}] POSIX time: {start:>18.2f} - {end:<18.2f} " \
                  "{data:>11.2f} KB".format(
                type=type_,
                start=start,
                end=end,
                data=data_size / 1024.0, # convert to KB
                )

        # update time boundaries
        prev = min_time.setdefault(type_, start)
        if start < prev:
            min_time[type_] = start
        prev = max_time.setdefault(type_, end)
        if end > prev:
            max_time[type_] = end

        # save the duration
        if type_ not in durations:
            durations[type_] = []
        durations[type_].append(duration)

        # add to running totals
        total[type_] = total.get(type_, 0) + data_size

def print_results(total, durations, min_time, max_time, errors, success):
    for type_ in total.keys():
        trans_success = success.get(type_, 0)
        trans_fail    = errors.get(type_, 0)
        trans         = trans_success + trans_fail
        avail         = trans_success * 100.0 / trans
        elapsed       = max_time[type_] - min_time[type_]
        data          = total[type_] / 1024.0 / 1024.0 # convert to MB
        resp_time     = sum(durations[type_]) / float(NANOSECONDS) / \
                        len(durations[type_])
        trans_rate    = trans / elapsed
        data_rate     = data / elapsed
        conc          = trans_rate * resp_time
        trans_long    = max(durations[type_]) / float(NANOSECONDS)
        trans_short   = min(durations[type_]) / float(NANOSECONDS)

        print OUTPUT_FORMAT.format(
            type=type_,
            trans_success=trans_success,
            trans_fail=trans_fail,
            trans=trans,
            avail=avail,
            elapsed=elapsed,
            data=data,
            resp_time=resp_time,
            trans_rate=trans_rate,
            data_rate=data_rate,
            conc=conc,
            trans_long=trans_long,
            trans_short=trans_short,
            )

if __name__ == '__main__':
    main()

