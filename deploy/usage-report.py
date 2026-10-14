#!/usr/bin/env python3
"""Read recorded flow-session token usage without exposing session text or credentials."""
import argparse
import collections
import datetime
import json
import pathlib
import subprocess

METRICS = ("input_tokens", "cached_input_tokens", "cache_write_input_tokens",
           "output_tokens", "reasoning_output_tokens", "total_tokens")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", default="/var/lib/trading/flow/wallets")
    parser.add_argument("--database", default="trading_flow")
    parser.add_argument("--container", default="trading-staging-db-1")
    args = parser.parse_args()
    proc = subprocess.run(
        ["docker", "exec", "-i", args.container, "sh", "-c",
         'export PGPASSWORD="$POSTGRES_PASSWORD"; exec psql -X -At -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1"',
         "usage-report", args.database],
        input="SELECT id,code FROM wallets;", text=True, capture_output=True, check=True)
    codes = dict(line.split("|") for line in proc.stdout.strip().splitlines())
    root = pathlib.Path(args.root)
    groups = collections.defaultdict(collections.Counter)
    days = collections.defaultdict(collections.Counter)
    seen = set()
    first = last = None
    checks = []
    for file in sorted(root.glob("*/codex-home/sessions/**/*.jsonl")):
        group = codes.get(file.relative_to(root).parts[0], "SHARED_RESEARCH")
        thread = collections.Counter()
        reported = None
        for line in file.open():
            if '"token_usage_record"' not in line:
                continue
            try:
                record = json.loads(line)
            except json.JSONDecodeError:
                continue  # A currently running process may have an incomplete last line.
            if record.get("type") != "token_usage_record":
                continue
            payload = record.get("payload", {})
            response_id = payload.get("response_id")
            if not response_id or response_id in seen:
                continue
            seen.add(response_id)
            usage = payload.get("usage", {})
            stamp = record["timestamp"]
            first = min(first, stamp) if first else stamp
            last = max(last, stamp) if last else stamp
            day = stamp[:10]
            for metric in METRICS:
                value = int(usage.get(metric, 0) or 0)
                groups[group][metric] += value
                days[day][metric] += value
                thread[metric] += value
            groups[group]["responses"] += 1
            days[day]["responses"] += 1
            reported = payload.get("thread_token_usage")
        if reported:
            checks.append({"session": file.name,
                           "matches_thread_counter": all(thread[k] == int(reported.get(k, 0) or 0)
                                                         for k in METRICS)})
    total = collections.Counter()
    for counts in groups.values():
        total.update(counts)
    print(json.dumps({
        "version": "recorded-token-usage-v1",
        "generated_at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "scope": "Current flow wallet sessions and shared research sessions; excludes this development chat and other runtime environments.",
        "method": "Sum token_usage_record.usage once per unique response_id; compare to final thread_token_usage. Cached input is part of input; reasoning output is part of output.",
        "first_record_at": first, "last_record_at": last,
        "total": dict(total),
        "groups": {key: dict(value) for key, value in sorted(groups.items())},
        "daily_utc": {key: dict(value) for key, value in sorted(days.items())},
        "thread_checks": checks,
        "limitation": "Recorded token counts are not a dollar bill or a project-specific subscription quota percentage. Unrecorded/unfinished inference is not measured.",
    }, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
