#!/usr/bin/env bash

set -euo pipefail

function usage {
  echo "USAGE:"
  echo "-p|--push          - push images after build"
  echo "--commit-version   - commit .version and create tag"
  echo "--update-revision  - update .revision file with short git sha and tag if any"
  echo "--night-build      - use tag night-build as indicator of non-production build"
  echo "--shell            - shell '${IMAGE_LONG_NAME}' image after build"
  echo "--run              - run '${IMAGE_LONG_NAME}' image after build"
}

function image_size {
  local image_full_name=$1

  if [ -z "${image_full_name}" ]; then
    echo > 2 "Empty image name, can't get size. Ignoring"
    echo "N/A"
    return
  fi

  # http://www.pixelbeat.org/docs/numfmt.html
  local size_formatted=$( \
    docker image inspect "${image_full_name}" --format='{{.Size}}' | \
    numfmt --from=iec --to-unit=Ki --format="%'f KiB")

  echo $size_formatted
}

function inspect_build {
  local image_full_name=$1

  echo
  echo "> Image built:  ${image_full_name}"
  echo "> Image size:   $(image_size ${image_full_name})"
  echo
}

function run_container() {
  docker run --rm \
    --network iot \
    --name "$CONTAINER_NAME" \
    "$@"
}

_push_images=""
_commit_version=""
_update_revision=""
_night_build=""
_run_shell_in_container_after_build=""
_run_server_in_container_after_build=""

while [[ $# -gt 0 ]]
do
  key="${1}"
  case ${key} in
    -p|--push)
      _push_images="YES"
      shift ;;
    --commit-version)
      _commit_version="YES"
      shift ;;
    --update-revision)
      _update_revision="YES"
      shift ;;
    --night-build)
      _night_build="YES"
      shift ;;
    --shell)
      _run_shell_in_container_after_build="YES"
      shift ;;
    --run)
      _run_server_in_container_after_build="YES"
      shift ;;
    -h | --help)
      usage
      exit 0
      shift ;;
    *) shift ;;
  esac
done

USER_ID="andrewostroumov"
APP_VERSION=$(cat .version)

IMAGE_NAME="mobile-http-user-agent"
IMAGE_LONG_NAME="${USER_ID}/${IMAGE_NAME}"

IMAGE_MAJOR_VERSION=$(cut -d "." -f1 <<< "$APP_VERSION")
IMAGE_MINOR_VERSION=$(cut -d "." -f2 <<< "$APP_VERSION")
IMAGE_PATCH_VERSION=$(cut -d "." -f3 <<< "$APP_VERSION")

DOCKER_FILE="Dockerfile"
CONTAINER_NAME=${IMAGE_NAME}

echo "Building '${IMAGE_LONG_NAME}' image from '${DOCKER_FILE}'..."
echo "Push            : ${_push_images:-NO}"
echo "Commit version  : ${_commit_version:-NO} ($APP_VERSION)"
echo "Update revision : ${_update_revision:-NO}"
echo "Night build     : ${_night_build:-NO}"
echo "Shell           : ${_run_shell_in_container_after_build:-NO}"
echo "Run             : ${_run_server_in_container_after_build:-NO}"
echo

if [ -n "${_commit_version}" ]; then
  if [ -z "$(git status -s | grep .version)" ]; then
    echo "No new version to commit, please update .version file"
    exit 1
  fi

  git add .version
  git ci -m "Bump version to $APP_VERSION"
  git tag $APP_VERSION
  git push origin --tags
  git push origin
fi

SHORT_SHA=$(git rev-parse --short HEAD)
GIT_TAG=$(git tag --points-at HEAD)

IMAGE_VERSION="${IMAGE_MAJOR_VERSION}.${IMAGE_MINOR_VERSION}.${IMAGE_PATCH_VERSION}-${SHORT_SHA}"
IMAGE_PRIMARY_ID="${IMAGE_LONG_NAME}:${IMAGE_VERSION}"

NIGHT_BUILD="night-build"
NIGHT_BUILD_SHA="${NIGHT_BUILD}-${SHORT_SHA}"

#export $(grep -v '^#' .env.docker | xargs)

if [ -n "${_update_revision}" ]; then
  echo -n "${SHORT_SHA}" > .revision

  if [ -n "${GIT_TAG}" ]; then
    echo -n ":${GIT_TAG}" >> .revision
  fi
fi

if [ -n "${_night_build}" ]; then
  docker build \
    --pull \
    -t ${IMAGE_LONG_NAME}:${NIGHT_BUILD} \
    -t ${IMAGE_LONG_NAME}:${NIGHT_BUILD_SHA} \
    -f "${DOCKER_FILE}" \
    .

  inspect_build ${IMAGE_LONG_NAME}:${NIGHT_BUILD_SHA}
else
  docker build \
    --pull \
    -t ${IMAGE_PRIMARY_ID} \
    -t ${IMAGE_LONG_NAME}:${IMAGE_MAJOR_VERSION} \
    -t ${IMAGE_LONG_NAME}:${IMAGE_MAJOR_VERSION}.${IMAGE_MINOR_VERSION} \
    -t ${IMAGE_LONG_NAME}:${IMAGE_MAJOR_VERSION}.${IMAGE_MINOR_VERSION}.${IMAGE_PATCH_VERSION} \
    -f "${DOCKER_FILE}" \
    .

  inspect_build ${IMAGE_PRIMARY_ID}
fi

if [ -n "${_run_shell_in_container_after_build}" ]; then
  if [ -n "${_night_build}" ]; then
    run_container -it ${IMAGE_LONG_NAME}:${NIGHT_BUILD_SHA} sh
  else
    run_container -it ${IMAGE_PRIMARY_ID} sh
  fi
fi

if [ -n "${_run_server_in_container_after_build}" ]; then
    if [ -n "${_night_build}" ]; then
    run_container -it ${IMAGE_LONG_NAME}:${NIGHT_BUILD_SHA}
  else
    run_container -it ${IMAGE_PRIMARY_ID}
  fi
fi


if [ -n "${_push_images}" ]; then
  if [ -n "${_night_build}" ]; then
    docker push ${IMAGE_LONG_NAME}:${NIGHT_BUILD}
    docker push ${IMAGE_LONG_NAME}:${NIGHT_BUILD_SHA}
  else
    docker push ${IMAGE_PRIMARY_ID}
    docker push ${IMAGE_LONG_NAME}:${IMAGE_MAJOR_VERSION}
    docker push ${IMAGE_LONG_NAME}:${IMAGE_MAJOR_VERSION}.${IMAGE_MINOR_VERSION}
    docker push ${IMAGE_LONG_NAME}:${IMAGE_MAJOR_VERSION}.${IMAGE_MINOR_VERSION}.${IMAGE_PATCH_VERSION}
  fi
fi