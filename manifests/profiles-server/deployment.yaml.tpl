apiVersion: apps/v1
kind: Deployment
metadata:
  name: profiles-server
  namespace: {{.Namespace}}
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: profiles-server
    spec:
      serviceAccountName: profiles-server-service-account
      containers:
        - name: profiles-server
          # TODO: change
          image: ghcr.io/weaveworks/profiles-server:{{.ProfilesVersion}}
          args: ["--helm-repo-namespace", "$(RUNTIME_NAMESPACE)"]
          ports:
            - containerPort: 8000
              protocol: TCP
          imagePullPolicy: IfNotPresent
          env:
            - name: RUNTIME_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
  selector:
    matchLabels:
      app: profiles-server
