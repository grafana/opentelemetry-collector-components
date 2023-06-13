local drone = import 'lib/drone/drone.libsonnet';
local images = import 'lib/drone/images.libsonnet';
local triggers = import 'lib/drone/triggers.libsonnet';
local vault = import 'lib/vault/vault.libsonnet';

local pipeline = drone.pipeline;
local step = drone.step;
local withStep = drone.withStep;
local withSteps = drone.withSteps;

local apps = [
  'otel-grafana',
];

local goos = 'linux';
local goarch = 'amd64';

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

local check = {
  step: step(
    'check',
    $.commands,
    environment=$.environment
  ),
  commands: [
    'make ci',
  ],
  environment: {
    DISTRIBUTIONS: std.join(",", apps)
  }
};

local buildDistributions = {
  step: step(
    'build distributions',
    $.commands,
    environment=$.environment
  ),
  commands: [
    'make generate-sources',
  ],
  environment: {
    DISTRIBUTIONS: std.join(",", apps)
  }
};

local buildBinaries = {
  step(app): step(
    'build binaries',
    $.commands,
    platform=$.platform,
    environment=$.environment
  ),
  commands: [
    'go build -o ../../../%(app)s -C distributions/%(app)s/_build -ldflags="-s -w" -trimpath' % {app: app}
  for app in apps],
  platform: {
    os: goos,
    arch: goarch,
  },
  environment: {
    CGO_ENABLED: 0,
  },
};

local buildAndPushImages = {
  step(app): step(
    '%s: build and push' % app,
    [],
    image=$.pluginName,
    settings=$.settings(app),
    platform=$.platform,
  ),
  pluginName: 'plugins/docker',

  // settings generates the CI Pipeline step settings
  settings(app): {
    repo: $._repo(app),
    registry: $._registry,
    dockerfile: './distributions/%s/Dockerfile' % app,
    username: '${DRONE_REPO_OWNER}',
    password: { from_secret: 'gh_token' },
  },

  // image generates the image for the given app
  image(app): $._registry + '/' + $._repo(app),
  platform: {
    os: goos,
    arch: goarch,
  },
  _repo(app):: $._registry + '/' +'grafana/opentelemetry-collector-components/' + app,
  _registry:: 'ghcr.io',
};

[
  pipeline('check')
  + withStep(check.step)
  + triggers.main,

  pipeline('test gcom api processor')
  + withStep(
    step('verify', commands=['cd components/processor/gcomapiprocessor && go test -v ./...'])
  )
  + triggers.main,

  pipeline('build', depends_on=['check', 'test gcom api processor'])
  + withStep(buildDistributions.step)
  + withSteps([buildBinaries.step(app) for app in apps])
  + withStep(generateTags.step)
  + withSteps([buildAndPushImages.step(app) for app in apps])
  + withImagePullSecrets
  + triggers.main,

  pipeline('launch argo workflow', depends_on=['build'])
  + withStep(generateTags.step)
  + withStep(
    step(
      'launch workflow',
      commands=[],
      settings={
        namespace: 'otlp-gateway-v2-cd',
        token: { from_secret: 'argo_token' },
        command: std.strReplace(|||
          submit --from workflowtemplate/deploy-otlp-gateway-v2
          --name otlp-gateway-v2-deploy-$(cat .tag)
          --parameter dockertag=$(cat .tag)
          --parameter commit=${DRONE_COMMIT}
          --parameter commit_author=${DRONE_COMMIT_AUTHOR}
          --parameter commit_link=${DRONE_COMMIT_LINK}
        |||, '\n', ' '),
        add_ci_labels: true,
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
  vault.secret('argo_token', 'infra/data/ci/argo-workflows/trigger-service-account', 'token'),
]
