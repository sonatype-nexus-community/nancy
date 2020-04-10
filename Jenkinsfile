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
