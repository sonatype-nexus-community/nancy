#
# Copyright 2018-present Sonatype Inc.
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
FROM alpine:latest as builder

RUN apk add --no-cache ca-certificates \
    && update-ca-certificates 2>/dev/null || true \
    && rm -rf /var/cache/apk/*

#--------------------------------
# Deployment Image
#--------------------------------
FROM scratch

#Import from builder image
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY nancy /

ENTRYPOINT [ "/nancy" ]
