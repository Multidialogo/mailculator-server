#!/bin/sh

script_dir=$(dirname "$(realpath -s "$0")")

coverage() {
  temp_path="$script_dir/coverage.out"
  report_dir="$script_dir/.coverage"
  report_filename="report.html"
  report_path="$report_dir/$report_filename"

  export PAYLOAD_STORAGE_PATH=testdata/.out/json
  export MYSQL_HOST=127.0.0.1
  export MYSQL_PORT=3306
  export MYSQL_USER=root
  export MYSQL_PASSWORD=password
  export MYSQL_DATABASE=mailculator

  go mod tidy
  go test ./... -coverpkg=./... -coverprofile=$temp_path

  cov=$(go tool cover -func $temp_path | grep -E "^total" | grep -o -E "[0-9]*\.[0-9]*%$")
  echo "Total coverage: ${cov}"

  mkdir -p $report_dir
  go tool cover -html=$temp_path -o $report_path
  echo "Report exported at $report_path"
  rm $temp_path
}

if ! docker compose -f "$script_dir/compose.yml" --profile test-deps up -d --build --force-recreate; then
  echo "Could not start test dependencies"
  docker compose -f "$script_dir/compose.yml" --profile test-deps down --remove-orphans
  exit 1
fi

# run in subshell to avoid exporting env variables
(coverage)

docker compose -f "$script_dir/compose.yml" --profile test-deps down --remove-orphans
