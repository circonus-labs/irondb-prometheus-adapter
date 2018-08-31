#!/usr/bin/env bash

LOGFILE=$1
MSGOUTDIR=$2

mkdir -p $MSGOUTDIR && cd $MSGOUTDIR;

# log message format:
# Jun 21 21:04:20 api4-gcp-ia irondb-prometheus-adapter: 00004a10  63 68 65 63 6b 5f 75 75  69 64 22 3a 62 22 63 64  |check_uuid":b"cd|
# will output one ciml_*.log file per metric list flatbuffer message as well as one ciml_*.log.bin binary file per metric list for
# replay

grep -E ": [a-f0-9]{8}" $LOGFILE | \
    cut -f 1,2,3,4,5 -d ' ' --complement | \
    gawk '/[a-f0-9]{8}  .. 00 00 00 43 49 4d 4c/
            { ++a; fn=sprintf("./ciml_%02d.log", a);}
            {print $0 >> fn;}';


cd -;
