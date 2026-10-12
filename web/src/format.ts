// Server money stays decimal strings. Format cents using integers; never sum
// account values or revalue positions with floating-point in the browser.
export function money(value:string|null|undefined){
 if(value==null)return '—';const match=/^(-?)(\d+)(?:\.(\d+))?$/.exec(value);if(!match)return '—';
 let cents=BigInt(match[2])*100n+BigInt((match[3]||'').padEnd(2,'0').slice(0,2));if((match[3]||'').length>2&&Number(match[3]![2])>=5)cents++;
 const major=(cents/100n).toLocaleString('en-US');return `${match[1]}$${major}.${(cents%100n).toString().padStart(2,'0')}`;
}
export function escape(value:unknown){return String(value??'').replace(/[&<>"']/g,x=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[x]!))}
export function date(value:string|null|undefined){if(!value)return '—';const d=new Date(value);return Number.isNaN(d.valueOf())?'—':new Intl.DateTimeFormat('tr-TR',{dateStyle:'short',timeStyle:'medium',timeZone:'Europe/Istanbul'}).format(d)}
export const states:Record<string,string>={accepted:'Kabul edildi',waiting:'Bekliyor',partially_filled:'Kısmi gerçekleşti',filled:'Gerçekleşti',cancelled:'İptal edildi',expired:'Süresi doldu',rejected:'Reddedildi',running:'Çalışıyor',completed:'Tamamlandı',failed:'Başarısız',pending:'Planlandı',ready:'Karar hazır'};
export const personas:Record<string,string>={'US-FAST':'Trend / momentum','US-EVENT':'Olay odaklı','US-PATIENT':'Orta / uzun vadeli','QQQ-FAST':'Hızlı fırsat arayışı','QQQ-PATIENT':'Uzun vadeli yatırımcı','TQQQ-FAST':'Hızlı fırsat arayışı','TQQQ-PATIENT':'Uzun vadeli yatırımcı','EWC-BALANCED':'Dengeli yatırımcı'};
