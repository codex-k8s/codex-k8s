githubConfigUrl: https://github.com/${GITHUB_REPO}
githubConfigSecret: gha-runner-scale-set-secret
runnerScaleSetName: ${RUNNER_SCALE_SET_NAME}
minRunners: ${RUNNER_MIN}
maxRunners: ${RUNNER_MAX}
template:
  spec:
    containers:
      - name: runner
        image: ${RUNNER_IMAGE}
        command: ["/home/runner/run.sh"]
