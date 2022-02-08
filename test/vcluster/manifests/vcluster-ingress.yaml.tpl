apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/backend-protocol: HTTPS
    nginx.ingress.kubernetes.io/ssl-passthrough: "true"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  name: {{.Name}}
  namespace: {{.Name}}
spec:
  rules:
  - host: {{.Name}}.k3s
    http:
      paths:
      - backend:
          service:
            name: {{.Name}}
            port:
              number: 443
        path: /
        pathType: ImplementationSpecific
