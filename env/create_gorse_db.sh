#!/bin/bash

set -e

psql -v ON_ERROR_STOP=1 -U bit101 bit101 <<-EOSQL
    CREATE DATABASE gorse;
EOSQL