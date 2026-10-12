#!/usr/bin/env python3
"""Trusted host orchestration of auditable 50+20 reviews and a later allocation mandate."""
import argparse,concurrent.futures,datetime,json,os,pathlib,pwd,re,subprocess,urllib.request,uuid,time
from decimal import Decimal
from zoneinfo import ZoneInfo
ROOT=pathlib.Path('/var/lib/trading/flow/reviews')
WALLETS=['US-FAST','US-EVENT','US-PATIENT','QQQ-FAST','QQQ-PATIENT','TQQQ-FAST','TQQQ-PATIENT','EWC-BALANCED']
def sql(query):
 p=subprocess.run(['docker','exec','-i','trading-staging-db-1','sh','-c','export PGPASSWORD="$POSTGRES_PASSWORD"; exec psql -X -At -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d trading_flow'],input=query,text=True,capture_output=True,check=True)
 return p.stdout.strip()
def quoted(v):return "'"+str(v).replace("'","''")+"'"
def save(path,data):
 path.parent.mkdir(parents=True,exist_ok=True);path.parent.chmod(0o700)
 uid=pwd.getpwnam('trading');os.chown(path.parent,uid.pw_uid,uid.pw_gid)
 tmp=path.with_suffix('.tmp');tmp.write_text(json.dumps(data,ensure_ascii=False));tmp.chmod(0o600);os.chown(tmp,uid.pw_uid,uid.pw_gid);tmp.replace(path)
def capture(symbol,name,urls,stock=False,exchange=None):
 errors=[]
 for url in urls:
  try:
   req=urllib.request.Request('http://127.0.0.1:18882/read-source',data=json.dumps({'url':url}).encode(),headers={'Content-Type':'application/json'})
   with urllib.request.urlopen(req,timeout=55) as r:d=json.load(r)
   doc=d['document'];text=doc['text']
   if len(text)<100:raise ValueError('body_too_short')
   if symbol and not re.search(r'(?i)(?<![A-Z0-9])'+re.escape(symbol.replace('.','-'))+r'(?![A-Z0-9])',doc.get('title','').replace('.','-')):raise ValueError('ticker_identity_missing')
   if stock:
    if 'ETF' in doc.get('title','').upper():raise ValueError('not_individual_stock')
    normalized=text.upper().replace('-','.')
    listing=re.search(r'\b'+re.escape(exchange)+r'\s*:\s*'+re.escape(symbol.replace('-','.'))+r'(?![A-Z0-9])',normalized)
    yahoo_listing=exchange=='NASDAQ' and any(x in text[:5000] for x in ['NasdaqGS','NasdaqGM','NasdaqCM']) or exchange=='NYSE' and 'NYSE -' in text[:5000]
    if not listing and not yahoo_listing:raise ValueError('exchange_listing_unverified')
   start=max(0,text.find('Market Cap')-800) if 'Market Cap' in text else 0
   excerpt=text[start:start+8500]
   return {'symbol':symbol,'name':name,'evidence_id':d['id'],'url':doc['url'],'text':excerpt,'observed_at':doc['observed_at']}
  except Exception as exc:errors.append(type(exc).__name__)
 raise RuntimeError('source_capture_failed:'+symbol+':'+','.join(errors))
def jobs(campaign,kind):
 return json.loads(sql("SELECT coalesce(json_agg(t ORDER BY code),'[]') FROM (SELECT w.id::text AS wallet_id,w.code,j.id::text AS job_id,j.state,j.last_error,j.result,j.attempts FROM wallets w LEFT JOIN jobs j ON j.wallet_id=w.id AND j.kind="+quoted(kind)+" AND j.job_key="+quoted(campaign+':')+"||w.code) t;"))
def enqueue(campaign,kind,selection):
 payload=json.dumps({'campaign':campaign,'selection':selection})
 for code in WALLETS:
  jid=str(uuid.uuid5(uuid.NAMESPACE_URL,campaign+':'+kind+':'+code))
  sql("INSERT INTO jobs(id,kind,job_key,wallet_id,payload,available_at,deadline,priority) SELECT "+quoted(jid)+","+quoted(kind)+","+quoted(campaign+':'+code)+",id,"+quoted(payload)+",now(),now()+interval '3 hours',2 FROM wallets WHERE code="+quoted(code)+" ON CONFLICT(kind,job_key) DO NOTHING;")
def verified(campaign,kind,count):
 rows=jobs(campaign,kind)
 if len(rows)!=8 or any(r['state']!='completed' or not r['result'] or r['result'].get('verified') is not True or r['result'].get('evaluated_count')!=count for r in rows):raise RuntimeError('all_eight_reviews_not_verified:'+kind)
 return rows

def main():
 p=argparse.ArgumentParser();p.add_argument('mode',choices=['capture','screen','deep','status','purchase','run']);p.add_argument('--campaign',required=True);p.add_argument('--selection',choices=['inside','outside'],default='outside');args=p.parse_args();uuid.UUID(args.campaign)
 base=ROOT/args.campaign;ROOT.mkdir(mode=0o700,parents=True,exist_ok=True);u=pwd.getpwnam('trading');os.chown(ROOT,u.pw_uid,u.pw_gid)
 if args.mode=='capture':
  (ROOT/'research-only.lock').touch(mode=0o600)
  assets=[a for a in json.loads(pathlib.Path('/opt/trading/staging-src/config/assets.json').read_text()) if a['category']=='US'];assert len(assets)==50
  with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
   futures=[pool.submit(capture,a['symbol'],a['name'],[a['aliases']['stockanalysis'],a['aliases']['yahoo']]) for a in assets]
   docs=[]
   for f in concurrent.futures.as_completed(futures):
    d=f.result();docs.append(d);print(json.dumps({'captured':d['symbol'],'count':len(docs),'total':50}),flush=True)
  market=[]
  for sym,name,url in [('QQQ','Nasdaq-100','https://finance.yahoo.com/quote/QQQ/'),('SPY','S&P 500','https://finance.yahoo.com/quote/SPY/'),('IWM','Russell 2000','https://finance.yahoo.com/quote/IWM/'),('DIA','Dow Jones','https://finance.yahoo.com/quote/DIA/'),('EWC','Canada','https://finance.yahoo.com/quote/EWC/'),('','Fed macro outlook','https://www.federalreserve.gov/newsevents/pressreleases/monetary20260916a.htm')]:
   try:market.append(capture(sym,name,[url]))
   except Exception as e:print(json.dumps({'market_source_gap':name,'reason':str(e)}),flush=True)
  save(base/'universe.json',{'documents':sorted(docs,key=lambda x:x['symbol']),'market':market})
  print(json.dumps({'captured_stocks':len(docs),'market_sources':len(market),'campaign':args.campaign}),flush=True)
 elif args.mode=='screen':
  data=json.loads((base/'universe.json').read_text());assert len(data['documents'])==50
  save(base/'campaign.json',{'id':args.campaign,'selection':args.selection,'phase':'screening','started_at':datetime.datetime.now(datetime.timezone.utc).isoformat()})
  enqueue(args.campaign,'wallet_screen',args.selection);print('Eight screen jobs enqueued; buying instruction NOT issued.')
 elif args.mode=='deep':
  rows=verified(args.campaign,'wallet_screen',50);data=json.loads((base/'universe.json').read_text());fixed={d['symbol']:d for d in data['documents']}
  for r in rows:
   file=base/(r['wallet_id']+'-deep.json')
   if not file.exists():
    docs=[]
    for c in r['result']['report']['watchlist']:
     symbol=c['symbol'];assert re.fullmatch(r'[A-Z][A-Z0-9.-]{0,9}',symbol)
     d=fixed.get(symbol)
     if not d:d=capture(symbol,c['name'],['https://stockanalysis.com/stocks/'+symbol.lower().replace('.','-')+'/','https://finance.yahoo.com/quote/'+symbol.replace('.','-')+'/'],stock=True,exchange=c['exchange'])
     docs.append(d)
     print(json.dumps({'wallet':r['code'],'deep_capture':symbol,'count':len(docs)}),flush=True)
    assert len(docs)==20;save(file,{'documents':docs,'market':data['market']})
   print(json.dumps({'wallet':r['code'],'deep_dossier_ready':True}),flush=True)
  enqueue(args.campaign,'wallet_deep',args.selection);print('Eight detailed 20-stock jobs enqueued; buying instruction NOT issued.')
 elif args.mode=='status':
  print(json.dumps({'campaign':args.campaign,'screen':[{k:r[k] for k in ['code','state','last_error','attempts']} for r in jobs(args.campaign,'wallet_screen')],'deep':[{k:r[k] for k in ['code','state','last_error','attempts']} for r in jobs(args.campaign,'wallet_deep')]},ensure_ascii=False))
 elif args.mode=='purchase':
  screens=verified(args.campaign,'wallet_screen',50);deep=verified(args.campaign,'wallet_deep',20)
  for r in screens:
   dataset=json.loads((base/(r['wallet_id']+'-deep.json')).read_text());docs={d['symbol']:d for d in dataset['documents']}
   statements=['BEGIN;', 'UPDATE wallet_watchlist_members SET active=false WHERE wallet_id='+quoted(r['wallet_id'])+';']
   for candidate in r['result']['report']['watchlist']:
    symbol=candidate['symbol'];document=docs[symbol]
    aid=str(uuid.uuid5(uuid.NAMESPACE_URL,'trading:asset:'+symbol))
    # Use the Go UUID v5-equivalent SHA1 namespace function used by asset seed.
    aliases={'stockanalysis':'https://stockanalysis.com/stocks/'+symbol.lower().replace('.','-')+'/', 'yahoo':'https://finance.yahoo.com/quote/'+symbol.replace('.','-')+'/', 'webull':'https://www.webull.com/quote/'+candidate['exchange'].lower()+'-'+symbol.lower().replace('.','-')}
    statements.append('INSERT INTO assets(id,symbol,name,category,exchange,aliases,fixed) VALUES('+quoted(aid)+','+quoted(symbol)+','+quoted(candidate['name'])+",'US',"+quoted(candidate['exchange'])+','+quoted(json.dumps(aliases))+',false) ON CONFLICT(id) DO NOTHING;')
    deepjob=next(x['job_id'] for x in deep if x['wallet_id']==r['wallet_id'])
    statements.append('INSERT INTO wallet_watchlist_members(wallet_id,asset_id,review_job_id,active) VALUES('+quoted(r['wallet_id'])+','+quoted(aid)+','+quoted(deepjob)+',true) ON CONFLICT(wallet_id,asset_id) DO UPDATE SET review_job_id=excluded.review_job_id,active=true;')
   statements.append('COMMIT;');sql('\n'.join(statements))
  cal=json.loads(pathlib.Path('/opt/trading/staging-src/config/calendar.json').read_text());now=datetime.datetime.now(datetime.timezone.utc);et=ZoneInfo('America/New_York');available=now
  local=now.astimezone(et)
  for offset in range(15):
   day=local.date()+datetime.timedelta(days=offset)
   if day.weekday()>=5 or day.isoformat() in cal['market_holidays']:continue
   opened=datetime.datetime.combine(day,datetime.time(9,30),et);closed=datetime.datetime.combine(day,datetime.time(13 if day.isoformat() in cal['early_close'] else 16),et)
   if now>=closed-datetime.timedelta(minutes=15):continue
   available=max(now,opened+datetime.timedelta(seconds=30)).astimezone(datetime.timezone.utc);break
  else:raise RuntimeError('next_market_session_missing')
  payload=json.dumps({'campaign':args.campaign,'mandatory_allocation':True})
  for code in WALLETS:
   jid=str(uuid.uuid5(uuid.NAMESPACE_URL,args.campaign+':purchase:'+code))
   sql("INSERT INTO jobs(id,kind,job_key,wallet_id,payload,available_at,deadline,priority) SELECT "+quoted(jid)+",'wallet_allocation',"+quoted(args.campaign+':purchase:'+code)+",id,"+quoted(payload)+','+quoted(available.isoformat())+'::timestamptz,'+quoted(available.isoformat())+"::timestamptz+interval '60 minutes',1 FROM wallets WHERE code="+quoted(code)+" ON CONFLICT(kind,job_key) DO NOTHING;")
  print('Second instruction issued to all eight only after 50+20 reviews verified.')
 elif args.mode=='run':
  last=None
  def run_stage(mode):
   subprocess.run(['python3',__file__,mode,'--campaign',args.campaign,'--selection',args.selection],check=True)
  while True:
   screens=jobs(args.campaign,'wallet_screen');deep=jobs(args.campaign,'wallet_deep')
   counts={'screen_verified':sum(r['state']=='completed' for r in screens),'deep_verified':sum(r['state']=='completed' for r in deep)}
   if counts!=last:print(json.dumps(counts),flush=True);last=counts
   for kind,rows in [('wallet_screen',screens),('wallet_deep',deep)]:
    for r in rows:
     if r['state']=='failed' and r['attempts']<3:
      sql('UPDATE jobs SET state=\'pending\',last_error=NULL,available_at=now(),deadline=now()+interval \'2 hours\' WHERE id='+quoted(r['job_id'])+' AND state=\'failed\';')
   if counts['screen_verified']==8 and not any(r['job_id'] for r in deep):
    try:run_stage('deep')
    except Exception as exc:print(json.dumps({'stage':'deep_capture','status':'retrying','reason':type(exc).__name__}),flush=True);time.sleep(30)
   elif counts['deep_verified']==8:
    allocated=sql('SELECT count(*) FROM jobs WHERE kind=\'wallet_allocation\' AND payload->>\'campaign\'='+quoted(args.campaign)+';')
    if allocated=='0':run_stage('purchase')
    balances=json.loads(sql("SELECT coalesce(json_agg(t ORDER BY code),'[]') FROM (SELECT w.code,(state->>'available_cash_usd')::numeric AS cash,(state->>'reserved_cash_usd')::numeric AS reserved,(state->>'unsettled_cash_usd')::numeric AS unsettled FROM wallets w) t;"))
    reached=sum(Decimal(str(w['cash']))+Decimal(str(w['reserved']))+Decimal(str(w['unsettled']))<=20000 for w in balances)
    if reached==8:
     (ROOT/'research-only.lock').unlink(missing_ok=True);save(base/'completed.json',{'phase':'complete','balances':balances,'verified_at':datetime.datetime.now(datetime.timezone.utc).isoformat()});print(json.dumps({'allocation_reached':8,'status':'complete','balances':balances}),flush=True);return
    for w in balances:
     if Decimal(str(w['cash']))+Decimal(str(w['reserved']))+Decimal(str(w['unsettled']))<=20000:continue
     detail=json.loads(sql("SELECT json_build_object('wallet_id',id::text,'open_orders',(SELECT count(*) FROM orders WHERE wallet_id=wallets.id AND state IN ('accepted','waiting','partially_filled')),'last_state',(SELECT state FROM jobs WHERE wallet_id=wallets.id AND kind='wallet_allocation' AND payload->>'campaign'="+quoted(args.campaign)+" ORDER BY created_at DESC LIMIT 1),'allocation_jobs',(SELECT count(*) FROM jobs WHERE wallet_id=wallets.id AND kind='wallet_allocation' AND payload->>'campaign'="+quoted(args.campaign)+")) FROM wallets WHERE code="+quoted(w['code'])+";"))
     if (detail['open_orders'] and Decimal(str(w['cash']))<=20000) or detail['last_state'] in ['pending','running'] or detail['allocation_jobs']>=4:continue
     cal=json.loads(pathlib.Path('/opt/trading/staging-src/config/calendar.json').read_text());now=datetime.datetime.now(datetime.timezone.utc);et=ZoneInfo('America/New_York');local=now.astimezone(et)
     for offset in range(15):
      day=local.date()+datetime.timedelta(days=offset)
      if day.weekday()>=5 or day.isoformat() in cal['market_holidays']:continue
      opened=datetime.datetime.combine(day,datetime.time(9,30),et);closed=datetime.datetime.combine(day,datetime.time(13 if day.isoformat() in cal['early_close'] else 16),et)
      if now>=closed-datetime.timedelta(minutes=15):continue
      available=max(now,opened+datetime.timedelta(seconds=30)).astimezone(datetime.timezone.utc);break
     else:raise RuntimeError('next_market_session_missing')
     key=args.campaign+':purchase:retry-'+str(detail['allocation_jobs'])+':'+w['code'];jid=str(uuid.uuid5(uuid.NAMESPACE_URL,key));payload=json.dumps({'campaign':args.campaign,'mandatory_allocation':True})
     sql('INSERT INTO jobs(id,kind,job_key,wallet_id,payload,available_at,deadline,priority) VALUES('+quoted(jid)+",'wallet_allocation',"+quoted(key)+','+quoted(detail['wallet_id'])+','+quoted(payload)+','+quoted(available.isoformat())+'::timestamptz,'+quoted(available.isoformat())+"::timestamptz+interval '60 minutes',1) ON CONFLICT(kind,job_key) DO NOTHING;")
     print(json.dumps({'allocation_retry':w['code'],'attempt':detail['allocation_jobs']+1,'available_at':available.isoformat()}),flush=True)
    print(json.dumps({'allocation_reached':reached,'balances':balances}),flush=True)
   time.sleep(15)
if __name__=='__main__':main()
