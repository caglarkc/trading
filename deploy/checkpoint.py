#!/usr/bin/env python3
"""Trusted host-only encrypted DB/session checkpoint. No model credentials included."""
import argparse,datetime,fcntl,hashlib,json,os,pathlib,pwd,re,shutil,subprocess,tarfile,tempfile,time
DB_CONTAINER='trading-staging-db-1'
AGE='/opt/trading/bin/age'
NAME=re.compile(r'^trading_[a-z0-9_]{1,45}$')

def sql(db,query):
 if not NAME.fullmatch(db):raise ValueError('invalid_database')
 return subprocess.check_output(['docker','exec','-i',DB_CONTAINER,'sh','-c','export PGPASSWORD="$POSTGRES_PASSWORD"; exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -At','checkpoint',db],input=query.encode(),stderr=subprocess.PIPE).decode().strip()
def sha(path):
 h=hashlib.sha256()
 with open(path,'rb') as f:
  for b in iter(lambda:f.read(1024*1024),b''):h.update(b)
 return h.hexdigest()
def fingerprint(db):
 # The dump also contains all other tables. These financial/session fingerprints
 # are compared after restore; job leases and runtime pause state are excluded.
 out={}
 for name,query in {
  'wallets':"SELECT md5(coalesce(string_agg(id::text||code||state::text||account_version::text,'' ORDER BY id),'')) FROM wallets",
  'journal':"SELECT md5(coalesce(string_agg(wallet_id::text||transaction_id::text||account_code||amount::text,'' ORDER BY id),'')) FROM journal_lines",
  'sessions':"SELECT md5(coalesce(string_agg(wallet_id::text||coalesce(session_id,''),'' ORDER BY wallet_id),'')) FROM wallet_sessions",
  'memory':"SELECT md5(coalesce(string_agg(wallet_id::text||version::text||summary,'' ORDER BY wallet_id,version),'')) FROM memory_versions",
  'orders':"SELECT md5(coalesce(string_agg(id::text||wallet_id::text||body::text,'' ORDER BY id),'')) FROM orders",
  'fills':"SELECT md5(coalesce(string_agg(id::text||wallet_id::text||body::text,'' ORDER BY id),'')) FROM fills",
  'receipts':"SELECT md5(coalesce(string_agg(id::text||wallet_id::text||body::text||coalesce(delivered_at::text,''),'' ORDER BY id),'')) FROM receipts"
 }.items():out[name]=sql(db,query)
 return out

def checkpoint_lock(root,limit=900):
 root=pathlib.Path(root);root.mkdir(parents=True,exist_ok=True)
 file=root/'checkpoint.lock'
 fd=os.open(file,os.O_CREAT|os.O_RDWR,0o600);uid=pwd.getpwnam('trading');os.fchown(fd,uid.pw_uid,uid.pw_gid)
 started=time.monotonic()
 while True:
  try:fcntl.flock(fd,fcntl.LOCK_EX|fcntl.LOCK_NB);return fd
  except BlockingIOError:
   if time.monotonic()-started>limit:os.close(fd);raise RuntimeError('checkpoint_writer_timeout')
   time.sleep(.2)

def backup(args):
 root=pathlib.Path(args.root);out=pathlib.Path(args.out);out.mkdir(parents=True,exist_ok=True);out.chmod(0o700)
 if shutil.disk_usage(out).free<2*1024**3:raise RuntimeError('disk_space_critical')
 was_paused=sql(args.db,'SELECT paused FROM runtime_control WHERE id')=='t'
 sql(args.db,'UPDATE runtime_control SET paused=true,updated_at=now() WHERE id')
 lock=None
 try:
  start=time.monotonic()
  while sql(args.db,"SELECT count(*) FROM jobs WHERE state='running' AND lease_until>now()")!='0':
   if time.monotonic()-start>900:raise RuntimeError('checkpoint_job_drain_timeout')
   time.sleep(1)
  lock=checkpoint_lock(root)
  if list(root.glob('*/codex-home/auth.json')):raise RuntimeError('active_or_orphan_auth_requires_reconciliation')
  stamp=datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%SZ')
  final=out/('daily-'+stamp+'.age')
  with tempfile.TemporaryDirectory(prefix='.checkpoint-',dir=out) as temp:
   temp=pathlib.Path(temp);dump=temp/'database.dump'
   with open(dump,'xb') as f:
    subprocess.run(['docker','exec',DB_CONTAINER,'sh','-c','export PGPASSWORD="$POSTGRES_PASSWORD"; exec pg_dump -Fc -U "$POSTGRES_USER" -d "$1"','checkpoint',args.db],stdout=f,stderr=subprocess.PIPE,check=True)
   manifest={'version':'checkpoint-v1','environment':args.db,'created_at':stamp,'release':args.release,'schema':sql(args.db,'SELECT max(version_id) FROM goose_db_version WHERE is_applied'),'financial_fingerprint':fingerprint(args.db),'files':{},'auth':'excluded; use current canonical login after restore','sessions':json.loads(sql(args.db,"SELECT coalesce(json_object_agg(wallet_id::text,session_id),'{}'::json) FROM wallet_sessions"))}
   bundle=temp/'checkpoint.tar.gz'
   with tarfile.open(bundle,'w:gz') as tar:
    def add(path,name):
     if path.is_symlink() or not path.is_file():raise RuntimeError('checkpoint_nonregular_file')
     manifest['files'][name]=sha(path);tar.add(path,arcname=name,recursive=False)
    add(dump,'database.dump')
    for p in sorted(root.rglob('*')):
     if p.name in ('auth.json','bridge.sock','runner.lock','checkpoint.lock') or '.atomic-' in p.name:continue
     if p.is_symlink():raise RuntimeError('checkpoint_session_symlink')
     if p.is_file():add(p,'sessions/'+p.relative_to(root).as_posix())
    for p in sorted(pathlib.Path(args.config).glob('*.json')):add(p,'config/'+p.name)
    meta=temp/'manifest.json';meta.write_text(json.dumps(manifest,sort_keys=True));tar.add(meta,arcname='manifest.json',recursive=False)
   pending=out/('.'+final.name+'.pending')
   try:
    subprocess.run([AGE,'-R',args.recipient,'-o',str(pending),str(bundle)],stderr=subprocess.PIPE,check=True)
    pending.chmod(0o600)
    with open(pending,'rb') as f:os.fsync(f.fileno())
    os.rename(pending,final)
   finally:
    if pending.exists():pending.unlink()
  # Seven daily and four weekly encrypted checkpoints, only this namespace.
  daily=sorted(out.glob('daily-????????T??????Z.age'),reverse=True)
  for p in daily[7:]:p.unlink()
  week=datetime.datetime.now(datetime.timezone.utc).strftime('%G-W%V')
  weekly=out/('weekly-'+week+'.age')
  if not weekly.exists():shutil.copy2(final,weekly)
  for p in sorted(out.glob('weekly-????-W??.age'),reverse=True)[4:]:p.unlink()
  status={'path':str(final),'sha256':sha(final),'bytes':final.stat().st_size,'status':'passed','database':args.db,'schema':manifest['schema'],'auth':'excluded'}
  sql(args.db,"INSERT INTO system_events(kind,body) VALUES('checkpoint_backup','"+json.dumps(status).replace("'","''")+"'::jsonb)")
  print(json.dumps(status))
 finally:
  if lock is not None:fcntl.flock(lock,fcntl.LOCK_UN);os.close(lock)
  if not was_paused:sql(args.db,'UPDATE runtime_control SET paused=false,updated_at=now() WHERE id')

def verify_bundle(bundle,dest):
 # Never use extractall on a restore archive. Refuse links and traversal.
 total=0;seen=set()
 with tarfile.open(bundle,'r:gz') as tar:
  for member in tar:
   name=member.name
   if name in seen or pathlib.PurePosixPath(name).is_absolute() or '..' in pathlib.PurePosixPath(name).parts:raise RuntimeError('unsafe_archive_path')
   seen.add(name)
   if not member.isfile() or not (name in ('database.dump','manifest.json') or name.startswith(('sessions/','config/'))):raise RuntimeError('unsafe_archive_member')
   total+=member.size
   if total>1024**3 or member.size>256*1024**2:raise RuntimeError('archive_size_limit')
   path=dest/name;path.parent.mkdir(parents=True,exist_ok=True);path.parent.chmod(0o700)
   with tar.extractfile(member) as src,open(path,'xb') as dst:shutil.copyfileobj(src,dst)
   path.chmod(0o600)
 manifest=json.loads((dest/'manifest.json').read_text())
 if manifest.get('version')!='checkpoint-v1':raise RuntimeError('unsupported_checkpoint')
 if set(manifest['files'])!=seen-{'manifest.json'}:raise RuntimeError('manifest_file_set_mismatch')
 for name,digest in manifest['files'].items():
  if sha(dest/name)!=digest:raise RuntimeError('checkpoint_hash_mismatch')
 if any(pathlib.PurePosixPath(n).name=='auth.json' for n in seen):raise RuntimeError('auth_restore_forbidden')
 return manifest

def restore(args):
 if not NAME.fullmatch(args.db) or not args.db.startswith('trading_restore_'):raise RuntimeError('restore_requires_new_namespace')
 root=pathlib.Path(args.root)
 if root.exists():raise RuntimeError('restore_session_root_exists')
 # Never silently replace a database or private session root.
 if sql('trading_staging',"SELECT count(*) FROM pg_database WHERE datname='"+args.db+"'")!='0':raise RuntimeError('restore_database_exists')
 root.parent.mkdir(parents=True,exist_ok=True)
 with tempfile.TemporaryDirectory(prefix='.restore-',dir=root.parent) as temp:
  temp=pathlib.Path(temp);manifest=verify_bundle(args.archive,temp)
  if args.release and manifest['release']!=args.release:raise RuntimeError('restore_release_mismatch')
  sql('trading_staging','CREATE DATABASE '+args.db+' OWNER trading_owner')
  with open(temp/'database.dump','rb') as f:
   subprocess.run(['docker','exec','-i',DB_CONTAINER,'sh','-c','export PGPASSWORD="$POSTGRES_PASSWORD"; exec pg_restore --exit-on-error --no-owner --role=trading_owner -U "$POSTGRES_USER" -d "$1"','checkpoint',args.db],stdin=f,stderr=subprocess.PIPE,check=True)
  if fingerprint(args.db)!=manifest['financial_fingerprint']:raise RuntimeError('restore_financial_fingerprint_mismatch')
  if sql(args.db,'SELECT max(version_id) FROM goose_db_version WHERE is_applied')!=manifest['schema']:raise RuntimeError('restore_schema_mismatch')
  sql(args.db,"UPDATE runtime_control SET paused=true,execution_enabled=false,live_acceptance='{}',updated_at=now();UPDATE jobs SET state='expired',last_error='restore_requires_new_session_job',fencing_token=fencing_token+1,lease_until=NULL WHERE state IN ('pending','running');")
  shutil.move(temp/'sessions',root)
  uid=pwd.getpwnam('trading')
  for p in [root,*root.rglob('*')]:os.chown(p,uid.pw_uid,uid.pw_gid);p.chmod(0o700 if p.is_dir() else 0o600)
  config=root.parent/(root.name+'-config');shutil.move(temp/'config',config);config.chmod(0o700)
  print(json.dumps({'status':'passed','database':args.db,'session_root':str(root),'schema':manifest['schema'],'financial_session_fingerprints':'matched','session_links':len(manifest['sessions']),'execution_enabled':False,'paused':True,'auth':'current canonical login required'}))

if __name__=='__main__':
 parser=argparse.ArgumentParser();parser.add_argument('mode',choices=['backup','restore']);parser.add_argument('--db',required=True);parser.add_argument('--root',required=True);parser.add_argument('--out',default='/var/lib/trading/backups/flow');parser.add_argument('--config',default='/opt/trading/staging-src/config');parser.add_argument('--recipient',default='/etc/trading/backup-recipient.txt');parser.add_argument('--release',required=True);parser.add_argument('--archive')
 args=parser.parse_args()
 try:
  if args.mode=='backup':backup(args)
  else:restore(args)
 except Exception as exc:
  # Subprocess error objects may include confidential command arguments: report
  # only the error category, never environment/stderr or credential values.
  print(json.dumps({'status':'failed','reason':str(exc) if isinstance(exc,(RuntimeError,ValueError)) else type(exc).__name__}));raise SystemExit(1)
