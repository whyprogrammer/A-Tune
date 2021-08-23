#!/bin/sh

# RW_MODE
RW_MODE="-i 0 -i 1 -i 2"

# BLOCKSIZE
BLOCKSIZE=128k

# FILENAME
FILENAME=

# FILESIZE
FILESIZE=100g

iozone $RW_MODE -s $FILESIZE -r $BLOCKSIZE -f $FILENAME -Rb iozone-$FILESIZE-`date +%s`.xls
