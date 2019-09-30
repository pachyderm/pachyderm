#!/usr/bin/env bash

pip install pipenv
pipenv install --dev
mkdocs build
