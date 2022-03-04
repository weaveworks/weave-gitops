load('ext://restart_process', 'docker_build_with_restart')

local_resource(
    'gitops-server',
    'GOOS=linux GOARCH=amd64 make gitops-server',
    deps=[
        './cmd', 
        './pkg',
        "./core",
    ]
)

docker_build_with_restart(
    'localhost:5001/weaveworks/wego-app', 
    '.',
    only=[
        './bin',
    ],
    dockerfile="dev.dockerfile",
    entrypoint='/app/build/gitops-server -l',
    live_update=[
        sync('./bin', '/app/build'),
    ],
)

k8s_yaml([
    'tools/dev-manifests/wego-app.yaml',
    'tools/dev-manifests/role.yaml',
    'tools/dev-manifests/role-binding.yaml',
    'tools/dev-manifests/helm_watcher_role.yaml',
    'tools/dev-manifests/helm_watcher_role_binding.yaml',
    'tools/dev-manifests/service-account.yaml',
])

k8s_resource('wego-app', port_forwards='9000', resource_deps=['gitops-server'])
