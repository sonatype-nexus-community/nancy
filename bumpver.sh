export LAST_VERSION=$(git describe --abbrev=0 --tags)
echo $LAST_VERSION

export LAST_PREFIX=$(cut -d'.' -f1,2 <<< $LAST_VERSION)
echo $LAST_PREFIX

export LAST_SUFFIX=$(cut -d'.' -f3 <<< $LAST_VERSION)
echo $LAST_SUFFIX

export NEW_SUFFIX=$(expr "$LAST_SUFFIX" + 1)
echo $NEW_SUFFIX

export NEW_VERSION="$LAST_PREFIX.$NEW_SUFFIX"
echo $NEW_VERSION

