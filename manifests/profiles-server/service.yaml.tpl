apiVersion: v1
kind: Service
metadata:
  name: profiles-server
  namespace: {{.Namespace}}
spec:
  selector:
    app: profiles-server
  ports:
    - protocol: TCP
      port: 8000
      targetPort: 8000
