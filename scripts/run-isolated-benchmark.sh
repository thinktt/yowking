#!/usr/bin/env bash
set -Eeuo pipefail

isolated_cpus="${ISOLATED_CPUS:-14,15}"
housekeeping_cpus="${HOUSEKEEPING_CPUS:-0-13}"
isolated_slice="${ISOLATED_SLICE:-king-benchmark-isolated.slice}"
boost_path=/sys/devices/system/cpu/cpufreq/boost

if (($# == 0)); then
  echo "usage: $0 command [args...]" >&2
  exit 2
fi

sudo -n true
sudo systemctl daemon-reload

previous_boost=""
if [[ -e "$boost_path" ]]; then
  previous_boost="$(<"$boost_path")"
fi
previous_system_cpus="$(systemctl show system.slice -p AllowedCPUs --value)"
previous_user_cpus="$(systemctl show user.slice -p AllowedCPUs --value)"
previous_init_cpus="$(systemctl show init.scope -p AllowedCPUs --value)"
all_cpus="$(< /sys/fs/cgroup/cpuset.cpus.effective)"

restore_cpus() {
  local unit=$1
  local previous=$2
  if [[ -n "$previous" ]]; then
    sudo systemctl set-property --runtime "$unit" "AllowedCPUs=$previous"
    return
  fi
  sudo systemctl set-property --runtime "$unit" "AllowedCPUs=$all_cpus"
  sudo systemctl revert "$unit"
}

restore() {
  local status=$?
  trap - EXIT INT TERM HUP
  restore_cpus system.slice "$previous_system_cpus" || true
  restore_cpus user.slice "$previous_user_cpus" || true
  restore_cpus init.scope "$previous_init_cpus" || true
  sudo systemctl stop "$isolated_slice" || true
  if [[ -n "$previous_boost" ]]; then
    printf '%s\n' "$previous_boost" | sudo tee "$boost_path" >/dev/null || true
  fi
  echo "host settings restored: boost=${previous_boost:-unavailable} ordinary_cpus=${previous_system_cpus:-all}"
  exit "$status"
}
trap restore EXIT INT TERM HUP

sudo systemctl start "$isolated_slice"
sudo systemctl set-property --runtime "$isolated_slice" "AllowedCPUs=$isolated_cpus"
sudo systemctl set-property --runtime system.slice "AllowedCPUs=$housekeeping_cpus"
sudo systemctl set-property --runtime user.slice "AllowedCPUs=$housekeeping_cpus"
sudo systemctl set-property --runtime init.scope "AllowedCPUs=$housekeeping_cpus"
if [[ -n "$previous_boost" ]]; then
  printf '0\n' | sudo tee "$boost_path" >/dev/null
fi

echo "isolated benchmark environment:"
echo "  boost=${previous_boost:-unavailable}"
echo "  isolated_slice=$isolated_slice cpus=$(systemctl show "$isolated_slice" -p EffectiveCPUs --value)"
echo "  system.slice cpus=$(systemctl show system.slice -p EffectiveCPUs --value)"

"$@"
