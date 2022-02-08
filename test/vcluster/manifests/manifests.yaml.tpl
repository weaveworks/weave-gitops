---
# Source: vcluster/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: vc-{{.Name}}
  namespace: {{.Name}}
  labels:
    app: vcluster
    chart: "vcluster-0.5.3"
    release: "{{.Name}}"
    heritage: "Helm"
---
# Source: vcluster/templates/rbac/role.yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: {{.Name}}
  namespace: {{.Name}}
  labels:
    app: vcluster
    chart: "vcluster-0.5.3"
    release: "{{.Name}}"
    heritage: "Helm"
rules:
  - apiGroups: [""]
    resources: ["configmaps", "secrets", "services", "pods", "pods/attach", "pods/portforward", "pods/exec", "endpoints", "persistentvolumeclaims"]
    verbs: ["create", "delete", "patch", "update", "get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events", "pods/log"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses"]
    verbs: ["create", "delete", "patch", "update", "get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["statefulsets", "replicasets", "deployments"]
    verbs: ["get", "list", "watch"]
---
# Source: vcluster/templates/rbac/rolebinding.yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: {{.Name}}
  namespace: {{.Name}}
  labels:
    app: vcluster
    chart: "vcluster-0.5.3"
    release: "{{.Name}}"
    heritage: "Helm"
subjects:
  - kind: ServiceAccount
    name: vc-{{.Name}}
    namespace: {{.Name}}
roleRef:
  kind: Role
  name: {{.Name}}
  apiGroup: rbac.authorization.k8s.io
---
# Source: vcluster/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Name}}
  labels:
    app: vcluster
    chart: "vcluster-0.5.3"
    release: "{{.Name}}"
    heritage: "Helm"
spec:
  type: ClusterIP
  ports:
    - name: https
      port: 443
      targetPort: 8443
      protocol: TCP
  selector:
    app: vcluster
    release: {{.Name}}
---
# Source: vcluster/templates/statefulset-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}-headless
  namespace: {{.Name}}
  labels:
    app: {{.Name}}
    chart: "vcluster-0.5.3"
    release: "{{.Name}}"
    heritage: "Helm"
spec:
  ports:
    - name: https
      port: 443
      targetPort: 8443
      protocol: TCP
  clusterIP: None
  selector:
    app: vcluster
    release: "{{.Name}}"
---
# Source: vcluster/templates/statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{.Name}}
  namespace: {{.Name}}
  labels:
    app: vcluster
    chart: "vcluster-0.5.3"
    release: "{{.Name}}"
    heritage: "Helm"
spec:
  serviceName: {{.Name}}-headless
  replicas: 1
  selector:
    matchLabels:
      app: vcluster
      release: {{.Name}}
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName:
        resources:
          requests:
            storage: 5Gi
  template:
    metadata:
      labels:
        app: vcluster
        release: {{.Name}}
    spec:
      terminationGracePeriodSeconds: 10
      nodeSelector:
        {}
      affinity:
        {}
      tolerations:
        []
      serviceAccountName: vc-{{.Name}}
      volumes:
      containers:
      - image: rancher/k3s:v1.21.4-k3s1
        name: vcluster
        # k3s has a problem running as pid 1 and disabled agents on cgroupv2
        # nodes as it will try to evacuate the cgroups there. Starting k3s
        # through a shell makes it non pid 1 and prevents this from happening
        command:
          - /bin/sh
        args:
          - -c
          - /bin/k3s
            server
            --write-kubeconfig=/data/k3s-config/kube-config.yaml
            --data-dir=/data
            --disable=traefik,servicelb,metrics-server,local-storage,coredns
            --disable-network-policy
            --disable-agent
            --disable-scheduler
            --disable-cloud-controller
            --flannel-backend=none
            --kube-controller-manager-arg=controllers=*,-nodeipam,-nodelifecycle,-persistentvolume-binder,-attachdetach,-persistentvolume-expander,-cloud-node-lifecycle
            --service-cidr=10.43.0.0/12
            && true
        env:
          []
        securityContext:
          allowPrivilegeEscalation: false
        volumeMounts:
          - mountPath: /data
            name: data
        resources:
          limits:
            memory: 2Gi
          requests:
            cpu: 200m
            memory: 256Mi
      - name: syncer
        image: "loftsh/vcluster:0.5.3"
        args:
          - --name={{.Name}}
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8443
            scheme: HTTPS
          failureThreshold: 10
          initialDelaySeconds: 60
          periodSeconds: 2
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8443
            scheme: HTTPS
          failureThreshold: 30
          periodSeconds: 2
        securityContext:
          allowPrivilegeEscalation: false
        env:
          - name: DEFAULT_IMAGE_REGISTRY
            value:
        volumeMounts:
          - mountPath: /data
            name: data
            readOnly: true
        resources:
          limits:
            memory: 1Gi
          requests:
            cpu: 100m
            memory: 128Mi
