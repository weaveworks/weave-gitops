export const rs = {
  apiVersion: "apps/v1",
  kind: "ReplicaSet",
  metadata: {
    annotations: {
      "deployment.kubernetes.io/desired-replicas": "1",
      "deployment.kubernetes.io/max-replicas": "2",
      "deployment.kubernetes.io/revision": "2",
    },
    creationTimestamp: "2023-01-25T23:52:06Z",
    generation: 2,
    labels: {
      app: "notification-controller",
      "pod-template-hash": "85cf789788",
    },
    managedFields: [
      {
        apiVersion: "apps/v1",
        fieldsType: "FieldsV1",
        fieldsV1: {
          "f:metadata": {
            "f:annotations": {
              ".": {},
              "f:deployment.kubernetes.io/desired-replicas": {},
              "f:deployment.kubernetes.io/max-replicas": {},
              "f:deployment.kubernetes.io/revision": {},
            },
            "f:labels": {
              ".": {},
              "f:app": {},
              "f:pod-template-hash": {},
            },
            "f:ownerReferences": {
              ".": {},
              'k:{"uid":"29c765c3-9af3-469a-8b80-344921fc7724"}': {},
            },
          },
          "f:spec": {
            "f:replicas": {},
            "f:selector": {},
            "f:template": {
              "f:metadata": {
                "f:annotations": {
                  ".": {},
                  "f:prometheus.io/port": {},
                  "f:prometheus.io/scrape": {},
                },
                "f:labels": {
                  ".": {},
                  "f:app": {},
                  "f:pod-template-hash": {},
                },
              },
              "f:spec": {
                "f:containers": {
                  'k:{"name":"manager"}': {
                    ".": {},
                    "f:args": {},
                    "f:env": {
                      ".": {},
                      'k:{"name":"RUNTIME_NAMESPACE"}': {
                        ".": {},
                        "f:name": {},
                        "f:valueFrom": {
                          ".": {},
                          "f:fieldRef": {},
                        },
                      },
                    },
                    "f:image": {},
                    "f:imagePullPolicy": {},
                    "f:livenessProbe": {
                      ".": {},
                      "f:failureThreshold": {},
                      "f:httpGet": {
                        ".": {},
                        "f:path": {},
                        "f:port": {},
                        "f:scheme": {},
                      },
                      "f:periodSeconds": {},
                      "f:successThreshold": {},
                      "f:timeoutSeconds": {},
                    },
                    "f:name": {},
                    "f:ports": {
                      ".": {},
                      'k:{"containerPort":8080,"protocol":"TCP"}': {
                        ".": {},
                        "f:containerPort": {},
                        "f:name": {},
                        "f:protocol": {},
                      },
                      'k:{"containerPort":9090,"protocol":"TCP"}': {
                        ".": {},
                        "f:containerPort": {},
                        "f:name": {},
                        "f:protocol": {},
                      },
                      'k:{"containerPort":9292,"protocol":"TCP"}': {
                        ".": {},
                        "f:containerPort": {},
                        "f:name": {},
                        "f:protocol": {},
                      },
                      'k:{"containerPort":9440,"protocol":"TCP"}': {
                        ".": {},
                        "f:containerPort": {},
                        "f:name": {},
                        "f:protocol": {},
                      },
                    },
                    "f:readinessProbe": {
                      ".": {},
                      "f:failureThreshold": {},
                      "f:httpGet": {
                        ".": {},
                        "f:path": {},
                        "f:port": {},
                        "f:scheme": {},
                      },
                      "f:periodSeconds": {},
                      "f:successThreshold": {},
                      "f:timeoutSeconds": {},
                    },
                    "f:resources": {
                      ".": {},
                      "f:limits": {
                        ".": {},
                        "f:cpu": {},
                        "f:memory": {},
                      },
                      "f:requests": {
                        ".": {},
                        "f:cpu": {},
                        "f:memory": {},
                      },
                    },
                    "f:securityContext": {
                      ".": {},
                      "f:allowPrivilegeEscalation": {},
                      "f:capabilities": {
                        ".": {},
                        "f:drop": {},
                      },
                      "f:readOnlyRootFilesystem": {},
                      "f:runAsNonRoot": {},
                      "f:seccompProfile": {
                        ".": {},
                        "f:type": {},
                      },
                    },
                    "f:terminationMessagePath": {},
                    "f:terminationMessagePolicy": {},
                    "f:volumeMounts": {
                      ".": {},
                      'k:{"mountPath":"/tmp"}': {
                        ".": {},
                        "f:mountPath": {},
                        "f:name": {},
                      },
                    },
                  },
                },
                "f:dnsPolicy": {},
                "f:nodeSelector": {},
                "f:restartPolicy": {},
                "f:schedulerName": {},
                "f:securityContext": {
                  ".": {},
                  "f:fsGroup": {},
                },
                "f:serviceAccount": {},
                "f:serviceAccountName": {},
                "f:terminationGracePeriodSeconds": {},
                "f:volumes": {
                  ".": {},
                  'k:{"name":"temp"}': {
                    ".": {},
                    "f:emptyDir": {},
                    "f:name": {},
                  },
                },
              },
            },
          },
        },
        manager: "kube-controller-manager",
        operation: "Update",
        time: "2023-01-25T23:52:06Z",
      },
      {
        apiVersion: "apps/v1",
        fieldsType: "FieldsV1",
        fieldsV1: {
          "f:status": {
            "f:observedGeneration": {},
            "f:replicas": {},
          },
        },
        manager: "kube-controller-manager",
        operation: "Update",
        subresource: "status",
        time: "2023-01-25T23:52:59Z",
      },
    ],
    name: "notification-controller-85cf789788",
    namespace: "flux-system",
    ownerReferences: [
      {
        apiVersion: "apps/v1",
        blockOwnerDeletion: true,
        controller: true,
        kind: "Deployment",
        name: "notification-controller",
        uid: "29c765c3-9af3-469a-8b80-344921fc7724",
      },
    ],
    resourceVersion: "26778",
    uid: "a7f716f3-3c64-444d-81d4-9d5417570c29",
  },
  spec: {
    replicas: 0,
    selector: {
      matchLabels: {
        app: "notification-controller",
        "pod-template-hash": "85cf789788",
      },
    },
    template: {
      metadata: {
        annotations: {
          "prometheus.io/port": "8080",
          "prometheus.io/scrape": "true",
        },
        creationTimestamp: null,
        labels: {
          app: "notification-controller",
          "pod-template-hash": "85cf789788",
        },
      },
      spec: {
        containers: [
          {
            args: [
              "--watch-all-namespaces=true",
              "--log-level=info",
              "--log-encoding=json",
              "--enable-leader-election",
            ],
            env: [
              {
                name: "RUNTIME_NAMESPACE",
                valueFrom: {
                  fieldRef: {
                    apiVersion: "v1",
                    fieldPath: "metadata.namespace",
                  },
                },
              },
            ],
            image: "ghcr.io/fluxcd/notification-controller:v0.29.0",
            imagePullPolicy: "IfNotPresent",
            livenessProbe: {
              failureThreshold: 3,
              httpGet: {
                path: "/healthz",
                port: "healthz",
                scheme: "HTTP",
              },
              periodSeconds: 10,
              successThreshold: 1,
              timeoutSeconds: 1,
            },
            name: "manager",
            ports: [
              {
                containerPort: 9090,
                name: "http",
                protocol: "TCP",
              },
              {
                containerPort: 9292,
                name: "http-webhook",
                protocol: "TCP",
              },
              {
                containerPort: 8080,
                name: "http-prom",
                protocol: "TCP",
              },
              {
                containerPort: 9440,
                name: "healthz",
                protocol: "TCP",
              },
            ],
            readinessProbe: {
              failureThreshold: 3,
              httpGet: {
                path: "/readyz",
                port: "healthz",
                scheme: "HTTP",
              },
              periodSeconds: 10,
              successThreshold: 1,
              timeoutSeconds: 1,
            },
            resources: {
              limits: {
                cpu: "1",
                memory: "1Gi",
              },
              requests: {
                cpu: "100m",
                memory: "64Mi",
              },
            },
            securityContext: {
              allowPrivilegeEscalation: false,
              capabilities: {
                drop: ["ALL"],
              },
              readOnlyRootFilesystem: true,
              runAsNonRoot: true,
              seccompProfile: {
                type: "RuntimeDefault",
              },
            },
            terminationMessagePath: "/dev/termination-log",
            terminationMessagePolicy: "File",
            volumeMounts: [
              {
                mountPath: "/tmp",
                name: "temp",
              },
            ],
          },
        ],
        dnsPolicy: "ClusterFirst",
        nodeSelector: {
          "kubernetes.io/os": "linux",
        },
        restartPolicy: "Always",
        schedulerName: "default-scheduler",
        securityContext: {
          fsGroup: 1337,
        },
        serviceAccount: "notification-controller",
        serviceAccountName: "notification-controller",
        terminationGracePeriodSeconds: 10,
        volumes: [
          {
            emptyDir: {},
            name: "temp",
          },
        ],
      },
    },
  },
  status: {
    observedGeneration: 2,
    replicas: 0,
  },
};

export const pod = {
  apiVersion: "v1",
  kind: "Pod",
  metadata: {
    creationTimestamp: "2023-01-25T23:51:06Z",
    generateName: "podinfo-img-674ffb68d-",
    labels: {
      app: "podinfo-img",
      "pod-template-hash": "674ffb68d",
    },
    managedFields: [
      {
        apiVersion: "v1",
        fieldsType: "FieldsV1",
        fieldsV1: {
          "f:metadata": {
            "f:generateName": {},
            "f:labels": {
              ".": {},
              "f:app": {},
              "f:pod-template-hash": {},
            },
            "f:ownerReferences": {
              ".": {},
              'k:{"uid":"08d29a0b-da8f-490c-8c3c-877aff19b1dd"}': {},
            },
          },
          "f:spec": {
            "f:containers": {
              'k:{"name":"podinfo"}': {
                ".": {},
                "f:image": {},
                "f:imagePullPolicy": {},
                "f:name": {},
                "f:ports": {
                  ".": {},
                  'k:{"containerPort":9898,"protocol":"TCP"}': {
                    ".": {},
                    "f:containerPort": {},
                    "f:name": {},
                    "f:protocol": {},
                  },
                },
                "f:resources": {},
                "f:terminationMessagePath": {},
                "f:terminationMessagePolicy": {},
              },
            },
            "f:dnsPolicy": {},
            "f:enableServiceLinks": {},
            "f:restartPolicy": {},
            "f:schedulerName": {},
            "f:securityContext": {},
            "f:terminationGracePeriodSeconds": {},
          },
        },
        manager: "kube-controller-manager",
        operation: "Update",
        time: "2023-01-25T23:51:06Z",
      },
      {
        apiVersion: "v1",
        fieldsType: "FieldsV1",
        fieldsV1: {
          "f:status": {
            "f:conditions": {
              'k:{"type":"ContainersReady"}': {
                ".": {},
                "f:lastProbeTime": {},
                "f:lastTransitionTime": {},
                "f:status": {},
                "f:type": {},
              },
              'k:{"type":"Initialized"}': {
                ".": {},
                "f:lastProbeTime": {},
                "f:lastTransitionTime": {},
                "f:status": {},
                "f:type": {},
              },
              'k:{"type":"Ready"}': {
                ".": {},
                "f:lastProbeTime": {},
                "f:lastTransitionTime": {},
                "f:status": {},
                "f:type": {},
              },
            },
            "f:containerStatuses": {},
            "f:hostIP": {},
            "f:phase": {},
            "f:podIP": {},
            "f:podIPs": {
              ".": {},
              'k:{"ip":"10.244.0.30"}': {
                ".": {},
                "f:ip": {},
              },
            },
            "f:startTime": {},
          },
        },
        manager: "kubelet",
        operation: "Update",
        subresource: "status",
        time: "2023-02-06T16:49:30Z",
      },
    ],
    name: "podinfo-img-674ffb68d-zdhg5",
    namespace: "flux-system",
    ownerReferences: [
      {
        apiVersion: "apps/v1",
        blockOwnerDeletion: true,
        controller: true,
        kind: "ReplicaSet",
        name: "podinfo-img-674ffb68d",
        uid: "08d29a0b-da8f-490c-8c3c-877aff19b1dd",
      },
    ],
    resourceVersion: "1102793",
    uid: "effc29d4-1471-4034-a923-dcbd25d43b6a",
  },
  spec: {
    containers: [
      {
        image: "ghcr.io/stefanprodan/podinfo:5.0.0",
        imagePullPolicy: "IfNotPresent",
        name: "podinfo",
        ports: [
          {
            containerPort: 9898,
            name: "http",
            protocol: "TCP",
          },
        ],
        resources: {},
        terminationMessagePath: "/dev/termination-log",
        terminationMessagePolicy: "File",
        volumeMounts: [
          {
            mountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
            name: "kube-api-access-89r6l",
            readOnly: true,
          },
        ],
      },
    ],
    dnsPolicy: "ClusterFirst",
    enableServiceLinks: true,
    nodeName: "wge-dev-control-plane",
    preemptionPolicy: "PreemptLowerPriority",
    priority: 0,
    restartPolicy: "Always",
    schedulerName: "default-scheduler",
    securityContext: {},
    serviceAccount: "default",
    serviceAccountName: "default",
    terminationGracePeriodSeconds: 30,
    tolerations: [
      {
        effect: "NoExecute",
        key: "node.kubernetes.io/not-ready",
        operator: "Exists",
        tolerationSeconds: 300,
      },
      {
        effect: "NoExecute",
        key: "node.kubernetes.io/unreachable",
        operator: "Exists",
        tolerationSeconds: 300,
      },
    ],
    volumes: [
      {
        name: "kube-api-access-89r6l",
        projected: {
          defaultMode: 420,
          sources: [
            {
              serviceAccountToken: {
                expirationSeconds: 3607,
                path: "token",
              },
            },
            {
              configMap: {
                items: [
                  {
                    key: "ca.crt",
                    path: "ca.crt",
                  },
                ],
                name: "kube-root-ca.crt",
              },
            },
            {
              downwardAPI: {
                items: [
                  {
                    fieldRef: {
                      apiVersion: "v1",
                      fieldPath: "metadata.namespace",
                    },
                    path: "namespace",
                  },
                ],
              },
            },
          ],
        },
      },
    ],
  },
  status: {
    conditions: [
      {
        lastProbeTime: null,
        lastTransitionTime: "2023-01-25T23:51:06Z",
        status: "True",
        type: "Initialized",
      },
      {
        lastProbeTime: null,
        lastTransitionTime: "2023-02-06T16:49:23Z",
        status: "True",
        type: "Ready",
      },
      {
        lastProbeTime: null,
        lastTransitionTime: "2023-02-06T16:49:23Z",
        status: "True",
        type: "ContainersReady",
      },
      {
        lastProbeTime: null,
        lastTransitionTime: "2023-01-25T23:51:06Z",
        status: "True",
        type: "PodScheduled",
      },
    ],
    containerStatuses: [
      {
        containerID:
          "containerd://a302a701b738dd9755879d67748b820a2e5dd4f7a07c62bbbee95398854bae3b",
        image: "ghcr.io/stefanprodan/podinfo:5.0.0",
        imageID:
          "ghcr.io/stefanprodan/podinfo@sha256:d15a206e4ee462e82ab722ed84dfa514ab9ed8d85100d591c04314ae7c2162ee",
        lastState: {
          terminated: {
            containerID:
              "containerd://3b5a461aa494218c158c99770b3d34f90537f2fd03226d357870eb0f325b97bd",
            exitCode: 255,
            finishedAt: "2023-02-06T16:48:14Z",
            reason: "Unknown",
            startedAt: "2023-02-02T16:00:43Z",
          },
        },
        name: "podinfo",
        ready: true,
        restartCount: 11,
        started: true,
        state: {
          running: {
            startedAt: "2023-02-06T16:49:22Z",
          },
        },
      },
    ],
    hostIP: "172.18.0.2",
    phase: "Running",
    podIP: "10.244.0.30",
    podIPs: [
      {
        ip: "10.244.0.30",
      },
    ],
    qosClass: "BestEffort",
    startTime: "2023-01-25T23:51:06Z",
  },
};
