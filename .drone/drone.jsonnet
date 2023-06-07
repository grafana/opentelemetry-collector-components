local drone = import 'lib/drone/drone.libsonnet';
local images = import 'lib/drone/images.libsonnet';
local triggers = import 'lib/drone/triggers.libsonnet';
local vault = import 'lib/vault/vault.libsonnet';

local pipeline = drone.pipeline;
local step = drone.step;
local withInlineStep = drone.withInlineStep;
local withStep = drone.withStep;
local withSteps = drone.withSteps;

local dockerPluginName = 'plugins/gcr';

local dockerPluginBaseSettings = {
  registry: 'ghcr.io',
  repo: 'grafana/opentelemetry-collector-components',
  json_key: {
    from_secret: 'gcr_admin',
  },
};
// TODO can we get these values from env variable DISTRIBUTIONS?
local apps = [
  'otel-grafana',
//  'tracing',
//  'sidecar'
];
local goos = [
//  'windows',
  'linux',
//  'darwin'
];

local goarch = [
//  '386',
  'amd64',
  'arm64',
//  'ppc64le'
];

local withImagePullSecrets = {
  image_pull_secrets: ['dockerconfigjson'],
};

local generateTags = {
  step: step('generate tags', $.commands),
  commands: [
    'DOCKER_TAG=$(bash scripts/generate-tags.sh)',
    // `.tag` is the file consumed by the `deploy-image` plugin.
    'echo -n "$${DOCKER_TAG}" > .tag',
    // `.tags` is the file consumed by the Docker (GCR included) plugins to tag the built Docker image accordingly.
    'echo -n "$${DOCKER_TAG}" > .tags',
    // Print the contents of .tags for debugging purposes.
    'tail -n +1 .tags',
  ],
};

local buildDistributions = {
  step: step('build distributions', $.commands),
  commands: [
    'make generate-sources',
  ],
};

local buildBinaries = {
  step(os, arch): step(
    'build binaries',
    $.commands,
    platform={
      os: os,
      arch: arch,
    },
  ),
  commands: [
    'go build -C ./distributions/%s/_build -ldflags="-s -w" -trimpath' % app
  for app in apps],
  environment: {
    CGO_ENABLED: 0,
  },
};

local buildAndPushImages = {
  step(app, os, arch): step(
    '%s: build and push' % app,
    [],
    image=buildAndPushImages.pluginName,
    settings=buildAndPushImages.settings(app),
    platform=buildAndPushImages.platform(os, arch),
  ),
  pluginName: 'plugins/docker',

  // settings generates the CI Pipeline step settings
  settings(app): {
    repo: $._repo(app),
    registry: $._registry,
    dockerfile: './distributions/%s/Dockerfile' % app,
    // TODO use DOCKER_REPO_OWNER env instead of 'grafana'
    username: 'grafana',
    password: { from_secret: 'gh_token' },
  },

  // image generates the image for the given app
  image(app): $._registry + '/' + $._repo(app),
  platform(os, arch): {
    os: os,
    arch: arch,
  },
  _repo(app):: 'grafana/opentelemetry-collector-components/' + app,
  _registry:: 'ghcr.io',
};

[
  pipeline('build distributions')
  + withStep(buildDistributions.step)
  + triggers.pr
  + triggers.main,
] +
[
  pipeline('build binaries for %s %s' % [os, arch], depends_on=['build distributions'])
  + withStep(buildBinaries.step(os, arch))
  + triggers.pr
  + triggers.main,
  for os in goos
  for arch in goarch
]
//+ [
//  pipeline('build and push images for linux and amd64', depends_on=['build distributions'])
//  + withStep(generateTags.step)
//  + withSteps([buildAndPushImages.step(app, 'linux', 'amd64') for app in apps])
//  + withImagePullSecrets
//  + triggers.pr
//]
+ std.filter(function(p) p != null, [
  if (os == 'darwin' && arch == '386') || (os == 'windows' && arch == 'arm64') then null else
    pipeline('build and push images for os: %s, arch: %s' % [os, arch], depends_on=['build distributions'])
    + withStep(generateTags.step)
    + withSteps([buildAndPushImages.step(app, os, arch) for app in apps])
    + withImagePullSecrets
    + triggers.pr
    + triggers.main,
  for os in goos
  for arch in goarch
])
+ [
  pipeline('launch argo workflow')
  + withStep(
    step(
      'launch workflow',
      commands=[],
      settings={
        namespace: 'otlp-gateway-v2-cd',
        token: { from_secret: 'argo_token' },
        command: std.strReplace(|||
          list
        |||, '\n', ' '),
//        add_ci_labels: true,
      },
      image='us.gcr.io/kubernetes-dev/drone/plugins/argo-cli'
    )
  )
  + withImagePullSecrets
  + triggers.main,
]
+ [
  vault.secret('dockerconfigjson', 'secret/data/common/gcr', '.dockerconfigjson'),
  vault.secret('gh_token', 'infra/data/ci/github/grafanabot', 'pat'),
  vault.secret('gcr_admin', 'infra/data/ci/gcr-admin', 'service-account'),
  vault.secret('argo_token', 'infra/data/ci/argo-workflows/trigger-service-account', 'token'),
]
