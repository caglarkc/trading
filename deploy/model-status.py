#!/usr/bin/env python3
"""Read-only remote model/service status. Run on the trading host as operator/root."""
import json
import subprocess
import argparse
import re

parser = argparse.ArgumentParser()
parser.add_argument('--key', default='user-startup-2026-10-12')
args = parser.parse_args()
if not re.fullmatch(r'[a-z0-9-]{1,100}', args.key):
    parser.error('invalid job key')


def sql(query):
    result = subprocess.run(
        ['docker', 'exec', '-i', 'trading-staging-db-1', 'sh', '-c',
         'export PGPASSWORD="$POSTGRES_PASSWORD"; exec psql -X -At -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d trading_flow'],
        input=query, text=True, capture_output=True, check=True,
    )
    return json.loads(result.stdout)


service = subprocess.check_output(
    ['systemctl', 'show', 'trading-models.service', '-p', 'ActiveState',
     '-p', 'SubState', '-p', 'NRestarts', '-p', 'UnitFileState', '-p', 'MainPID'], text=True,
)
wallets = sql(f"""SELECT coalesce(json_agg(t ORDER BY code),'[]') FROM (
 SELECT w.code, j.state AS startup_state, j.last_error,
 d.state AS decision_state, d.finished_at, d.output->'actions' AS actions,
 (SELECT count(*) FROM receipts r WHERE r.wallet_id=w.id AND delivered_at IS NULL) AS pending_receipts,
 (SELECT count(*) FROM memory_versions m WHERE m.wallet_id=w.id) AS memory_versions
 FROM wallets w LEFT JOIN jobs j ON j.wallet_id=w.id AND j.job_key='{args.key}:'||w.code
 LEFT JOIN decision_runs d ON d.job_id=j.id
) t;""")
control = sql('SELECT row_to_json(t) FROM (SELECT paused,execution_enabled,live_acceptance FROM runtime_control) t;')
research = sql(f"SELECT row_to_json(t) FROM (SELECT state,last_error,result FROM jobs WHERE kind='research' AND job_key='{args.key}') t;")
counts = sql("SELECT json_build_object('wallets',(SELECT count(*) FROM wallets),'fills',(SELECT count(*) FROM fills),'orders',(SELECT count(*) FROM orders),'utc',now());")
print(json.dumps({'service': dict(line.split('=', 1) for line in service.splitlines()),
                  'control': control, 'research': research, 'counts': counts,
                  'wallets': wallets}, ensure_ascii=False))
