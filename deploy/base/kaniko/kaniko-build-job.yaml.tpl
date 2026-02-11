apiVersion: batch/v1
kind: Job
metadata:
  name: ${CODEXK8S_KANIKO_JOB_NAME}
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s-kaniko
    app.kubernetes.io/component: ${CODEXK8S_KANIKO_COMPONENT}
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 600
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s-kaniko
        app.kubernetes.io/component: ${CODEXK8S_KANIKO_COMPONENT}
    spec:
      # For single-node staging we rely on the node loopback registry and therefore need hostNetwork.
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      restartPolicy: Never
      volumes:
        - name: workspace
          emptyDir: {}
      initContainers:
        - name: clone
          image: alpine/git:2.47.2
          imagePullPolicy: IfNotPresent
          env:
            - name: GIT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-git-token
                  key: token
            - name: CODEXK8S_GITHUB_REPO
              value: "${CODEXK8S_GITHUB_REPO}"
            - name: CODEXK8S_BUILD_REF
              value: "${CODEXK8S_BUILD_REF}"
          command:
            - sh
            - -ec
            - |
              git clone "https://x-access-token:${GIT_TOKEN}@github.com/${CODEXK8S_GITHUB_REPO}.git" /workspace
              cd /workspace
              git checkout --detach "${CODEXK8S_BUILD_REF}"
          volumeMounts:
            - name: workspace
              mountPath: /workspace
      containers:
        - name: kaniko
          image: gcr.io/kaniko-project/executor:v1.23.2-debug
          imagePullPolicy: IfNotPresent
          args:
            - --context=${CODEXK8S_KANIKO_CONTEXT}
            - --dockerfile=${CODEXK8S_KANIKO_DOCKERFILE}
            - --destination=${CODEXK8S_KANIKO_DESTINATION_LATEST}
            - --destination=${CODEXK8S_KANIKO_DESTINATION_SHA}
            - --insecure
            - --insecure-registry=${CODEXK8S_INTERNAL_REGISTRY_HOST}
            - --skip-tls-verify-registry=${CODEXK8S_INTERNAL_REGISTRY_HOST}
          resources:
            requests:
              cpu: ${CODEXK8S_KANIKO_RESOURCES_REQUEST_CPU}
              memory: ${CODEXK8S_KANIKO_RESOURCES_REQUEST_MEMORY}
            limits:
              cpu: ${CODEXK8S_KANIKO_RESOURCES_LIMIT_CPU}
              memory: ${CODEXK8S_KANIKO_RESOURCES_LIMIT_MEMORY}
          volumeMounts:
            - name: workspace
              mountPath: /workspace
