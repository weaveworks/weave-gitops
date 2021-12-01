apiVersion: apps/v1
kind: Deployment
metadata:
  name: wego-app
  namespace: {{.Namespace}}
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: wego-app
    spec:
      serviceAccountName: wego-app-service-account
      containers:
        - name: wego-app
          image: ghcr.io/weaveworks/wego-app:{{.AppVersion}}
          args: ["ui", "run", "-l", "--helm-repo-namespace", "$(RUNTIME_NAMESPACE)"]
          ports:
            - containerPort: 9001
              protocol: TCP
          imagePullPolicy: IfNotPresent
          env:
            - name: RUNTIME_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
  selector:
    matchLabels:
      app: wego-app
