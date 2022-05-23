# Anyone seen Fast and Furious? :)

advanced_go_dev_mode = os.getenv('FAST_AND_FURIOUSER')

load('ext://restart_process', 'docker_build_with_restart')

if advanced_go_dev_mode:

    local_resource(
        'gitops-server',
        'GOOS=linux GOARCH=amd64 make gitops-server',
        deps=[
            './cmd',
            './pkg',
            './core',
            './api',
        ]
    )

    local_resource(
        'ui-server',
        'make ui',
        deps=[
            './ui',
        ]
    )

    docker_build_with_restart(
        'localhost:5001/weaveworks/wego-app',
        '.',
        only=[
            './bin',
        ],
        dockerfile="dev.dockerfile",
        entrypoint="/app/build/gitops-server --log-level=debug --insecure --dev-mode --dev-user {dev_user}".format(
            dev_user=os.getenv("DEV_USER", "wego-admin")
        ),
        live_update=[
            sync('./bin', '/app/build'),
        ],
    )
else:
    docker_build(
        'localhost:5001/weaveworks/wego-app',
        '.',
        dockerfile="gitops-server.dockerfile",
    )


k8s_yaml(helm('./charts/gitops-server', name='dev', namespace='flux-system', values='./tools/helm-values-dev.yaml'))
k8s_yaml(helm('./tools/charts/dev', name='dev', namespace='flux-system', values='./tools/charts/dev/values.yaml'))

k8s_resource('dev-weave-gitops', port_forwards='9001', resource_deps=['gitops-server', 'ui-server'] if advanced_go_dev_mode else [])
