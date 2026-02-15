githubConfigUrl: https://github.com/${CODEXK8S_GITHUB_REPO}
githubConfigSecret: gha-runner-scale-set-secret
runnerScaleSetName: ${CODEXK8S_RUNNER_SCALE_SET_NAME}
minRunners: ${CODEXK8S_RUNNER_MIN}
maxRunners: ${CODEXK8S_RUNNER_MAX}
template:
  spec:
    containers:
      - name: runner
        image: ${CODEXK8S_RUNNER_IMAGE}
        command: ["/home/runner/run.sh"]
