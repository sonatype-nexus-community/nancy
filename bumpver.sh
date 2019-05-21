#!/bin/bash
#
# Copyright 2018 Sonatype Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

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
