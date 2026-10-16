#!/usr/bin/env python3
"""Read wallet strategy versions and closed-fill review coverage on the host."""
import argparse
import json
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--full', action='store_true', help='include full recorded reviews and rules')
args = parser.parse_args()
query = """SELECT coalesce(json_agg(t ORDER BY code),'[]') FROM (
 SELECT w.code,s.version,s.created_at,s.reviewed_through_seq,
 coalesce(jsonb_array_length(s.report->'trade_reviews'),0) AS latest_review_count,
 (SELECT count(*) FROM fills f JOIN orders o ON o.wallet_id=f.wallet_id AND o.id=f.order_id
  WHERE f.wallet_id=w.id AND o.body->>'side'='SELL') AS total_closed_portions,
 (SELECT count(*) FROM fills f JOIN orders o ON o.wallet_id=f.wallet_id AND o.id=f.order_id
  WHERE f.wallet_id=w.id AND o.body->>'side'='SELL' AND f.seq>coalesce(s.reviewed_through_seq,0)) AS pending_closed_portions,
 s.report, (SELECT row_to_json(j) FROM (SELECT state,last_error,finished_at FROM jobs
  WHERE wallet_id=w.id AND kind='wallet_learning' ORDER BY created_at DESC LIMIT 1) j) AS latest_job
 FROM wallets w LEFT JOIN LATERAL (SELECT * FROM strategy_versions WHERE wallet_id=w.id ORDER BY version DESC LIMIT 1) s ON true
) t;"""
p = subprocess.run(
    ['docker', 'exec', '-i', 'trading-staging-db-1', 'sh', '-c',
     'export PGPASSWORD="$POSTGRES_PASSWORD"; exec psql -X -At -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d trading_flow'],
    input=query, text=True, capture_output=True, check=True,
)
wallets = json.loads(p.stdout)
if not args.full:
    for w in wallets:
        report = w.pop('report') or {}
        w['strategy_summary'] = report.get('strategy_summary')
        w['rules'] = [{'id': r['id'], 'status': r['status'], 'current_rule': r['current_rule']}
                      for r in report.get('rules', [])]
print(json.dumps({'wallets': wallets}, ensure_ascii=False, indent=2))
