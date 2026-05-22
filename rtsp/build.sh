#!/usr/bin/env bash

set -euo pipefail

target="${1:-all}"
shift || true

goos="${GOOS:-}"
goarch="${GOARCH:-}"

output_suffix=""
if [[ "${goos}" == "windows" ]]; then
  output_suffix=".exe"
fi

usage() {
  echo "用法: ./build.sh [service|ui|all|run-ui|help] [-- <rtsp-ui 参数>]"
  echo "示例: ./build.sh all"
  echo "示例: ./build.sh run-ui -- -url rtsp://127.0.0.1:8554/live -port 8080"
  echo "示例: GOOS=windows GOARCH=amd64 ./build.sh all"
  echo "浏览器查看 "
  echo "1. rtsp-ui.exe -url rtsp://127.0.0.1:8554/live -port 8080 -ui "
  echo "2. 浏览器打开http://127.0.0.1:8080 "
}

build_service() {
  echo "==> 构建纯服务版"
  env GOOS="${goos}" GOARCH="${goarch}" go build -o "rtsp_pull_server${output_suffix}" .
}

build_ui() {
  echo "==> 构建内置 UI 版"
  env GOOS="${goos}" GOARCH="${goarch}" go build -tags ui -o "rtsp-ui${output_suffix}" .
}

extract_port() {
  local port="8080"
  local prev=""

  for arg in "$@"; do
    if [[ "${prev}" == "-port" ]]; then
      port="${arg}"
      break
    fi
    case "${arg}" in
      -port=*)
        port="${arg#-port=}"
        break
        ;;
    esac
    prev="${arg}"
  done

  echo "${port}"
}

open_browser() {
  local url="$1"

  if command -v xdg-open >/dev/null 2>&1; then
    xdg-open "${url}" >/dev/null 2>&1 &
    return
  fi

  if command -v open >/dev/null 2>&1; then
    open "${url}" >/dev/null 2>&1 &
    return
  fi

  if command -v cmd.exe >/dev/null 2>&1; then
    cmd.exe /C start "" "${url}" >/dev/null 2>&1
    return
  fi

  echo "请手动打开: ${url}"
}

run_ui() {
  if [[ -n "${goos}" && "${goos}" != "$(go env GOOS)" ]]; then
    echo "run-ui 不支持交叉编译目标 GOOS=${goos}，请先本机构建再运行"
    exit 1
  fi

  build_ui

  local port
  port="$(extract_port "$@")"
  local url="http://127.0.0.1:${port}/"

  echo "==> 启动内置 UI 服务"
  ./rtsp-ui${output_suffix} -ui "$@" &
  local pid=$!

  trap 'kill "${pid}" >/dev/null 2>&1 || true' EXIT INT TERM

  sleep 2
  echo "==> 打开页面: ${url}"
  open_browser "${url}"

  wait "${pid}"
}

case "${target}" in
  service)
    build_service
    ;;
  ui)
    build_ui
    ;;
  all)
    build_service
    build_ui
    ;;
  run-ui)
    run_ui "$@"
    ;;
  help|-h|--help)
    usage
    ;;
  *)
    usage
    exit 1
    ;;
esac

if [[ "${target}" != "run-ui" && "${target}" != "help" && "${target}" != "-h" && "${target}" != "--help" ]]; then
  echo "==> 构建完成"
fi
