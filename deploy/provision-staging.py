#!/usr/bin/env python3
"""Run only on the authorized Hetzner host. Creates secrets without printing them."""
import pathlib, secrets, os
root=pathlib.Path('/etc/trading/staging');root.mkdir(parents=True,exist_ok=True);root.chmod(0o700)
if not (root/'db.env').exists():
    values={k:secrets.token_hex(32) for k in ('POSTGRES_PASSWORD','OWNER_PASSWORD','AGENT_PASSWORD','WORKER_PASSWORD','OPERATOR_PASSWORD')}
    values.update(POSTGRES_USER='trading_admin',POSTGRES_DB='trading_staging',POSTGRES_INITDB_ARGS='--auth-host=scram-sha-256 --auth-local=scram-sha-256')
    def write(name,vals):
        p=root/name
        fd=os.open(p,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
        with os.fdopen(fd,'w') as f:
            f.write(''.join(k+'='+v+'\n' for k,v in vals.items()))
    def dsn(role,key,host='db'):
        return f'postgresql://trading_{role}:{values[key]}@{host}:5432/trading_staging?sslmode=disable'
    secret=secrets.token_hex(32);operator=secrets.token_hex(32)
    write('db.env',values)
    write('owner.env',{'DATABASE_OWNER_URL':dsn('owner','OWNER_PASSWORD')})
    write('api.env',{'DATABASE_AGENT_URL':dsn('agent','AGENT_PASSWORD'),'DATABASE_OPERATOR_URL':dsn('operator','OPERATOR_PASSWORD'),'AGENT_SIGNING_SECRET':secret,'OPERATOR_TOKEN':operator})
    write('worker.env',{'DATABASE_WORKER_URL':dsn('worker','WORKER_PASSWORD')})
    write('runner.env',{'DATABASE_WORKER_URL':f"postgresql://trading_worker:{values['WORKER_PASSWORD']}@/trading_staging?host=/var/lib/trading/staging/pgsocket&sslmode=disable",'AGENT_SIGNING_SECRET':secret,'API_URL':'http://127.0.0.1:18880'})
    print('New isolated staging secrets created; no values printed.')
else:
    for name in ('owner.env','api.env','worker.env','runner.env'):
        if not (root/name).exists(): raise SystemExit('Incomplete provision; preserve existing credentials and repair explicitly.')
    print('Existing staging secrets retained.')
