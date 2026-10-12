import { defineConfig } from 'vite';
import { readFileSync } from 'node:fs';
import { homedir } from 'node:os';
import http from 'node:http';

export default defineConfig({
 server:{host:'127.0.0.1',port:5173,strictPort:true},
 plugins:[{name:'read-only-operator-proxy',configureServer(server){
  const file=process.env.TRADING_OPERATOR_FILE||`${homedir()}/.config/trading/operator.env`;
  let token='';try{token=readFileSync(file,'utf8').split('\n').find(x=>x.startsWith('OPERATOR_TOKEN='))?.slice(15).trim()||''}catch{}
  server.middlewares.use('/api', (req,res)=>{
   if(req.method!=='GET'||req.url!=='/observation'){res.statusCode=403;res.end('read_only_route');return}
   const origin=req.headers.origin;
   if(origin&&origin!=='http://127.0.0.1:5173'&&origin!=='http://localhost:5173'){res.statusCode=403;res.end('invalid_origin');return}
   res.setHeader('Content-Type','application/json');res.setHeader('Cache-Control','no-store');
   if(!token){res.statusCode=503;res.end('{"error":"operator_config_missing"}');return}
   const upstream=http.request({hostname:'127.0.0.1',port:18883,path:'/operator/observation',method:'GET',headers:{Authorization:`Bearer ${token}`},timeout:10000},r=>{res.statusCode=r.statusCode||502;r.pipe(res)});
   upstream.on('timeout',()=>upstream.destroy());upstream.on('error',()=>{if(!res.headersSent){res.statusCode=503;res.end('{"error":"tunnel_unavailable"}')}else res.destroy()});
   res.on('close',()=>upstream.destroy());upstream.end();
  });
 }}]
});
