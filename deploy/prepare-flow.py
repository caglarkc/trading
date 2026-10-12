#!/usr/bin/env python3
"""Create a clean engineering-flow DB without deleting staging financial history."""
import pathlib,subprocess,json,os
base=pathlib.Path('/etc/trading/staging');root=pathlib.Path('/etc/trading/flow');root.mkdir(parents=True,exist_ok=True);root.chmod(0o700)
def read(name):return dict(l.split('=',1) for l in (base/name).read_text().splitlines())
def psql(sql,db='trading_staging'):
 return subprocess.check_output(['docker','exec','trading-staging-db-1','sh','-c','export PGPASSWORD="$POSTGRES_PASSWORD"; exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -Atc "$2"','flow',db,sql],text=True)
if psql("SELECT 1 FROM pg_database WHERE datname='trading_flow'").strip()!='1':psql('CREATE DATABASE trading_flow OWNER trading_owner')
psql('REVOKE ALL ON SCHEMA public FROM PUBLIC; GRANT USAGE,CREATE ON SCHEMA public TO trading_owner; GRANT USAGE ON SCHEMA public TO trading_agent,trading_worker,trading_operator;','trading_flow')
for name in ('owner.env','api.env','worker.env','runner.env'):
 vals=read(name);vals={k:v.replace('/trading_staging','/trading_flow') for k,v in vals.items()}
 if name=='runner.env':vals.update(COLLECTOR_URL='http://127.0.0.1:18882',API_URL='http://127.0.0.1:18880')
 p=root/name
 if not p.exists():
  fd=os.open(p,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
  with os.fdopen(fd,'w') as f:f.write(''.join(k+'='+v+'\n' for k,v in vals.items()))
print(json.dumps({'database':'trading_flow','staging_history':'preserved','secrets':'not_printed'}))
