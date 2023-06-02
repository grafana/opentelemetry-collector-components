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
  registry: 'us.gcr.io',
  repo: 'kubernetes-dev/otlp-gateway',
  json_key: {
    from_secret: 'gcr_admin',
  },
};

local apps = [
  'otlp-gateway',
];

local withImagePullSecrets = {
  image_pull_secrets: ['dockerconfigjson'],
};

local commentCoverageLintReport = {
  step: step('coverage + lint', $.commands, image=$.image, environment=$.environment),
  commands: [
    // Build drone utilities.
    'scripts/build-drone-utilities.sh',
    // Generate the raw coverage report.
    'go test -coverprofile=coverage.out ./...',
    // Process the raw coverage report.
    '.drone/coverage > coverage_report.out',
    // Generate the lint report.
    'scripts/generate-lint-report.sh',
    // Combine the reports.
    'cat coverage_report.out > report.out',
    'echo "" >> report.out',
    'cat lint.out >> report.out',
    // Submit the comment to GitHub.
    '.drone/ghcomment -id "Go coverage report:" -bodyfile report.out',
  ],
  environment: {
    GRAFANABOT_PAT: { from_secret: 'gh_token' },
  },
  image: images._images.goLint,
};

local imagePullSecrets = { image_pull_secrets: ['dockerconfigjson'] };

local generateTags = {
  step: step('generate tags', $.commands),
  commands: [
    'DOCKER_TAG=$(bash scripts/generate-tags.sh)',
    // `.tag` is the file consumed by the `deploy-image` plugin.
    'echo -n "$${DOCKER_TAG}" > .tag',
    // `.tags` is the file consumed by the Docker (GCR inluded) plugins to tag the built Docker image accordingly.
    'echo -n "$${DOCKER_TAG}" > .tags',
    // Print the contents of .tags for debugging purposes.
    'tail -n +1 .tags',
  ],
};

local buildAndPushImages = {
  // step builds the pipeline step to build and push a docker image
  step(app): step(
    '%s: build and push' % app,
    [],
    image=buildAndPushImages.pluginName,
    settings=buildAndPushImages.settings(app),
  ),

  pluginName: 'plugins/gcr',

  // settings generates the CI Pipeline step settings
  settings(app): {
    repo: $._repo(app),
    registry: $._registry,
    dockerfile: './Dockerfile',
    json_key: { from_secret: 'gcr_admin' },
    mirror: 'https://mirror.gcr.io',
    build_args: ['cmd=' + app],
  },

  // image generates the image for the given app
  image(app): $._registry + '/' + $._repo(app),

  _repo(app):: 'kubernetes-dev/' + app,
  _registry:: 'us.gcr.io',
};

local withDockerSockVolume = {
  volumes+: [
    {
      name: 'dockersock',
      path: '/var/run',
    },
  ],
};

local withDockerInDockerService = {
  services: [
    {
      name: 'docker',
      image: images._images.dind,
      entrypoint: ['dockerd'],
      command: [
        '--tls=false',
        '--host=tcp://0.0.0.0:2375',
        '--registry-mirror=https://mirror.gcr.io',
      ],
      privileged: true,
    } + withDockerSockVolume,
  ],
  environment+: {
    DOCKERD_ROOTLESS_ROOTLESSKIT_FLAGS: '-p 0.0.0.0:2376:2376/tcp',
  },
  volumes+: [
    {
      name: 'dockersock',
      temp: {},
    },
  ],
};

[
//  pipeline('check')
//  + withInlineStep('test', ['go test ./...'])
//  + triggers.pr
//  + triggers.main,
//
//  pipeline('coverageLintReport')
//  + withStep(commentCoverageLintReport.step)
//  + triggers.pr,
//
//  pipeline('build', depends_on=['check'])
//  + withStep(generateTags.step)
//  + withSteps([buildAndPushImages.step(app) for app in apps])
//  + imagePullSecrets
//  + triggers.pr
//  + triggers.main,

  pipeline('launch argo workflow', depends_on=['check', 'build'])
//  + withStep(generateTags.step)
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
        add_ci_labels: true,
      },
      image='us.gcr.io/kubernetes-dev/drone/plugins/argo-cli'
    )
  )
  + withImagePullSecrets
  + triggers.main
  + triggers.pr
]
+ [
  vault.secret('dockerconfigjson', 'secret/data/common/gcr', '.dockerconfigjson'),
  vault.secret('gh_token', 'infra/data/ci/github/grafanabot', 'pat'),
  vault.secret('gcr_admin', 'infra/data/ci/gcr-admin', 'service-account'),
  vault.secret('argo_token', 'infra/data/ci/argo-workflows/trigger-service-account', 'token'),
]
