load('ext://restart_process', 'docker_build_with_restart')

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

def helmfiles(chart, values):
	watch_file(chart)
	watch_file(values)
	return local('./tools/bin/helm template dev {c} -f {v}'.format(c=chart, v=values))

k8s_yaml(helmfiles('./charts/gitops-server', './tools/helm-values-dev.yaml'))
k8s_yaml(helmfiles('./tools/charts/dev', './tools/charts/dev/values.yaml'))

k8s_resource('dev-weave-gitops', port_forwards='9001', resource_deps=['gitops-server', 'ui-server'])
