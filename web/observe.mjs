import { spawn } from 'node:child_process';
import { createServer } from 'vite';
let stopping=false,ssh,timer;
function connect(){
 ssh=spawn('ssh',['-N','-o','BatchMode=yes','-o','StrictHostKeyChecking=yes','-o','ExitOnForwardFailure=yes','-o','ServerAliveInterval=15','-o','ServerAliveCountMax=2','-L','127.0.0.1:18883:127.0.0.1:18880','root@167.235.158.118'],{stdio:['ignore','ignore','inherit']});
 ssh.on('exit',()=>{if(!stopping)timer=setTimeout(connect,10000)});
}
connect();const server=await createServer();await server.listen();server.printUrls();
async function stop(){if(stopping)return;stopping=true;clearTimeout(timer);ssh?.kill('SIGTERM');await server.close();process.exit(0)}
process.on('SIGINT',stop);process.on('SIGTERM',stop);
