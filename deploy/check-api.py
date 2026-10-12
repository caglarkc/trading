#!/usr/bin/env python3
import pathlib,json,urllib.request,urllib.error,uuid,time,hmac,hashlib,base64
cfg=dict(l.split('=',1) for l in pathlib.Path('/etc/trading/staging/api.env').read_text().splitlines())
def request(path,token=None,method='GET'):
    req=urllib.request.Request('http://127.0.0.1:18880'+path,method=method)
    if token:req.add_header('Authorization','Bearer '+token)
    try:
        with urllib.request.urlopen(req,timeout=5) as r:return r.status,json.load(r)
    except urllib.error.HTTPError as e:return e.code,None
assert request('/operator/wallets')[0]==401
status,rows=request('/operator/wallets',cfg['OPERATOR_TOKEN']);assert status==200 and len(rows)==8
assert sum(float(r['account']['available_cash_usd']) for r in rows)==800000
assert request('/operator/wallets',cfg['OPERATOR_TOKEN'],'POST')[0]==405
wallet=rows[0]['id'];claims={'wallet':wallet,'decision':str(uuid.uuid4()),'job':str(uuid.uuid4()),'exp':int(time.time())+120}
enc=lambda b:base64.urlsafe_b64encode(b).decode().rstrip('=')
part=enc(json.dumps(claims,separators=(',',':')).encode());token=part+'.'+enc(hmac.new(cfg['AGENT_SIGNING_SECRET'].encode(),part.encode(),hashlib.sha256).digest())
status,own=request('/agent/account?wallet_id='+rows[1]['id'],token);assert status==200 and own['id']==wallet
assert request('/operator/wallets',token)[0]==401
print(json.dumps({'operator_wallets':8,'capital_usd':'800000','unauthenticated':'denied','operator_write':'denied','wallet_query_override':'ignored','agent_operator_access':'denied'}))
