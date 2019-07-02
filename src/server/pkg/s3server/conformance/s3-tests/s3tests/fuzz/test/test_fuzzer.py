"""
Unit-test suite for the S3 fuzzer

The fuzzer is a grammar-based random S3 operation generator
that produces random operation sequences in an effort to
crash the server.  This unit-test suite does not test
S3 servers, but rather the fuzzer infrastructure.

It works by running the fuzzer off of a simple grammar,
and checking the producted requests to ensure that they
include the expected sorts of operations in the expected
proportions.
"""
import sys
import itertools
import nose
import random
import string
import yaml

from ..headers import *

from nose.tools import eq_ as eq
from nose.tools import assert_true
from nose.plugins.attrib import attr

from ...functional.utils import assert_raises

_decision_graph = {}

def check_access_denied(fn, *args, **kwargs):
    e = assert_raises(boto.exception.S3ResponseError, fn, *args, **kwargs)
    eq(e.status, 403)
    eq(e.reason, 'Forbidden')
    eq(e.error_code, 'AccessDenied')


def build_graph():
    graph = {}
    graph['start'] = {
        'set': {},
        'choices': ['node2']
    }
    graph['leaf'] = {
        'set': {
            'key1': 'value1',
            'key2': 'value2'
        },
        'headers': [
            ['1-2', 'random-header-{random 5-10 printable}', '{random 20-30 punctuation}']
        ],
        'choices': []
    }
    graph['node1'] = {
        'set': {
            'key3': 'value3',
            'header_val': [
                '3 h1',
                '2 h2',
                'h3'
            ]
        },
        'headers': [
            ['1-1', 'my-header', '{header_val}'],
        ],
        'choices': ['leaf']
    }
    graph['node2'] = {
        'set': {
            'randkey': 'value-{random 10-15 printable}',
            'path': '/{bucket_readable}',
            'indirect_key1': '{key1}'
        },
        'choices': ['leaf']
    }
    graph['bad_node'] = {
        'set': {
            'key1': 'value1'
        },
        'choices': ['leaf']
    }
    graph['nonexistant_child_node'] = {
        'set': {},
        'choices': ['leafy_greens']
    }
    graph['weighted_node'] = {
        'set': {
            'k1': [
                'foo',
                '2 bar',
                '1 baz'
            ]
        },
        'choices': [
            'foo',
            '2 bar',
            '1 baz'
        ]
    }
    graph['null_choice_node'] = {
        'set': {},
        'choices': [None]
    }
    graph['repeated_headers_node'] = {
        'set': {},
        'headers': [
            ['1-2', 'random-header-{random 5-10 printable}', '{random 20-30 punctuation}']
        ],
        'choices': ['leaf']
    }
    graph['weighted_null_choice_node'] = {
        'set': {},
        'choices': ['3 null']
    }
    return graph


#def test_foo():
    #graph_file = open('request_decision_graph.yml', 'r')
    #graph = yaml.safe_load(graph_file)
    #eq(graph['bucket_put_simple']['set']['grantee'], 0)


def test_load_graph():
    graph_file = open('request_decision_graph.yml', 'r')
    graph = yaml.safe_load(graph_file)
    graph['start']


def test_descend_leaf_node():
    graph = build_graph()
    prng = random.Random(1)
    decision = descend_graph(graph, 'leaf', prng)

    eq(decision['key1'], 'value1')
    eq(decision['key2'], 'value2')
    e = assert_raises(KeyError, lambda x: decision[x], 'key3')


def test_descend_node():
    graph = build_graph()
    prng = random.Random(1)
    decision = descend_graph(graph, 'node1', prng)

    eq(decision['key1'], 'value1')
    eq(decision['key2'], 'value2')
    eq(decision['key3'], 'value3')


def test_descend_bad_node():
    graph = build_graph()
    prng = random.Random(1)
    assert_raises(DecisionGraphError, descend_graph, graph, 'bad_node', prng)


def test_descend_nonexistant_child():
    graph = build_graph()
    prng = random.Random(1)
    assert_raises(KeyError, descend_graph, graph, 'nonexistant_child_node', prng)


def test_expand_random_printable():
    prng = random.Random(1)
    got = expand({}, '{random 10-15 printable}', prng)
    eq(got, '[/pNI$;92@')


def test_expand_random_binary():
    prng = random.Random(1)
    got = expand({}, '{random 10-15 binary}', prng)
    eq(got, '\xdfj\xf1\xd80>a\xcd\xc4\xbb')


def test_expand_random_printable_no_whitespace():
    prng = random.Random(1)
    for _ in xrange(1000):
        got = expand({}, '{random 500 printable_no_whitespace}', prng)
        assert_true(reduce(lambda x, y: x and y, [x not in string.whitespace and x in string.printable for x in got]))


def test_expand_random_binary_no_whitespace():
    prng = random.Random(1)
    for _ in xrange(1000):
        got = expand({}, '{random 500 binary_no_whitespace}', prng)
        assert_true(reduce(lambda x, y: x and y, [x not in string.whitespace for x in got]))


def test_expand_random_no_args():
    prng = random.Random(1)
    for _ in xrange(1000):
        got = expand({}, '{random}', prng)
        assert_true(0 <= len(got) <= 1000)
        assert_true(reduce(lambda x, y: x and y, [x in string.printable for x in got]))


def test_expand_random_no_charset():
    prng = random.Random(1)
    for _ in xrange(1000):
        got = expand({}, '{random 10-30}', prng)
        assert_true(10 <= len(got) <= 30)
        assert_true(reduce(lambda x, y: x and y, [x in string.printable for x in got]))


def test_expand_random_exact_length():
    prng = random.Random(1)
    for _ in xrange(1000):
        got = expand({}, '{random 10 digits}', prng)
        assert_true(len(got) == 10)
        assert_true(reduce(lambda x, y: x and y, [x in string.digits for x in got]))


def test_expand_random_bad_charset():
    prng = random.Random(1)
    assert_raises(KeyError, expand, {}, '{random 10-30 foo}', prng)


def test_expand_random_missing_length():
    prng = random.Random(1)
    assert_raises(ValueError, expand, {}, '{random printable}', prng)


def test_assemble_decision():
    graph = build_graph()
    prng = random.Random(1)
    decision = assemble_decision(graph, prng)

    eq(decision['key1'], 'value1')
    eq(decision['key2'], 'value2')
    eq(decision['randkey'], 'value-{random 10-15 printable}')
    eq(decision['indirect_key1'], '{key1}')
    eq(decision['path'], '/{bucket_readable}')
    assert_raises(KeyError, lambda x: decision[x], 'key3')


def test_expand_escape():
    prng = random.Random(1)
    decision = dict(
        foo='{{bar}}',
        )
    got = expand(decision, '{foo}', prng)
    eq(got, '{bar}')


def test_expand_indirect():
    prng = random.Random(1)
    decision = dict(
        foo='{bar}',
        bar='quux',
        )
    got = expand(decision, '{foo}', prng)
    eq(got, 'quux')


def test_expand_indirect_double():
    prng = random.Random(1)
    decision = dict(
        foo='{bar}',
        bar='{quux}',
        quux='thud',
        )
    got = expand(decision, '{foo}', prng)
    eq(got, 'thud')


def test_expand_recursive():
    prng = random.Random(1)
    decision = dict(
        foo='{foo}',
        )
    e = assert_raises(RecursionError, expand, decision, '{foo}', prng)
    eq(str(e), "Runaway recursion in string formatting: 'foo'")


def test_expand_recursive_mutual():
    prng = random.Random(1)
    decision = dict(
        foo='{bar}',
        bar='{foo}',
        )
    e = assert_raises(RecursionError, expand, decision, '{foo}', prng)
    eq(str(e), "Runaway recursion in string formatting: 'foo'")


def test_expand_recursive_not_too_eager():
    prng = random.Random(1)
    decision = dict(
        foo='bar',
        )
    got = expand(decision, 100*'{foo}', prng)
    eq(got, 100*'bar')


def test_make_choice_unweighted_with_space():
    prng = random.Random(1)
    choice = make_choice(['foo bar'], prng)
    eq(choice, 'foo bar')

def test_weighted_choices():
    graph = build_graph()
    prng = random.Random(1)

    choices_made = {}
    for _ in xrange(1000):
        choice = make_choice(graph['weighted_node']['choices'], prng)
        if choices_made.has_key(choice):
            choices_made[choice] += 1
        else:
            choices_made[choice] = 1

    foo_percentage = choices_made['foo'] / 1000.0
    bar_percentage = choices_made['bar'] / 1000.0
    baz_percentage = choices_made['baz'] / 1000.0
    nose.tools.assert_almost_equal(foo_percentage, 0.25, 1)
    nose.tools.assert_almost_equal(bar_percentage, 0.50, 1)
    nose.tools.assert_almost_equal(baz_percentage, 0.25, 1)


def test_null_choices():
    graph = build_graph()
    prng = random.Random(1)
    choice = make_choice(graph['null_choice_node']['choices'], prng)

    eq(choice, '')


def test_weighted_null_choices():
    graph = build_graph()
    prng = random.Random(1)
    choice = make_choice(graph['weighted_null_choice_node']['choices'], prng)

    eq(choice, '')


def test_null_child():
    graph = build_graph()
    prng = random.Random(1)
    decision = descend_graph(graph, 'null_choice_node', prng)

    eq(decision, {})


def test_weighted_set():
    graph = build_graph()
    prng = random.Random(1)

    choices_made = {}
    for _ in xrange(1000):
        choice = make_choice(graph['weighted_node']['set']['k1'], prng)
        if choices_made.has_key(choice):
            choices_made[choice] += 1
        else:
            choices_made[choice] = 1

    foo_percentage = choices_made['foo'] / 1000.0
    bar_percentage = choices_made['bar'] / 1000.0
    baz_percentage = choices_made['baz'] / 1000.0
    nose.tools.assert_almost_equal(foo_percentage, 0.25, 1)
    nose.tools.assert_almost_equal(bar_percentage, 0.50, 1)
    nose.tools.assert_almost_equal(baz_percentage, 0.25, 1)


def test_header_presence():
    graph = build_graph()
    prng = random.Random(1)
    decision = descend_graph(graph, 'node1', prng)

    c1 = itertools.count()
    c2 = itertools.count()
    for header, value in decision['headers']:
        if header == 'my-header':
            eq(value, '{header_val}')
            assert_true(next(c1) < 1)
        elif header == 'random-header-{random 5-10 printable}':
            eq(value, '{random 20-30 punctuation}')
            assert_true(next(c2) < 2)
        else:
            raise KeyError('unexpected header found: %s' % header)

    assert_true(next(c1))
    assert_true(next(c2))


def test_duplicate_header():
    graph = build_graph()
    prng = random.Random(1)
    assert_raises(DecisionGraphError, descend_graph, graph, 'repeated_headers_node', prng)


def test_expand_headers():
    graph = build_graph()
    prng = random.Random(1)
    decision = descend_graph(graph, 'node1', prng)
    expanded_headers = expand_headers(decision, prng)

    for header, value in expanded_headers.iteritems():
        if header == 'my-header':
            assert_true(value in ['h1', 'h2', 'h3'])
        elif header.startswith('random-header-'):
            assert_true(20 <= len(value) <= 30)
            assert_true(string.strip(value, RepeatExpandingFormatter.charsets['punctuation']) is '')
        else:
            raise DecisionGraphError('unexpected header found: "%s"' % header)

