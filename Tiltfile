load('ext://restart_process', 'docker_build_with_restart')

local_resource(
    'gitops-server',
    'GOOS=linux GOARCH=amd64 make gitops-server',
    deps=[
        './cmd',
        './pkg',
        './core',
        './charts',
    ]
)

docker_build_with_restart(
    'localhost:5001/weaveworks/wego-app',
    '.',
    only=[
        './bin',
        './server.rsa.crt',
        './server.rsa.key',
    ],
    dockerfile="dev.dockerfile",
    entrypoint='/app/build/gitops-server --tls-server-certificate=/app/build/server.rsa.crt --tls-server-key=/app/build/server.rsa.key',
    live_update=[
        sync('./bin', '/app/build'),
    ],
)

helm_stuff = helm(
  './charts/weave-gitops',
  name='weave-gitops',
  namespace='flux-system',
  values=['./tools/helm-values-dev.yaml'],
)

k8s_yaml(helm_stuff)

k8s_resource('wego-app', port_forwards='9001', resource_deps=['gitops-server'], links=['https://localhost:9000'])
