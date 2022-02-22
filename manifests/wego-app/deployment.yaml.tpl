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
          image: {{.AppImage}}:{{.AppVersion}}
          args: ["ui", "run", "--host", "0.0.0.0", "-l", "--helm-repo-namespace", "{{.Namespace}}"]
          ports:
            - containerPort: 9001
              protocol: TCP
          imagePullPolicy: IfNotPresent
  selector:
    matchLabels:
      app: wego-app
