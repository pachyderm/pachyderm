from boto.s3.connection import S3Connection
from boto.exception import BotoServerError
from boto.s3.key import Key
from httplib import BadStatusLine
from optparse import OptionParser
from .. import common

import traceback
import itertools
import random
import string
import struct
import yaml
import sys
import re


class DecisionGraphError(Exception):
    """ Raised when a node in a graph tries to set a header or
        key that was previously set by another node
    """
    def __init__(self, value):
        self.value = value

    def __str__(self):
        return repr(self.value)


class RecursionError(Exception):
    """Runaway recursion in string formatting"""

    def __init__(self, msg):
        self.msg = msg

    def __str__(self):
        return '{0.__doc__}: {0.msg!r}'.format(self)


def assemble_decision(decision_graph, prng):
    """ Take in a graph describing the possible decision space and a random
        number generator and traverse the graph to build a decision
    """
    return descend_graph(decision_graph, 'start', prng)


def descend_graph(decision_graph, node_name, prng):
    """ Given a graph and a particular node in that graph, set the values in
        the node's "set" list, pick a choice from the "choice" list, and
        recurse.  Finally, return dictionary of values
    """
    node = decision_graph[node_name]

    try:
        choice = make_choice(node['choices'], prng)
        if choice == '':
            decision = {}
        else:
            decision = descend_graph(decision_graph, choice, prng)
    except IndexError:
        decision = {}

    for key, choices in node['set'].iteritems():
        if key in decision:
            raise DecisionGraphError("Node %s tried to set '%s', but that key was already set by a lower node!" %(node_name, key))
        decision[key] = make_choice(choices, prng)

    if 'headers' in node:
        decision.setdefault('headers', [])

        for desc in node['headers']:
            try:
                (repetition_range, header, value) = desc
            except ValueError:
                (header, value) = desc
                repetition_range = '1'

            try:
                size_min, size_max = repetition_range.split('-', 1)
            except ValueError:
                size_min = size_max = repetition_range

            size_min = int(size_min)
            size_max = int(size_max)

            num_reps = prng.randint(size_min, size_max)
            if header in [h for h, v in decision['headers']]:
                    raise DecisionGraphError("Node %s tried to add header '%s', but that header already exists!" %(node_name, header))
            for _ in xrange(num_reps):
                decision['headers'].append([header, value])

    return decision


def make_choice(choices, prng):
    """ Given a list of (possibly weighted) options or just a single option!,
        choose one of the options taking weights into account and return the
        choice
    """
    if isinstance(choices, str):
        return choices
    weighted_choices = []
    for option in choices:
        if option is None:
            weighted_choices.append('')
            continue
        try:
            (weight, value) = option.split(None, 1)
            weight = int(weight)
        except ValueError:
            weight = 1
            value = option

        if value == 'null' or value == 'None':
            value = ''

        for _ in xrange(weight):
            weighted_choices.append(value)

    return prng.choice(weighted_choices)


def expand_headers(decision, prng):
    expanded_headers = {} 
    for header in decision['headers']:
        h = expand(decision, header[0], prng)
        v = expand(decision, header[1], prng)
        expanded_headers[h] = v
    return expanded_headers


def expand(decision, value, prng):
    c = itertools.count()
    fmt = RepeatExpandingFormatter(prng)
    new = fmt.vformat(value, [], decision)
    return new


class RepeatExpandingFormatter(string.Formatter):
    charsets = {
        'printable_no_whitespace': string.printable.translate(None, string.whitespace),
        'printable': string.printable,
        'punctuation': string.punctuation,
        'whitespace': string.whitespace,
        'digits': string.digits
    }

    def __init__(self, prng, _recursion=0):
        super(RepeatExpandingFormatter, self).__init__()
        # this class assumes it is always instantiated once per
        # formatting; use that to detect runaway recursion
        self.prng = prng
        self._recursion = _recursion

    def get_value(self, key, args, kwargs):
        fields = key.split(None, 1)
        fn = getattr(self, 'special_{name}'.format(name=fields[0]), None)
        if fn is not None:
            if len(fields) == 1:
                fields.append('')
            return fn(fields[1])

        val = super(RepeatExpandingFormatter, self).get_value(key, args, kwargs)
        if self._recursion > 5:
            raise RecursionError(key)
        fmt = self.__class__(self.prng, _recursion=self._recursion+1)

        n = fmt.vformat(val, args, kwargs)
        return n

    def special_random(self, args):
        arg_list = args.split()
        try:
            size_min, size_max = arg_list[0].split('-', 1)
        except ValueError:
            size_min = size_max = arg_list[0]
        except IndexError:
            size_min = '0'
            size_max = '1000'

        size_min = int(size_min)
        size_max = int(size_max)
        length = self.prng.randint(size_min, size_max)

        try:
            charset_arg = arg_list[1]
        except IndexError:
            charset_arg = 'printable'

        if charset_arg == 'binary' or charset_arg == 'binary_no_whitespace':
            num_bytes = length + 8
            tmplist = [self.prng.getrandbits(64) for _ in xrange(num_bytes / 8)]
            tmpstring = struct.pack((num_bytes / 8) * 'Q', *tmplist)
            if charset_arg == 'binary_no_whitespace':
                tmpstring = ''.join(c for c in tmpstring if c not in string.whitespace)
            return tmpstring[0:length]
        else:
            charset = self.charsets[charset_arg]
            return ''.join([self.prng.choice(charset) for _ in xrange(length)]) # Won't scale nicely


def parse_options():
    parser = OptionParser()
    parser.add_option('-O', '--outfile', help='write output to FILE. Defaults to STDOUT', metavar='FILE')
    parser.add_option('--seed', dest='seed', type='int',  help='initial seed for the random number generator')
    parser.add_option('--seed-file', dest='seedfile', help='read seeds for specific requests from FILE', metavar='FILE')
    parser.add_option('-n', dest='num_requests', type='int',  help='issue NUM requests before stopping', metavar='NUM')
    parser.add_option('-v', '--verbose', dest='verbose', action="store_true",  help='turn on verbose output')
    parser.add_option('-d', '--debug', dest='debug', action="store_true",  help='turn on debugging (very verbose) output')
    parser.add_option('--decision-graph', dest='graph_filename',  help='file in which to find the request decision graph')
    parser.add_option('--no-cleanup', dest='cleanup', action="store_false", help='turn off teardown so you can peruse the state of buckets after testing')

    parser.set_defaults(num_requests=5)
    parser.set_defaults(cleanup=True)
    parser.set_defaults(graph_filename='request_decision_graph.yml')
    return parser.parse_args()


def randomlist(seed=None):
    """ Returns an infinite generator of random numbers
    """
    rng = random.Random(seed)
    while True:
        yield rng.randint(0,100000) #100,000 seeds is enough, right?


def populate_buckets(conn, alt):
    """ Creates buckets and keys for fuzz testing and sets appropriate
        permissions. Returns a dictionary of the bucket and key names.
    """
    breadable = common.get_new_bucket(alt)
    bwritable = common.get_new_bucket(alt)
    bnonreadable = common.get_new_bucket(alt)

    oreadable = Key(breadable)
    owritable = Key(bwritable)
    ononreadable = Key(breadable)
    oreadable.set_contents_from_string('oreadable body')
    owritable.set_contents_from_string('owritable body')
    ononreadable.set_contents_from_string('ononreadable body')

    breadable.set_acl('public-read')
    bwritable.set_acl('public-read-write')
    bnonreadable.set_acl('private')
    oreadable.set_acl('public-read')
    owritable.set_acl('public-read-write')
    ononreadable.set_acl('private')

    return dict(
        bucket_readable=breadable.name,
        bucket_writable=bwritable.name,
        bucket_not_readable=bnonreadable.name,
        bucket_not_writable=breadable.name,
        object_readable=oreadable.key,
        object_writable=owritable.key,
        object_not_readable=ononreadable.key,
        object_not_writable=oreadable.key,
    )


def _main():
    """ The main script
    """
    (options, args) = parse_options()
    random.seed(options.seed if options.seed else None)
    s3_connection = common.s3.main
    alt_connection = common.s3.alt

    if options.outfile:
        OUT = open(options.outfile, 'w')
    else:
        OUT = sys.stderr

    VERBOSE = DEBUG = open('/dev/null', 'w')
    if options.verbose:
        VERBOSE = OUT
    if options.debug:
        DEBUG = OUT
        VERBOSE = OUT

    request_seeds = None
    if options.seedfile:
        FH = open(options.seedfile, 'r')
        request_seeds = [int(line) for line in FH if line != '\n']
        print>>OUT, 'Seedfile: %s' %options.seedfile
        print>>OUT, 'Number of requests: %d' %len(request_seeds)
    else:
        if options.seed:
            print>>OUT, 'Initial Seed: %d' %options.seed
        print>>OUT, 'Number of requests: %d' %options.num_requests
        random_list = randomlist(options.seed)
        request_seeds = itertools.islice(random_list, options.num_requests)

    print>>OUT, 'Decision Graph: %s' %options.graph_filename

    graph_file = open(options.graph_filename, 'r')
    decision_graph = yaml.safe_load(graph_file)

    constants = populate_buckets(s3_connection, alt_connection)
    print>>VERBOSE, "Test Buckets/Objects:"
    for key, value in constants.iteritems():
        print>>VERBOSE, "\t%s: %s" %(key, value)

    print>>OUT, "Begin Fuzzing..."
    print>>VERBOSE, '='*80
    for request_seed in request_seeds:
        print>>VERBOSE, 'Seed is: %r' %request_seed
        prng = random.Random(request_seed)
        decision = assemble_decision(decision_graph, prng)
        decision.update(constants)

        method = expand(decision, decision['method'], prng)
        path = expand(decision, decision['urlpath'], prng)

        try:
            body = expand(decision, decision['body'], prng)
        except KeyError:
            body = ''

        try:
            headers = expand_headers(decision, prng)
        except KeyError:
            headers = {}

        print>>VERBOSE, "%r %r" %(method[:100], path[:100])
        for h, v in headers.iteritems():
            print>>VERBOSE, "%r: %r" %(h[:50], v[:50])
        print>>VERBOSE, "%r\n" % body[:100]

        print>>DEBUG, 'FULL REQUEST'
        print>>DEBUG, 'Method: %r' %method
        print>>DEBUG, 'Path: %r' %path
        print>>DEBUG, 'Headers:'
        for h, v in headers.iteritems():
            print>>DEBUG, "\t%r: %r" %(h, v)
        print>>DEBUG, 'Body: %r\n' %body

        failed = False # Let's be optimistic, shall we?
        try:
            response = s3_connection.make_request(method, path, data=body, headers=headers, override_num_retries=1)
            body = response.read()
        except BotoServerError, e:
            response = e
            body = e.body
            failed = True
        except BadStatusLine, e:
            print>>OUT, 'FAILED: failed to parse response (BadStatusLine); probably a NUL byte in your request?'
            print>>VERBOSE, '='*80
            continue

        if failed:
            print>>OUT, 'FAILED:'
            OLD_VERBOSE = VERBOSE
            OLD_DEBUG = DEBUG
            VERBOSE = DEBUG = OUT
        print>>VERBOSE, 'Seed was: %r' %request_seed
        print>>VERBOSE, 'Response status code: %d %s' %(response.status, response.reason)
        print>>DEBUG, 'Body:\n%s' %body
        print>>VERBOSE, '='*80
        if failed:
            VERBOSE = OLD_VERBOSE
            DEBUG = OLD_DEBUG

    print>>OUT, '...done fuzzing'

    if options.cleanup:
        common.teardown()


def main():
    common.setup()
    try:
        _main()
    except Exception as e:
        traceback.print_exc()
        common.teardown()

