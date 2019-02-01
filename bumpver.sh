#!/bin/bash
export LAST_PREFIX=$(cut -d'.' -f1,2 <<< $VERSION)
echo $LAST_PREFIX

export LAST_SUFFIX=$(cut -d'.' -f3 <<< $VERSION)
echo $LAST_SUFFIX

export NEW_SUFFIX=$(expr "$LAST_SUFFIX" + 1)
echo $NEW_SUFFIX

export NEW_VERSION="$LAST_PREFIX.$NEW_SUFFIX"
echo $NEW_VERSION

export VERSION=$NEW_VERSION
echo $VERSION
