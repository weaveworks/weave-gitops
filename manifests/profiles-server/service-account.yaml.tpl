apiVersion: v1
kind: ServiceAccount
metadata:
  name: profiles-server-service-account
  namespace: {{.Namespace}}
secrets:
  - name: profiles-server-service-account-token
