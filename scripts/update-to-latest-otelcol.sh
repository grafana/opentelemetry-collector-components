#!/bin/bash

REPO_DIR="$( cd "$(dirname $( dirname "${BASH_SOURCE[0]}" ))" &> /dev/null && pwd )"
COMPONENTS_DIR="${REPO_DIR}/components"

command -v gh 2>/dev/null
if [ $? != 0 ]; then
    echo "The command 'gh' is expected to exist and be configured in order to update to the latest otelcol."
    exit 1
fi

num_open_autoupdate_prs=$(gh pr list -l auto-update | wc -l)
if (( num_open_autoupdate_prs > 0 )); then
    echo "There are auto-update PRs waiting to be closed or merged. Skipping."
    exit 0
fi

while getopts d:c: flag
do
    case "${flag}" in
        d) directory=${OPTARG};;
        c) create_pr=${OPTARG};;
    esac
done

if [[ -z $directory ]]; then
    directory=$(mktemp -d)
    echo "Directory containing the release JSON files not provided. Created '${directory}' to host the latest from GitHub."
    gh api -H "Accept: application/vnd.github+json" /repos/open-telemetry/opentelemetry-collector/releases/latest > "${directory}/latest-core.json"
    gh api -H "Accept: application/vnd.github+json" /repos/open-telemetry/opentelemetry-collector-contrib/releases/latest > "${directory}/latest-contrib.json"
fi

# get the latest tag, without the "v" prefix
latest_core_version=$(jq -r .tag_name "${directory}/latest-core.json" | awk -F\/ '{print $NF}' | sed 's/^v//')
latest_contrib_version=$(jq -r .tag_name "${directory}/latest-contrib.json" | awk -F\/ '{print $NF}' | sed 's/^v//')

# in theory, we could have independent pull requests for each version bump, 
# but it's better to have the versions in sync at all times to prevent build failures
if [ $latest_core_version != $latest_contrib_version ]; then
    echo "The contrib and core versions aren't matching. This might be OK, but perhaps there's a release in process?"
    core_date=$(date -d $(jq -r .published_at "${directory}/latest-core.json") +%s)
    contrib_date=$(date -d $(jq -r .published_at "${directory}/latest-contrib.json") +%s)

    # the idea now is the following: if we just detected a new release of the core but not contrib yet,
    # then the release process might still be happening. Let's give it, say, 6 hours to complete.
    # If contrib's release is newer than core's, assume they are in sync already
    # If contrib is older than core, and core has been released more than 6 hours ago, go ahead anyway
    # Otherwise, skip the version bump
    if (( $contrib_date < $core_date )); then
        hours_between_releases=$(((contrib_date-core_date)/(60*60)))
        if (( hours_between_releases < 6 )); then
            echo "There seems to be a core release in process, skipping."
            exit 0
        fi
    fi
fi

branch="auto-update/core_${latest_core_version}_contrib_${latest_contrib_version}"

# perhaps there are changes, let's switch to a specific branch
git checkout -b "${branch}" main

# at this point, we are ready to start replacing the versions on the manifests
manifests=$(find ${REPO_DIR} -name manifest.yaml)
for manifest in $manifests; do
    echo "Updating $manifest"
    # the first token to replace is the otelcol_version
    sed -i "s~otelcol_version.*~otelcol_version: ${latest_core_version}~" $manifest

    # now, the collector gomod:
    sed -i "s~gomod: \(go\.opentelemetry\.io/collector.*\s\).*\$~gomod: \1v${latest_core_version}~" $manifest

    # and the contrib versions:
    # this captures a group with the "- gomod: component-path" content, and use it plus the version to compose the line 
    sed -i "s~\(.*github.com/open-telemetry/opentelemetry-collector-contrib/.*\s\).*~\1v${latest_contrib_version}~" $manifest
done

# Update the Makefile
sed -i "s/^OTELCOL_BUILDER_VERSION.*/OTELCOL_BUILDER_VERSION ?= ${latest_core_version}/" Makefile

# Update the go.mod files
gomods=$(find ${COMPONENTS_DIR} -name go.mod)
for gomod in $gomods; do
    pushd "$(dirname $gomod)"
    sed -i "s~\(go\.opentelemetry\.io/collector.*\s\).*\$~\1v${latest_core_version}~" $gomod
    sed -i "s~\(.*github.com/open-telemetry/opentelemetry-collector-contrib/.*\s\).*~\1v${latest_contrib_version}~" $gomod
    popd
done

git diff --quiet $manifests
if [[ $? == 0 ]]; then
    echo "We are already at the latest versions."
    exit 0
fi

# add only the files we might have changed
git add $manifests
git add Makefile
git add $gomods

# are there other changes?
git diff --quiet
if [[ $? != 0 ]]; then
    echo "More changes detected than expected! Aborting."
    exit 1
fi

git commit -sm "Bump OpenTelemetry core and/or contrib versions"
git push --set-upstream origin "${branch}"

if [[ "$create_pr" = true ]]; then
    echo "Creating the pull request on your behalf."
    gh pr create -l auto-update --title  "Bump OpenTelemetry core and/or contrib" --body "Use OpenTelemetry core v${latest_core_version} and contrib v${latest_contrib_version} in the manifests."
else
    echo "I could have created the pull request on your behalf if you had used the '-c' option."
fi
