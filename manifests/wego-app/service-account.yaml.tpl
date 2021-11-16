apiVersion: v1
kind: ServiceAccount
metadata:
  name: wego-app-service-account
  namespace: {{.Namespace}}
secrets:
  - name: wego-app-service-account-token
