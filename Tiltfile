load('ext://restart_process', 'docker_build_with_restart')

local_resource(
    'gitops-bin', 
    'GOOS=linux GOARCH=amd64 make bin', 
    deps=[
        './cmd', 
        './pkg',
    ]
)

docker_build_with_restart(
    'localhost:5001/weaveworks/wego-app', 
    '.',
    only=[
        './bin',
    ],
    dockerfile="dev.dockerfile",
    entrypoint='/app/build/gitops ui run -l',
    live_update=[
        sync('./bin', '/app/build'),
    ],
)

k8s_yaml([
    'tools/wego-app-dev.yaml',
])

k8s_resource('wego-app', port_forwards='9000', resource_deps=['gitops-bin'])
