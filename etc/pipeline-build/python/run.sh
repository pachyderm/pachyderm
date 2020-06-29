#!/bin/sh
cd /pfs/__build__
pip install *.whl
cd /pfs/__source__
python main.py
