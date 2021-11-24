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
          image: niki2401/profiles-server:{{.ProfilesVersion}}
          ports:
            - containerPort: 8000
              protocol: TCP
          imagePullPolicy: IfNotPresent
  selector:
    matchLabels:
      app: profiles-server
