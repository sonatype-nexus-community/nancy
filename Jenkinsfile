/*
 * Copyright 2018-present Sonatype Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
@Library(['private-pipeline-library', 'jenkins-shared']) _

dockerizedBuildPipeline(
  prepare: {
    githubStatusUpdate('pending')
  },
  buildAndTest: {
    sh '''
    go mod download
    go mod tidy
    go get -u github.com/jstemmer/go-junit-report
    go test ./... -v 2>&1 -p=1 | go-junit-report > test-results.xml
    CGO_ENABLED=0 GOOS=linux go build -o nancy .
    '''
  },
  vulnerabilityScan: {
    sh '''
    go list -m all | ./nancy iq -application nancy -stage stage 
    '''
  },
  testResults: [ 'test-results.xml' ],
  onSuccess: {
    githubStatusUpdate('success')
  },
  onFailure: {
    githubStatusUpdate('failure')
  }
)
