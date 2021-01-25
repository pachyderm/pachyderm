import {SimulationLinkDatum, SimulationNodeDatum} from 'd3';

import {Node, Link} from '@graphqlTypes';

export interface NodeDatum extends Node, SimulationNodeDatum {}

export interface LinkDatum
  extends Omit<Link, 'source' | 'target'>,
    SimulationLinkDatum<NodeDatum> {}
