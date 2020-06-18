#!/bin/sh

kubectl port-forward svc/postgres 32228:5432
