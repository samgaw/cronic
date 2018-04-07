function run_cronic() {
  local crontab="$1"
  local timeout="${2:-1s}"
  timeout --preserve-status --kill-after "30s" "$timeout" \
    "${BATS_TEST_DIRNAME}/../cronic" ${CRONIC_ARGS:-} "$crontab" 2>&1
}

@test "it starts" {
  run_cronic "${BATS_TEST_DIRNAME}/noop.crontab"
}

@test "it runs a cron job" {
  n="$(run_cronic "${BATS_TEST_DIRNAME}/hello.crontab" 5s | grep -iE "hello from crontab.*channel=stdout" | wc -l)"
  [[ "$n" -gt 3 ]]
}

@test "it passes the environment through" {
  VAR="hello from foo" run_cronic "${BATS_TEST_DIRNAME}/env.crontab" | grep -iE "hello from foo.*channel=stdout"
}

@test "it overrides the environment with the crontab" {
  VAR="hello from foo" run_cronic "${BATS_TEST_DIRNAME}/override.crontab" | grep -iE "hello from bar.*channel=stdout"
}

@test "it warns when USER is set" {
  run_cronic "${BATS_TEST_DIRNAME}/user.crontab" 1s | grep -iE "processes will not.*USER="
}

@test "it warns when a job is falling behind" {
  run_cronic "${BATS_TEST_DIRNAME}/timeout.crontab" 1s | grep -iE "job took too long to run"
}

@test "it warns repeatedly when a job is still running" {
  n="$(run_cronic "${BATS_TEST_DIRNAME}/timeout.crontab" 1s | grep -iE "job is still running" | wc -l)"
  [[ "$n" -eq 2 ]]
}

@test "it supports debug logging " {
  CRONIC_ARGS="-debug" run_cronic "${BATS_TEST_DIRNAME}/hello.crontab" | grep -iE "debug"
}

@test "it supports JSON logging " {
  CRONIC_ARGS="-json" run_cronic "${BATS_TEST_DIRNAME}/noop.crontab" | grep -iE "^{"
}
