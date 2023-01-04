export const objects = [
  {
    obj: {
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        annotations: {
          "meta.helm.sh/release-name": "podinfo",
          "meta.helm.sh/release-namespace": "default",
        },
        creationTimestamp: "2022-11-29T19:54:27Z",
        labels: {
          "app.kubernetes.io/managed-by": "Helm",
          "app.kubernetes.io/name": "podinfo",
          "app.kubernetes.io/version": "6.2.3",
          "helm.sh/chart": "podinfo-6.2.3",
          "helm.toolkit.fluxcd.io/name": "podinfo",
          "helm.toolkit.fluxcd.io/namespace": "default",
        },
        managedFields: [
          {
            apiVersion: "v1",
            fieldsType: "FieldsV1",
            fieldsV1: {
              "f:metadata": {
                "f:annotations": {
                  ".": {},
                  "f:meta.helm.sh/release-name": {},
                  "f:meta.helm.sh/release-namespace": {},
                },
                "f:labels": {
                  ".": {},
                  "f:app.kubernetes.io/managed-by": {},
                  "f:app.kubernetes.io/name": {},
                  "f:app.kubernetes.io/version": {},
                  "f:helm.sh/chart": {},
                  "f:helm.toolkit.fluxcd.io/name": {},
                  "f:helm.toolkit.fluxcd.io/namespace": {},
                },
              },
              "f:spec": {
                "f:internalTrafficPolicy": {},
                "f:ports": {
                  ".": {},
                  'k:{"port":9898,"protocol":"TCP"}': {
                    ".": {},
                    "f:name": {},
                    "f:port": {},
                    "f:protocol": {},
                    "f:targetPort": {},
                  },
                  'k:{"port":9999,"protocol":"TCP"}': {
                    ".": {},
                    "f:name": {},
                    "f:port": {},
                    "f:protocol": {},
                    "f:targetPort": {},
                  },
                },
                "f:selector": {},
                "f:sessionAffinity": {},
                "f:type": {},
              },
            },
            manager: "helm-controller",
            operation: "Update",
            time: "2022-11-29T19:54:27Z",
          },
        ],
        name: "podinfo",
        namespace: "default",
        resourceVersion: "2790651",
        uid: "8a65d235-f05d-46ea-be3d-58a3c51dc2b2",
      },
      spec: {
        clusterIP: "10.96.116.248",
        clusterIPs: ["10.96.116.248"],
        internalTrafficPolicy: "Cluster",
        ipFamilies: ["IPv4"],
        ipFamilyPolicy: "SingleStack",
        ports: [
          { name: "http", port: 9898, protocol: "TCP", targetPort: "http" },
          { name: "grpc", port: 9999, protocol: "TCP", targetPort: "grpc" },
        ],
        selector: { "app.kubernetes.io/name": "podinfo" },
        sessionAffinity: "None",
        type: "ClusterIP",
      },
      status: { loadBalancer: {} },
    },
    clusterName: "Default",
    tenant: "",
    uid: "8a65d235-f05d-46ea-be3d-58a3c51dc2b2",
    children: [],
  },
  {
    obj: {
      apiVersion: "apps/v1",
      kind: "Deployment",
      metadata: {
        annotations: {
          "deployment.kubernetes.io/revision": "1",
          "meta.helm.sh/release-name": "podinfo",
          "meta.helm.sh/release-namespace": "default",
        },
        creationTimestamp: "2022-11-29T19:54:27Z",
        generation: 2,
        labels: {
          "app.kubernetes.io/managed-by": "Helm",
          "app.kubernetes.io/name": "podinfo",
          "app.kubernetes.io/version": "6.2.3",
          "helm.sh/chart": "podinfo-6.2.3",
          "helm.toolkit.fluxcd.io/name": "podinfo",
          "helm.toolkit.fluxcd.io/namespace": "default",
        },
        managedFields: [
          {
            apiVersion: "apps/v1",
            fieldsType: "FieldsV1",
            fieldsV1: {
              "f:metadata": {
                "f:annotations": {
                  ".": {},
                  "f:meta.helm.sh/release-name": {},
                  "f:meta.helm.sh/release-namespace": {},
                },
                "f:labels": {
                  ".": {},
                  "f:app.kubernetes.io/managed-by": {},
                  "f:app.kubernetes.io/name": {},
                  "f:app.kubernetes.io/version": {},
                  "f:helm.sh/chart": {},
                  "f:helm.toolkit.fluxcd.io/name": {},
                  "f:helm.toolkit.fluxcd.io/namespace": {},
                },
              },
              "f:spec": {
                "f:progressDeadlineSeconds": {},
                "f:revisionHistoryLimit": {},
                "f:selector": {},
                "f:strategy": {
                  "f:rollingUpdate": {
                    ".": {},
                    "f:maxSurge": {},
                    "f:maxUnavailable": {},
                  },
                  "f:type": {},
                },
                "f:template": {
                  "f:metadata": {
                    "f:annotations": {
                      ".": {},
                      "f:prometheus.io/port": {},
                      "f:prometheus.io/scrape": {},
                    },
                    "f:labels": { ".": {}, "f:app.kubernetes.io/name": {} },
                  },
                  "f:spec": {
                    "f:containers": {
                      'k:{"name":"podinfo"}': {
                        ".": {},
                        "f:command": {},
                        "f:env": {
                          ".": {},
                          'k:{"name":"PODINFO_UI_COLOR"}': {
                            ".": {},
                            "f:name": {},
                            "f:value": {},
                          },
                        },
                        "f:image": {},
                        "f:imagePullPolicy": {},
                        "f:livenessProbe": {
                          ".": {},
                          "f:exec": { ".": {}, "f:command": {} },
                          "f:failureThreshold": {},
                          "f:initialDelaySeconds": {},
                          "f:periodSeconds": {},
                          "f:successThreshold": {},
                          "f:timeoutSeconds": {},
                        },
                        "f:name": {},
                        "f:ports": {
                          ".": {},
                          'k:{"containerPort":9797,"protocol":"TCP"}': {
                            ".": {},
                            "f:containerPort": {},
                            "f:name": {},
                            "f:protocol": {},
                          },
                          'k:{"containerPort":9898,"protocol":"TCP"}': {
                            ".": {},
                            "f:containerPort": {},
                            "f:name": {},
                            "f:protocol": {},
                          },
                          'k:{"containerPort":9999,"protocol":"TCP"}': {
                            ".": {},
                            "f:containerPort": {},
                            "f:name": {},
                            "f:protocol": {},
                          },
                        },
                        "f:readinessProbe": {
                          ".": {},
                          "f:exec": { ".": {}, "f:command": {} },
                          "f:failureThreshold": {},
                          "f:initialDelaySeconds": {},
                          "f:periodSeconds": {},
                          "f:successThreshold": {},
                          "f:timeoutSeconds": {},
                        },
                        "f:resources": {
                          ".": {},
                          "f:limits": { ".": {}, "f:memory": {} },
                          "f:requests": {
                            ".": {},
                            "f:cpu": {},
                            "f:memory": {},
                          },
                        },
                        "f:terminationMessagePath": {},
                        "f:terminationMessagePolicy": {},
                        "f:volumeMounts": {
                          ".": {},
                          'k:{"mountPath":"/data"}': {
                            ".": {},
                            "f:mountPath": {},
                            "f:name": {},
                          },
                        },
                      },
                    },
                    "f:dnsPolicy": {},
                    "f:restartPolicy": {},
                    "f:schedulerName": {},
                    "f:securityContext": {},
                    "f:terminationGracePeriodSeconds": {},
                    "f:volumes": {
                      ".": {},
                      'k:{"name":"data"}': {
                        ".": {},
                        "f:emptyDir": {},
                        "f:name": {},
                      },
                    },
                  },
                },
              },
            },
            manager: "helm-controller",
            operation: "Update",
            time: "2022-11-29T19:54:27Z",
          },
          {
            apiVersion: "apps/v1",
            fieldsType: "FieldsV1",
            fieldsV1: {
              "f:metadata": {
                "f:annotations": { "f:deployment.kubernetes.io/revision": {} },
              },
              "f:status": {
                "f:conditions": {
                  ".": {},
                  'k:{"type":"Available"}': {
                    ".": {},
                    "f:lastTransitionTime": {},
                    "f:lastUpdateTime": {},
                    "f:message": {},
                    "f:reason": {},
                    "f:status": {},
                    "f:type": {},
                  },
                  'k:{"type":"Progressing"}': {
                    ".": {},
                    "f:lastTransitionTime": {},
                    "f:lastUpdateTime": {},
                    "f:message": {},
                    "f:reason": {},
                    "f:status": {},
                    "f:type": {},
                  },
                },
                "f:observedGeneration": {},
              },
            },
            manager: "kube-controller-manager",
            operation: "Update",
            subresource: "status",
            time: "2022-11-29T19:54:32Z",
          },
          {
            apiVersion: "apps/v1",
            fieldsType: "FieldsV1",
            fieldsV1: { "f:spec": { "f:replicas": {} } },
            manager: "flagger",
            operation: "Update",
            time: "2022-11-29T19:54:40Z",
          },
        ],
        name: "podinfo",
        namespace: "default",
        resourceVersion: "2790907",
        uid: "7341d127-8682-43c8-8caa-6c44740d988e",
      },
      spec: {
        progressDeadlineSeconds: 600,
        replicas: 0,
        revisionHistoryLimit: 10,
        selector: { matchLabels: { "app.kubernetes.io/name": "podinfo" } },
        strategy: {
          rollingUpdate: { maxSurge: "25%", maxUnavailable: 1 },
          type: "RollingUpdate",
        },
        template: {
          metadata: {
            annotations: {
              "prometheus.io/port": "9898",
              "prometheus.io/scrape": "true",
            },
            creationTimestamp: null,
            labels: { "app.kubernetes.io/name": "podinfo" },
          },
          spec: {
            containers: [
              {
                command: [
                  "./podinfo",
                  "--port=9898",
                  "--cert-path=/data/cert",
                  "--port-metrics=9797",
                  "--grpc-port=9999",
                  "--grpc-service-name=podinfo",
                  "--level=info",
                  "--random-delay=false",
                  "--random-error=false",
                ],
                env: [{ name: "PODINFO_UI_COLOR", value: "#34577c" }],
                image: "ghcr.io/stefanprodan/podinfo:6.2.3",
                imagePullPolicy: "IfNotPresent",
                livenessProbe: {
                  exec: {
                    command: [
                      "podcli",
                      "check",
                      "http",
                      "localhost:9898/healthz",
                    ],
                  },
                  failureThreshold: 3,
                  initialDelaySeconds: 1,
                  periodSeconds: 10,
                  successThreshold: 1,
                  timeoutSeconds: 5,
                },
                name: "podinfo",
                ports: [
                  { containerPort: 9898, name: "http", protocol: "TCP" },
                  {
                    containerPort: 9797,
                    name: "http-metrics",
                    protocol: "TCP",
                  },
                  { containerPort: 9999, name: "grpc", protocol: "TCP" },
                ],
                readinessProbe: {
                  exec: {
                    command: [
                      "podcli",
                      "check",
                      "http",
                      "localhost:9898/readyz",
                    ],
                  },
                  failureThreshold: 3,
                  initialDelaySeconds: 1,
                  periodSeconds: 10,
                  successThreshold: 1,
                  timeoutSeconds: 5,
                },
                resources: {
                  limits: { memory: "256Mi" },
                  requests: { cpu: "100m", memory: "64Mi" },
                },
                terminationMessagePath: "/dev/termination-log",
                terminationMessagePolicy: "File",
                volumeMounts: [{ mountPath: "/data", name: "data" }],
              },
            ],
            dnsPolicy: "ClusterFirst",
            restartPolicy: "Always",
            schedulerName: "default-scheduler",
            securityContext: {},
            terminationGracePeriodSeconds: 30,
            volumes: [{ emptyDir: {}, name: "data" }],
          },
        },
      },
      status: {
        conditions: [
          {
            lastTransitionTime: "2022-11-29T19:54:31Z",
            lastUpdateTime: "2022-11-29T19:54:31Z",
            message: "Deployment has minimum availability.",
            reason: "MinimumReplicasAvailable",
            status: "True",
            type: "Available",
          },
          {
            lastTransitionTime: "2022-11-29T19:54:27Z",
            lastUpdateTime: "2022-11-29T19:54:32Z",
            message:
              'ReplicaSet "podinfo-7949dd8fb4" has successfully progressed.',
            reason: "NewReplicaSetAvailable",
            status: "True",
            type: "Progressing",
          },
        ],
        observedGeneration: 2,
      },
    },
    clusterName: "Default",
    tenant: "",
    uid: "7341d127-8682-43c8-8caa-6c44740d988e",
    children: [
      {
        obj: {
          apiVersion: "apps/v1",
          kind: "ReplicaSet",
          metadata: {
            annotations: {
              "deployment.kubernetes.io/desired-replicas": "0",
              "deployment.kubernetes.io/max-replicas": "0",
              "deployment.kubernetes.io/revision": "1",
              "meta.helm.sh/release-name": "podinfo",
              "meta.helm.sh/release-namespace": "default",
            },
            creationTimestamp: "2022-11-29T19:54:27Z",
            generation: 2,
            labels: {
              "app.kubernetes.io/name": "podinfo",
              "pod-template-hash": "7949dd8fb4",
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
                      "f:meta.helm.sh/release-name": {},
                      "f:meta.helm.sh/release-namespace": {},
                    },
                    "f:labels": {
                      ".": {},
                      "f:app.kubernetes.io/name": {},
                      "f:pod-template-hash": {},
                    },
                    "f:ownerReferences": {
                      ".": {},
                      'k:{"uid":"7341d127-8682-43c8-8caa-6c44740d988e"}': {},
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
                          "f:app.kubernetes.io/name": {},
                          "f:pod-template-hash": {},
                        },
                      },
                      "f:spec": {
                        "f:containers": {
                          'k:{"name":"podinfo"}': {
                            ".": {},
                            "f:command": {},
                            "f:env": {
                              ".": {},
                              'k:{"name":"PODINFO_UI_COLOR"}': {
                                ".": {},
                                "f:name": {},
                                "f:value": {},
                              },
                            },
                            "f:image": {},
                            "f:imagePullPolicy": {},
                            "f:livenessProbe": {
                              ".": {},
                              "f:exec": { ".": {}, "f:command": {} },
                              "f:failureThreshold": {},
                              "f:initialDelaySeconds": {},
                              "f:periodSeconds": {},
                              "f:successThreshold": {},
                              "f:timeoutSeconds": {},
                            },
                            "f:name": {},
                            "f:ports": {
                              ".": {},
                              'k:{"containerPort":9797,"protocol":"TCP"}': {
                                ".": {},
                                "f:containerPort": {},
                                "f:name": {},
                                "f:protocol": {},
                              },
                              'k:{"containerPort":9898,"protocol":"TCP"}': {
                                ".": {},
                                "f:containerPort": {},
                                "f:name": {},
                                "f:protocol": {},
                              },
                              'k:{"containerPort":9999,"protocol":"TCP"}': {
                                ".": {},
                                "f:containerPort": {},
                                "f:name": {},
                                "f:protocol": {},
                              },
                            },
                            "f:readinessProbe": {
                              ".": {},
                              "f:exec": { ".": {}, "f:command": {} },
                              "f:failureThreshold": {},
                              "f:initialDelaySeconds": {},
                              "f:periodSeconds": {},
                              "f:successThreshold": {},
                              "f:timeoutSeconds": {},
                            },
                            "f:resources": {
                              ".": {},
                              "f:limits": { ".": {}, "f:memory": {} },
                              "f:requests": {
                                ".": {},
                                "f:cpu": {},
                                "f:memory": {},
                              },
                            },
                            "f:terminationMessagePath": {},
                            "f:terminationMessagePolicy": {},
                            "f:volumeMounts": {
                              ".": {},
                              'k:{"mountPath":"/data"}': {
                                ".": {},
                                "f:mountPath": {},
                                "f:name": {},
                              },
                            },
                          },
                        },
                        "f:dnsPolicy": {},
                        "f:restartPolicy": {},
                        "f:schedulerName": {},
                        "f:securityContext": {},
                        "f:terminationGracePeriodSeconds": {},
                        "f:volumes": {
                          ".": {},
                          'k:{"name":"data"}': {
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
                time: "2022-11-29T19:54:27Z",
              },
              {
                apiVersion: "apps/v1",
                fieldsType: "FieldsV1",
                fieldsV1: {
                  "f:status": { "f:observedGeneration": {}, "f:replicas": {} },
                },
                manager: "kube-controller-manager",
                operation: "Update",
                subresource: "status",
                time: "2022-11-29T19:54:40Z",
              },
            ],
            name: "podinfo-7949dd8fb4",
            namespace: "default",
            ownerReferences: [
              {
                apiVersion: "apps/v1",
                blockOwnerDeletion: true,
                controller: true,
                kind: "Deployment",
                name: "podinfo",
                uid: "7341d127-8682-43c8-8caa-6c44740d988e",
              },
            ],
            resourceVersion: "2790906",
            uid: "b668770b-e267-42ad-b4d0-11a0fcbeb8b6",
          },
          spec: {
            replicas: 0,
            selector: {
              matchLabels: {
                "app.kubernetes.io/name": "podinfo",
                "pod-template-hash": "7949dd8fb4",
              },
            },
            template: {
              metadata: {
                annotations: {
                  "prometheus.io/port": "9898",
                  "prometheus.io/scrape": "true",
                },
                creationTimestamp: null,
                labels: {
                  "app.kubernetes.io/name": "podinfo",
                  "pod-template-hash": "7949dd8fb4",
                },
              },
              spec: {
                containers: [
                  {
                    command: [
                      "./podinfo",
                      "--port=9898",
                      "--cert-path=/data/cert",
                      "--port-metrics=9797",
                      "--grpc-port=9999",
                      "--grpc-service-name=podinfo",
                      "--level=info",
                      "--random-delay=false",
                      "--random-error=false",
                    ],
                    env: [{ name: "PODINFO_UI_COLOR", value: "#34577c" }],
                    image: "ghcr.io/stefanprodan/podinfo:6.2.3",
                    imagePullPolicy: "IfNotPresent",
                    livenessProbe: {
                      exec: {
                        command: [
                          "podcli",
                          "check",
                          "http",
                          "localhost:9898/healthz",
                        ],
                      },
                      failureThreshold: 3,
                      initialDelaySeconds: 1,
                      periodSeconds: 10,
                      successThreshold: 1,
                      timeoutSeconds: 5,
                    },
                    name: "podinfo",
                    ports: [
                      { containerPort: 9898, name: "http", protocol: "TCP" },
                      {
                        containerPort: 9797,
                        name: "http-metrics",
                        protocol: "TCP",
                      },
                      { containerPort: 9999, name: "grpc", protocol: "TCP" },
                    ],
                    readinessProbe: {
                      exec: {
                        command: [
                          "podcli",
                          "check",
                          "http",
                          "localhost:9898/readyz",
                        ],
                      },
                      failureThreshold: 3,
                      initialDelaySeconds: 1,
                      periodSeconds: 10,
                      successThreshold: 1,
                      timeoutSeconds: 5,
                    },
                    resources: {
                      limits: { memory: "256Mi" },
                      requests: { cpu: "100m", memory: "64Mi" },
                    },
                    terminationMessagePath: "/dev/termination-log",
                    terminationMessagePolicy: "File",
                    volumeMounts: [{ mountPath: "/data", name: "data" }],
                  },
                ],
                dnsPolicy: "ClusterFirst",
                restartPolicy: "Always",
                schedulerName: "default-scheduler",
                securityContext: {},
                terminationGracePeriodSeconds: 30,
                volumes: [{ emptyDir: {}, name: "data" }],
              },
            },
          },
          status: { observedGeneration: 2, replicas: 0 },
        },
        clusterName: "Default",
        tenant: "",
        uid: "b668770b-e267-42ad-b4d0-11a0fcbeb8b6",
        children: [],
      },
    ],
  },
  {
    obj: {
      apiVersion: "apps/v1",
      kind: "ReplicaSet",
      metadata: {
        annotations: {
          "deployment.kubernetes.io/desired-replicas": "0",
          "deployment.kubernetes.io/max-replicas": "0",
          "deployment.kubernetes.io/revision": "1",
          "meta.helm.sh/release-name": "podinfo",
          "meta.helm.sh/release-namespace": "default",
        },
        creationTimestamp: "2022-11-29T19:54:27Z",
        generation: 2,
        labels: {
          "app.kubernetes.io/name": "podinfo",
          "pod-template-hash": "7949dd8fb4",
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
                  "f:meta.helm.sh/release-name": {},
                  "f:meta.helm.sh/release-namespace": {},
                },
                "f:labels": {
                  ".": {},
                  "f:app.kubernetes.io/name": {},
                  "f:pod-template-hash": {},
                },
                "f:ownerReferences": {
                  ".": {},
                  'k:{"uid":"7341d127-8682-43c8-8caa-6c44740d988e"}': {},
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
                      "f:app.kubernetes.io/name": {},
                      "f:pod-template-hash": {},
                    },
                  },
                  "f:spec": {
                    "f:containers": {
                      'k:{"name":"podinfo"}': {
                        ".": {},
                        "f:command": {},
                        "f:env": {
                          ".": {},
                          'k:{"name":"PODINFO_UI_COLOR"}': {
                            ".": {},
                            "f:name": {},
                            "f:value": {},
                          },
                        },
                        "f:image": {},
                        "f:imagePullPolicy": {},
                        "f:livenessProbe": {
                          ".": {},
                          "f:exec": { ".": {}, "f:command": {} },
                          "f:failureThreshold": {},
                          "f:initialDelaySeconds": {},
                          "f:periodSeconds": {},
                          "f:successThreshold": {},
                          "f:timeoutSeconds": {},
                        },
                        "f:name": {},
                        "f:ports": {
                          ".": {},
                          'k:{"containerPort":9797,"protocol":"TCP"}': {
                            ".": {},
                            "f:containerPort": {},
                            "f:name": {},
                            "f:protocol": {},
                          },
                          'k:{"containerPort":9898,"protocol":"TCP"}': {
                            ".": {},
                            "f:containerPort": {},
                            "f:name": {},
                            "f:protocol": {},
                          },
                          'k:{"containerPort":9999,"protocol":"TCP"}': {
                            ".": {},
                            "f:containerPort": {},
                            "f:name": {},
                            "f:protocol": {},
                          },
                        },
                        "f:readinessProbe": {
                          ".": {},
                          "f:exec": { ".": {}, "f:command": {} },
                          "f:failureThreshold": {},
                          "f:initialDelaySeconds": {},
                          "f:periodSeconds": {},
                          "f:successThreshold": {},
                          "f:timeoutSeconds": {},
                        },
                        "f:resources": {
                          ".": {},
                          "f:limits": { ".": {}, "f:memory": {} },
                          "f:requests": {
                            ".": {},
                            "f:cpu": {},
                            "f:memory": {},
                          },
                        },
                        "f:terminationMessagePath": {},
                        "f:terminationMessagePolicy": {},
                        "f:volumeMounts": {
                          ".": {},
                          'k:{"mountPath":"/data"}': {
                            ".": {},
                            "f:mountPath": {},
                            "f:name": {},
                          },
                        },
                      },
                    },
                    "f:dnsPolicy": {},
                    "f:restartPolicy": {},
                    "f:schedulerName": {},
                    "f:securityContext": {},
                    "f:terminationGracePeriodSeconds": {},
                    "f:volumes": {
                      ".": {},
                      'k:{"name":"data"}': {
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
            time: "2022-11-29T19:54:27Z",
          },
          {
            apiVersion: "apps/v1",
            fieldsType: "FieldsV1",
            fieldsV1: {
              "f:status": { "f:observedGeneration": {}, "f:replicas": {} },
            },
            manager: "kube-controller-manager",
            operation: "Update",
            subresource: "status",
            time: "2022-11-29T19:54:40Z",
          },
        ],
        name: "podinfo-7949dd8fb4",
        namespace: "default",
        ownerReferences: [
          {
            apiVersion: "apps/v1",
            blockOwnerDeletion: true,
            controller: true,
            kind: "Deployment",
            name: "podinfo",
            uid: "7341d127-8682-43c8-8caa-6c44740d988e",
          },
        ],
        resourceVersion: "2790906",
        uid: "b668770b-e267-42ad-b4d0-11a0fcbeb8b6",
      },
      spec: {
        replicas: 0,
        selector: {
          matchLabels: {
            "app.kubernetes.io/name": "podinfo",
            "pod-template-hash": "7949dd8fb4",
          },
        },
        template: {
          metadata: {
            annotations: {
              "prometheus.io/port": "9898",
              "prometheus.io/scrape": "true",
            },
            creationTimestamp: null,
            labels: {
              "app.kubernetes.io/name": "podinfo",
              "pod-template-hash": "7949dd8fb4",
            },
          },
          spec: {
            containers: [
              {
                command: [
                  "./podinfo",
                  "--port=9898",
                  "--cert-path=/data/cert",
                  "--port-metrics=9797",
                  "--grpc-port=9999",
                  "--grpc-service-name=podinfo",
                  "--level=info",
                  "--random-delay=false",
                  "--random-error=false",
                ],
                env: [{ name: "PODINFO_UI_COLOR", value: "#34577c" }],
                image: "ghcr.io/stefanprodan/podinfo:6.2.3",
                imagePullPolicy: "IfNotPresent",
                livenessProbe: {
                  exec: {
                    command: [
                      "podcli",
                      "check",
                      "http",
                      "localhost:9898/healthz",
                    ],
                  },
                  failureThreshold: 3,
                  initialDelaySeconds: 1,
                  periodSeconds: 10,
                  successThreshold: 1,
                  timeoutSeconds: 5,
                },
                name: "podinfo",
                ports: [
                  { containerPort: 9898, name: "http", protocol: "TCP" },
                  {
                    containerPort: 9797,
                    name: "http-metrics",
                    protocol: "TCP",
                  },
                  { containerPort: 9999, name: "grpc", protocol: "TCP" },
                ],
                readinessProbe: {
                  exec: {
                    command: [
                      "podcli",
                      "check",
                      "http",
                      "localhost:9898/readyz",
                    ],
                  },
                  failureThreshold: 3,
                  initialDelaySeconds: 1,
                  periodSeconds: 10,
                  successThreshold: 1,
                  timeoutSeconds: 5,
                },
                resources: {
                  limits: { memory: "256Mi" },
                  requests: { cpu: "100m", memory: "64Mi" },
                },
                terminationMessagePath: "/dev/termination-log",
                terminationMessagePolicy: "File",
                volumeMounts: [{ mountPath: "/data", name: "data" }],
              },
            ],
            dnsPolicy: "ClusterFirst",
            restartPolicy: "Always",
            schedulerName: "default-scheduler",
            securityContext: {},
            terminationGracePeriodSeconds: 30,
            volumes: [{ emptyDir: {}, name: "data" }],
          },
        },
      },
      status: { observedGeneration: 2, replicas: 0 },
    },
    clusterName: "Default",
    tenant: "",
    uid: "b668770b-e267-42ad-b4d0-11a0fcbeb8b6",
    children: [],
  },
];
