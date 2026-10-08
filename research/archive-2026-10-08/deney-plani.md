# Trading Araştırması — Güncel Kararlar ve Deney Taslağı

Güncelleme: 8 Ekim 2026. **Adım adım planlama sürüyor; ürün kararları, kendi Go backend'imiz ve basit web gözlem ekranı netleşti; teknik mimari önerisi bölüm 0.17'dedir. Kodlama başlamadı.** Uygulamada kaynak erişimi henüz sınanmadı. Aşağıdaki güncel kullanıcı kararları önceki plan varsayımlarının önüne geçer.

Güncel kesinleşmiş kararların tek kaynağı bölüm 0'dır. Bölüm 1–9, 2 Ekim tarihli P1 sayısal araştırma taslağını korur; yeni canlı pilot için onaylanmış veya kilitlenmiş deney değildir. [MVP planı](../MVP-TEKNIK-PLAN.md) bu kayda yönlendirir; [matematik raporu](matematik-algoritmalari.md) ve [karar raporu](../KAPSAMLI-ARASTIRMA-RAPORU.md) önceki araştırma gerekçelerini taşır. Tüm planlama maddeleri dört kategori için ayrı ayrı netleştirilecek, ardından belgeler birlikte revize edilip kodlamaya geçilecektir.

## 0. Güncel kullanıcı kararları — canlı pilot

### 0.1. Kapsam ve erişim

- Dört kategori: ABD hisseleri, QQQ, TQQQ ve EWC. Her kategori ayrı değerlendirilecek; QQQ ve TQQQ ortak Nasdaq-100 maruziyeti nedeniyle bağımsız piyasa deneyleri sayılmayacak.
- Ana model erişimi GPT aboneliği üzerinden olacak. İlk teknik kurulum GPT abonelik bağlantısını çalıştırmak; Gemini ve model API seçenekleri daha sonra değerlendirilecek. Claude zorunlu değil, OpenRouter şimdilik kapsam dışı.
- Fiyat toplamada ek veri aboneliği veya veri sağlayıcısı API anahtarı kullanılmayacak. Yöntem modelin canlı web araması ve doğrudan ürün sayfası okumasıdır; arama önbelleği ve kısa sonuç metni tek başına fiyat kanıtı değildir. Abonelik kotaları içinde çalışılır.
- Kodlama, tüm planlama kararları ve son genel revizyon tamamlandıktan sonra başlayacak. Bu belge bağlantı veya canlı fiyat doğrulamasının yapıldığını iddia etmez.

### 0.2. Fiyat kaynakları — kesinleşmiş liste

Her alım/satım adayı için üç farklı siteden uygun fiyat kaydı zorunludur. İki kaynağın uyuşması üç kaynak şartının yerine geçmez.

| Kategori | Kaynak 1 | Kaynak 2 | Kaynak 3 | Karşılaştırılacak ürün |
|---|---|---|---|---|
| ABD hisseleri | Yahoo Finance: hissenin ürün sayfası | Nasdaq: aynı hissenin ürün sayfası | Stock Analysis: aynı hissenin ürün sayfası | Sabit varlık kimliği, hisse sınıfı ve geçerli sembolü eşleşen ABD hissesi |
| QQQ | [Yahoo Finance](https://finance.yahoo.com/quote/QQQ/) | [Nasdaq](https://www.nasdaq.com/market-activity/etf/qqq) | [Stock Analysis](https://stockanalysis.com/etf/qqq/) | USD cinsinden QQQ; Nasdaq endeks değeri veya NAV değil |
| TQQQ | [Yahoo Finance](https://finance.yahoo.com/quote/TQQQ/) | [Nasdaq](https://www.nasdaq.com/market-activity/etf/tqqq) | [Stock Analysis](https://stockanalysis.com/etf/tqqq/) | USD cinsinden TQQQ'nun kendi fiyatı; QQQ getirisinin üç katı değil |
| EWC | [Yahoo Finance](https://finance.yahoo.com/quote/EWC/) | [Nasdaq](https://www.nasdaq.com/market-activity/etf/ewc) | [Stock Analysis](https://stockanalysis.com/etf/ewc/) | NYSE Arca'da işlem gören USD cinsinden EWC |

Üç farklı site, üç bağımsız veri üreticisi anlamına gelmez. Bilinebilen feed/sağlayıcı adı kayda girer; ortak sağlayıcı bağımsızlık kanıtı sayılmaz. Kaynak listesi seçilmiştir, çalışma zamanı erişimi henüz doğrulanmamıştır: 4 Ekim kontrolünde Nasdaq fiyat alanı okunamadı ve Yahoo doğrudan sayfa erişimi başarısızdı. Kurulumda piyasa açıkken gerçek erişim sınanacak; erişilemeyen kaynak uydurulmuş fiyatla veya sessizce iki kaynakla tamamlanmayacak. Kaynak değişikliği yeni plan kararı olarak kaydedilecek.

### 0.3. Takvim ve zaman damgaları — kesinleşti

- ABD hisseleri, QQQ, TQQQ ve EWC için ilgili ABD borsa takvimi kullanılır; zaman dönüşümünde America/New_York esas alınır. EWC Kanada hisselerine maruziyet sağlar, Çin ürünü değildir; ABD'de işlem görür.
- Kullanıcıya gösterilen saatler GMT+3 / Europe/Istanbul'dur. Normal 09:30–16:00 ET seansı, ABD yaz saatinde 16:30–23:00; standart saatte 17:30–ertesi gün 00:00 İstanbul saatidir. ABD yaz saati geçişi, tatiller ve erken kapanışlar otomatik hesaba katılır; sabit yıl boyu saat kaydırması kullanılmaz.
- Saklamada UTC ve kaynağın özgün saat dilimi korunur; sunum GMT+3'tür. Fiyatın ait olduğu zaman ile sayfanın açıldığı/toplandığı zaman ayrı alanlardır.
- Fiyat kayıtlarının en yenisi ile en eskisi arasındaki fark en fazla **60 saniye** olmalıdır.
- Piyasa açıkken kontrol tamamlandığı anda her fiyat kaydı en fazla **60 saniye eski** olmalıdır. Üç eski kaydın kendi aralarında uyuşması güncellik şartını sağlamaz. Geleceğe ait veya zamanının uygunluğu kanıtlanamayan kayıt geçersizdir.
- Normal seans, seans öncesi ve seans sonrası fiyatlar birbirine karıştırılmaz. Piyasa kapalıyken son normal seans kapanışı aynı tarih/seans için araştırmada kullanılabilir; 60 saniye kuralını geçen canlı fiyat veya yeni işlem gerçekleşmesi gibi etiketlenmez.

### 0.4. Fiyat farkı — kesinleşti: en fazla %1

Kullanıcının son kararı **%1**'dir; daha önce konuşulan %0,01 eşiğinin yerine geçer. Üç sonlu ve pozitif fiyat p1, p2, p3 için:

```text
fark_yuzde = 100 * (max(p1, p2, p3) - min(p1, p2, p3)) / median(p1, p2, p3)
kabul: fark_yuzde <= 1.00
```

Bu, en yüksek ve en düşük fiyat arasındaki toplam açıklığın sınırıdır; ortanca değerin iki yanında ayrı ayrı ±%1 izin verilmez. Tam %1 kabul edilir. Eşik ondalık oran olarak 0.01'dir. Aynı ürün, para birimi, fiyat türü, işlem seansı ve bölüm 0.3 zaman koşulları da sağlanmalıdır.

Son işlem fiyatı yalnız son işlem fiyatıyla karşılaştırılır. Bid/ask, önceki kapanış, NAV ve düzeltilmiş geçmiş fiyat güncel son işlem fiyatının yerine konulmaz. %1 kaynak uyum filtresidir; gerçekleşebilir emir fiyatı, kayma sınırı veya kesin doğruluk garantisi değildir. Referans ve emir gerçekleşme fiyatı seçimi simülatör aşamasında ayrıca kararlaştırılacak; üç fiyatın ortancası otomatik fill fiyatı yapılmayacak.

### 0.5. Kayıt ve hata davranışı

Her kaynak kaydı en az kaynak/site, URL, sembol ve varlık kimliği, borsa, para birimi, fiyat türü, fiyat, fiyatın tarih/saat/saat dilimi, toplama zamanı, seans, gecikme etiketi ve bilinebilen feed/sağlayıcı adını içerir. Doğrulama kaydı üç kaynak kaydına bağlanır; fiyat açıklığı, zaman açıklığı, kontrol zamanı, kural sürümü ve kabul/ret nedenini taşır. Ham içerik saklama kapsamı ayrı veri saklama kararıyla netleşecek; kanıtlanmamış zaman veya sağlayıcı bilgisi uydurulmayacak.

- Üç uygun kayıt yoksa, fark >%1 ise veya zaman koşulu geçmezse durum `price_unverified`; alım/satım gerçekleşmez, emir bekler.
- Çelişkinin olası nedeni yeniden kontrol edilir; farklı seans, gecikme veya sembol yanlışı gizlenmez. Kaynaklardan biri sessizce çıkarılmaz.
- Model kendi bilgisiyle fiyat dolduramaz, sayıları uyuşturmak için değiştiremez veya eşiği otomatik genişletemez. Doğrulama sayı/zaman kuralları uygulama kodunda da denetlenecek.
- Karar öncesine ait fiyat daha sonra gerçekleşmiş işlem diye geçmişe yazılmaz. Fill zamanı, spread, kayma, emir ömrü ve tekrar deneme politikası sonraki simülatör kararıdır.
- Başarılı/başarısız doğrulama oranları, gecikme ve ret nedenleri kategori bazında raporlanacak. Erişim/fiyat sorunu araştırma başarısızlığından ayrılacak; eksik kayıt sıfır fiyat veya nötr sonuç değildir.

### 0.6. Haber ve diğer bilgi kapsamı — kesinleşti

Kullanıcı, uzman değerlendirmesinden yalnız web araması ve toplanacak bilgilerle ilgili önerilerin ele alınmasını istedi; bu kapsam değerlendirmesini kabul ederek kaydedilmesini istedi. Hisse sayısı, model bağlantısı ve mimari önerileri bu kararla değişmedi. Haber ve diğer bilgiler de ek ücretli veri aboneliği veya veri sağlayıcısı API anahtarı olmadan, canlı web araması ve erişilebilir sayfa/belge okumasıyla toplanır.

| Kategori | Toplanacak bilgiler | Öncelikli birincil kaynaklar | Tamamlayıcı arama kaynakları |
|---|---|---|---|
| ABD hisseleri | Şirket duyuruları; 8-K, 10-Q ve 10-K; doğrulanmış bilanço tarihi; bilanço sonuçları ve erişilebilen analist beklentisi; yönetim rehberliği ve değişiklikleri; birleşme/satın alma; yönetim değişikliği; önemli dava/düzenleyici kararlar; ilgili sektör olayları | SEC EDGAR; şirket yatırımcı ilişkileri; DOJ, FTC ve SEC duyuruları | AP, CNBC; Yahoo Finance ve Nasdaq'ın erişilebilir bilgi sayfaları; Business Wire, PR Newswire ve GlobeNewswire'daki özgün şirket açıklamaları |
| QQQ | İlgili fon bileşenlerinin önemli şirket olayları; Fed/FOMC takvimi, karar ve tutanaklar; ABD enflasyon, istihdam ve büyüme açıklamaları; olay öncesi beklenti ve olay sonrası sonuç; Nasdaq-100 yeniden dengeleme; teknoloji/yarı iletken ihracat kısıtlamaları, tarifeler ve somut etkili siyasi olaylar; fon duyuruları | Fed, BLS, BEA; Nasdaq endeks duyuruları; Invesco; Federal Register, USTR ve Commerce/BIS | AP, CNBC; erişilebilen diğer haberler; Investing.com ve Trading Economics ekonomik takvimleri |
| TQQQ | QQQ ile ortak Nasdaq-100 olay havuzu; ayrıca ProShares fon duyuruları, prospektüs/ek belge değişiklikleri, dağıtımlar ve bölünme/ters bölünme açıklamaları | QQQ'nun birincil kaynakları; ProShares ve ilgili SEC belgeleri | QQQ'nun tamamlayıcı kaynakları; TQQQ'ya özgü açıklamaları aktaran erişilebilir haberler |
| EWC | Kanada faiz kararı ve ekonomik yayın takvimleri; enflasyon, istihdam ve büyüme beklentisi/sonuçları; ilgili Kanada şirket haberleri; CAD/USD; enerji, petrol fiyat farkları ve boru hattı gelişmeleri; ABD–Kanada ticareti/tarifeler; fon bileşen ve sektör ağırlığı değişiklikleri, fon duyuruları | Bank of Canada; Statistics Canada; iShares; ilgili şirket yatırımcı ilişkileri ve erişilebilen SEDAR+ belgeleri | CBC; erişilebilir Reuters Canada ve diğer haberler; Investing.com ve Trading Economics takvimleri |

Bu liste kaynak önceliğidir; her sitenin tüm içeriğinin ücretsiz, düzenli veya otomatik okunabilir olduğu iddiası değildir. Ücretli/erişilemeyen metin okunmuş sayılmaz. SEDAR+ gibi kaynakların okunabilirliği kurulumda sınanır; özel otomasyonu bu aşamada kararlaştırılmadı. Şirket açıklama dağıtıcıları bağımsız habercilik yerine açıklamanın yayımlandığı belge kanalı olarak değerlendirilir.

Genel siyasi/jeopolitik olaylar somut bir şirket, sektör, enerji arzı, ticaret veya fon maruziyeti bağlantısı varsa kapsama girer. Yorum ve fiyat hareketini açıklayan yazılar bütünüyle atılmaz: içindeki kanıtlanabilir olgular ayrılır, yazar yorumu olgu gibi kullanılmaz. Sosyal medya veya arama başlığı tek başına doğrulanmış haber değildir. TQQQ için ortak QQQ olayları yeniden bağımsız haber veya bağımsız kanıt diye sayılmaz. Fon bileşenlerinden kaçının izleneceği evren aşamasında kararlaştırılacak.

### 0.7. Web araması ve doğrulama yöntemi — kesinleşti

1. Olayı canlı web aramasıyla bul; sembolün yanında tam şirket/fon adını, ilgili tarih aralığını ve gerektiğinde `site:` filtresini kullan.
2. Özgün şirket açıklaması, resmi kurum yayını veya fon belgesine ulaş; arama sonucundaki başlığı belgenin tamamının yerine koyma.
3. Aynı iddiayı ikinci kaynakta araştır; yayınların menşeini, aynı ajans/bültenin kopyası olup olmadığını ve hangi olguyu desteklediğini kaydet.
4. Yalnız erişilip okunabilen içeriği kanıt olarak çıkar. Yalnız başlık görülebiliyorsa kaydı `yalnız başlık` olarak işaretle; görünmeyen gövdeyi model tamamlamaz.
5. Olayı `iki bağımsız kaynakla destekli`, `tek kaynaklı`, `çelişkili` veya `yetersiz içerik` olarak etiketle; yeni kanıt geldikçe durum değişikliklerini zamanlarıyla koru.

Haber/diğer bilgi için hedef en az **iki bağımsız kaynakla aynı olgunun desteklenmesidir**. İki URL iki bağımsız doğrulama demek değildir. Aynı ajans haberi veya şirket bülteninin kopyaları tek menşe sayılır. Resmi belge, kurumun/şirketin ne açıkladığına birincil kanıttır; onu aktaran haber ikinci bağımsız kanıt sayılmaz. Resmi açıklamanın yayımlandığı olgu ile açıklamadaki iddianın doğruluğu ayrı tutulur; söylenti desteği tamamlanmış anlaşma anlamına gelmez.

**Tek kaynaklı haberler atılmaz.** Model bunları eldeki diğer kanıtlar ve bağlamla makullük açısından değerlendirebilir; bu yorum ayrı kaydedilir. Modelin hafızası veya “bunu zaten biliyorum” demesi ikinci kaynak yerine geçmez ve kaydı doğrulanmış hale getirmez. Çelişkili sayısal değerler sessizce birleştirilmez; kaynak değerleri ve çelişki korunur. Doğrulama durumu karar girdisinde görünür; bu durumların alım/satıma etkisi karar/risk aşamasında belirlenecek.

Başlangıç sorgu kalıpları (liste taramayı sınırlayan sabit bir sözlük değildir):

| Kategori | Örnek İngilizce aramalar |
|---|---|
| ABD hisseleri | `<COMPANY> <TICKER> 8-K earnings date announced`; `<COMPANY> earnings guidance consensus EPS revenue`; `<COMPANY> definitive agreement acquire`; `<COMPANY> departure officer appointment`; `<COMPANY> court ruling lawsuit` |
| QQQ | `FOMC meeting date statement rate expectations`; `US CPI release consensus forecast`; `US employment report consensus`; `Nasdaq-100 index rebalance announcement`; `Commerce BIS semiconductor export controls`; `USTR tariffs announcement`; ilgili şirket için ABD hisse sorguları |
| TQQQ | QQQ ortak sorguları; `ProShares TQQQ distribution announcement`; `ProShares UltraPro QQQ prospectus supplement`; `TQQQ split reverse split announcement` |
| EWC | `Bank of Canada interest rate announcement statement`; `Statistics Canada CPI Labour Force Survey GDP release`; `Canada economists consensus expected`; `CAD USD exchange rate`; `iShares MSCI Canada ETF holdings`; `Canada US tariffs trade negotiations`; `Western Canadian Select oil differential pipeline`; ilgili Kanada şirketinin tam adıyla duyuru aramaları |

### 0.8. Beklenti, gerçekleşen ve model yorumu — kesinleşti

- Beklenti verisi varsa kaynak, hangi dönem/ölçütü kapsadığı, birimi ve **açıklamadan önce yakalandığı zaman** kaydedilir. Olaydan sonra görülen beklenti, geçmişte bilinen beklenti diye geriye yazılmaz. Ücretsiz sayfada eksikse uydurulmaz; `beklenti mevcut değil` olarak kalır.
- Uzmanın “ücretsiz güvenilir consensus hiç yok” şeklindeki mutlak iddiası seçilmedi: [Investing.com](https://www.investing.com/economic-calendar) ve [Trading Economics](https://tradingeconomics.com/calendar) takvimlerinde beklenti/gerçekleşen alanları görülebiliyor. Bu, tüm olaylar için tam kapsam, aynı anket veya sürekli erişim garantisi değildir; kaynağı belirtilen gözlemler olarak kullanılacak.
- Trading Economics'in web tablosundaki **Consensus** ile kendi **Forecast** değeri ayrı alanlardır. Şirketin yönetim rehberliği analist consensus'u değildir; piyasa kaynaklı olasılık da ekonomist anketinin yerine aynı adla yazılmaz.
- Gerçekleşen değer mümkün olduğunda resmi yayından alınır. Önceki dönem değerleri ve revizyonlar, eski/yeni rehberlik ve tarihleri ayrı tutulur. Aynı dönem, ölçüt, birim ve karşılaştırılabilir tanım varsa gerçekleşen–beklenti farkı hesaplanır; karşılaştırılamayan değerlerde sürpriz üretilmez.
- Bir bilgi ilk haber, tekrar, söylentinin teyidi veya düzeltme olarak işaretlenir. Daha önce gözlenip gözlenmediği kaydedilir; aynı olayın kopyaları yeni olay diye çoğaltılmaz.
- Modelin etki yorumu, makullük değerlendirmesi ve belirsizliği kaynak olgusundan ayrılır. “Piyasa bunu fiyatladı” ifadesi salt haberden doğrulanmış gerçek çıkarmaz; varsa ilgili piyasa kanıtına bağlanan hipotez olarak saklanır.

### 0.9. Haber ve bilgi kayıtları — kesinleşti

Olay ile onu anlatan belgeler ayrı kayıtlanır. Asgari olay alanları: olay kimliği/türü, ilgili varlık veya kategori, planlı/gerçekleşen olay zamanı ve saat dilimi, ilk gözlem zamanı, olgu özeti, beklenti/gerçekleşen/önceki/revizyon/rehberlik değerlerinin kaynak ve dönemleri, yenilik/tekrar/düzeltme durumu, doğrulama durumu ve model yorumudur. Değerlerin para birimi/birimi korunur; eksik zaman veya sayı uydurulmaz.

Asgari belge alanları: URL, yayımlayıcı ve özgün menşe, erişilebilen yayın/güncelleme zamanı, ilk gözlem ve toplama zamanı, gerçekten okunan iddiayı destekleyen kısa metin parçası, gövdeye erişim durumu, desteklenen iddia ve diğer kaynaklarla kopya/bağımsızlık ilişkisi. Zamanlar UTC ve özgün saat dilimiyle saklanır, GMT+3 gösterilir. Sonradan gelen düzeltme eski karar anındaki kanıtı silmez.

Tam sayfa kopyası, hash ve ham içerik arşivinin kapsamı henüz belirlenmedi; saklama/erişim süresi ve teknik kanıt paketi kayıt/altyapı aşamasında netleştirilecek. Kısa kanıt parçası saklamak bütün web ortamının yeniden üretilebildiği anlamına gelmez. Modelin gördüğü bilgi, karar anından sonra öğrenilen bilgiden ayrılacak.

### 0.10. Toplama ve karar takvimi — kesinleşti

Kullanıcı her kategori için işlem gününde **iki rutin araştırma/karar oturumu** istedi: ilk oturum açılıştan 15–30 dakika sonra, ikinci oturum kapanıştan 1 saat önce. İlk oturumun başlangıcı, kullanıcının verdiği aralık içinde **açılış +30 dakika** olarak seçildi. Oturumlarda fiyat, haber ve diğer bilgiler yenilenir; analiz ve al/sat/bekle değerlendirmesi yapılır. Her oturumda mutlaka işlem yapmak gerekmez; karar ve simülatör kuralları sonraki aşamada netleştirilecek.

| Oturum | İlgili normal seansa göre | Normal ABD seansı, New York | GMT+3: ABD yaz saati | GMT+3: ABD standart saati |
|---|---|---|---|---|
| İlk rutin araştırma/karar | Açılış +30 dakika | 10:00 | 17:00 | 18:00 |
| İkinci rutin araştırma/karar | Kapanış −60 dakika | 15:00 | 22:00 | 23:00 |

- ABD hisseleri, QQQ, TQQQ ve EWC için aynı iki oturum kuralı uygulanır. EWC'nin işlem takvimi ABD'deki kotasyonun takvimidir; Kanada ekonomik olayları kendi açıklama zamanlarıyla ayrıca takip edilir.
- Saatler sabit İstanbul saatleri olarak yıl boyu yazılmaz; ilgili borsanın o günkü açılış/kapanışından hesaplanır, GMT+3 gösterilir. Tatil ve hafta sonunda rutin işlem oturumu yoktur. Erken kapanış gününde ikinci oturum erken kapanıştan 1 saat önceye kayar.
- Tablo oturum **başlangıç** zamanlarını gösterir. Gerçek başlama, bilgi toplamanın bitişi, karar ve emir zamanları ayrı kaydedilir. Toplamaya başlandığı anda görülen fiyat, daha sonra emir öncesi güncelmiş gibi kabul edilmez.
- Kullanıcı **önemli olaylarda ek kontrolü onayladı**. Planlı bilanço, faiz kararı ve önemli ekonomik açıklamalarda açıklama öncesi beklenti/takvim, açıklama sonrası gerçekleşen değer ve yeni haber kontrol edilir. Planlanmamış önemli olay fark edildiğinde ilgili kategori için ek bilgi/doğrulama kontrolü yapılır. Bu kontroller iki rutin oturuma ek olabilir; kesintisiz haber izleme taahhüdü değildir. Olay öncesi/sonrası dakika aralıkları ve tekrar denemeler zamanlayıcı tasarımında netleştirilecek.
- Her alım/satım öncesi fiyat **yeniden üç kaynaktan doğrulanır**; bölüm 0.2–0.5'teki aynı ürün/seans, en fazla %1 fiyat açıklığı, en fazla 60 saniye kaynak zamanı açıklığı ve en fazla 60 saniye fiyat yaşı koşulları geçerlidir. Başarısız doğrulamada işlem gerçekleşmez.
- Ek olay kontrolleri otomatik olarak ek işlem oturumu oluşturmaz; iki rutin oturum dışındaki emir/risk davranışı sonraki karar ve risk aşamasında ayrıca belirlenecek.

### 0.11. ABD hisse evreni — kesinleşti

Kullanıcı sabit ve dinamik izleme listesini birlikte kullanmayı kabul etti. Bu evren kararı **ABD bireysel hisse kategorisi** içindir; dört kategorinin cüzdan ve karakter yapısı bölüm 0.12'dedir.

- **Sabit liste: 50 farklı hisse.** Kullanıcı seçimi asistana bıraktı; aşağıdaki başlangıç listesi seçildi. Ölçüt, ABD'de işlem gören büyük şirketler ve farklı sektör/olay türlerinin temsilidir; en yüksek gelecekteki getiriyi sağlayacak 50 hisse sıralaması değildir. Liste günlük/haftalık model seçiminden bağımsızdır. Kotasyondan çıkma, birleşme veya sembol değişimi gibi zorunlu olaylarda varlık kimliği/pozisyon geçmişi korunacak.
- **Dinamik liste: sabit 50 dışında 20 farklı hisse.** Model seçer; sabit listeyle veya kendi içinde mükerrer varlık olamaz. Birlikte hedef izleme evreni 70 farklı ABD hissesidir.
- Dinamik liste **haftalık** ekleme/çıkarma ile güncellenebilir; günlük değiştirilemez. Kullanıcı seçim kuralı ve zamanı asistana bıraktı. Güncelleme ilgili ABD borsa haftasının **ilk işlem günü 09:00 America/New_York** saatinde başlayacak; o haftanın ilk rutin karar oturumundan önce liste sürümü oluşturulacak. Normal durumda GMT+3 karşılığı ABD yaz saatinde 16:00, standart saatte 17:00'dır. İlk işlem günü tatil nedeniyle pazartesi olmayabilir. Zamanında oluşmayan yeni sürüm geçmişe yazılmaz; önceki geçerli sürüm ve güncelleme başarısızlığı açık kaydedilir.
- Dinamik 20, ortak araştırma işlevi tarafından sabit 50 dışında seçilir: yaklaşan doğrulanmış bilanço/şirket olayları, yeni önemli açıklamalar, farklı sektör fırsatları ve erişilebilir fiyat/haber kanıtı esas alınır. Tek bir cüzdanın pozisyonu, hafızası veya kârı seçim girdisi olmaz. Cüzdanlara aynı izleme listesi sunulur; seçim gerekçelerindeki kaynak olguları ile model yorumu ayrılır. Üç ABD cüzdanının özel hafızası birbirine açılmadan ortak 20 liste üretiminin teknik ayrıntıları altyapıda kurulacak.
- Tam 20 uygun aday kanıtlanamıyorsa sahte sembol veya bilgiyle liste tamamlanmaz; eksiklik kaydedilir. İlk dinamik liste pilot başlangıcında seçilecek, bu belgede bugünden geriye dönük bir model seçimi gibi yazılmayacak. Her sürüm, eklenen/çıkarılan hisse, gerekçe, seçim zamanı ve yürürlüğe giriş zamanı kaydedilir; günlük değiştirilmez ve sonradan edinilmiş bilgiyle geçmiş liste yeniden yazılmaz.
- 70 hisse takip evrenidir; 70'inde aynı anda pozisyon açma veya hepsini her oturumda derin araştırma zorunluluğu değildir. İki rutin oturumda genel tarama yapılır; önemli gelişmesi bulunan, elde tutulan pozisyonla ilgili olan ve işlem düşünülen varlıklarda derinleşilir. Her emir için mevcut üç kaynaklı fiyat doğrulaması korunur.
- Üç cüzdan ortak 70 hisselik evreni ve ortak veri havuzunu kullanır; her biri ayrı portföy, karar günlüğü ve sonuç kaydı taşır. Her cüzdan için ayrı 20 hisse seçilerek toplam dinamik liste genişletilmez.

ABD kategorisinde **üç** sanal cüzdan olacak. Sekiz cüzdanın her biri **100.000 USD** başlangıç nakdiyle açılır. Karakterler ve modelin tutma süresini seçme serbestliği bölüm 0.12'de, en yeni kullanıcı kararları bölüm 0.14'te tutulur.

Sabit 50 başlangıç listesi — 4 Ekim 2026:

| Araştırma grubu | Sayı | Hisseler |
|---|---|---|
| Bilgi teknolojileri | 12 | NVDA — NVIDIA; AAPL — Apple; MSFT — Microsoft; AVGO — Broadcom; AMD — AMD; MU — Micron; ORCL — Oracle; CRM — Salesforce; CSCO — Cisco; QCOM — Qualcomm; AMAT — Applied Materials; PLTR — Palantir |
| İletişim hizmetleri | 5 | GOOGL — Alphabet A; META — Meta; NFLX — Netflix; DIS — Disney; VZ — Verizon |
| İsteğe bağlı tüketim | 5 | AMZN — Amazon; TSLA — Tesla; HD — Home Depot; MCD — McDonald's; BKNG — Booking |
| Temel tüketim | 5 | WMT — Walmart; COST — Costco; PG — Procter & Gamble; KO — Coca-Cola; PEP — PepsiCo |
| Finans | 8 | BRK.B — Berkshire Hathaway B; JPM — JPMorgan Chase; BAC — Bank of America; WFC — Wells Fargo; GS — Goldman Sachs; MS — Morgan Stanley; V — Visa; MA — Mastercard |
| Sağlık | 6 | LLY — Eli Lilly; JNJ — Johnson & Johnson; ABBV — AbbVie; MRK — Merck; UNH — UnitedHealth; TMO — Thermo Fisher |
| Sanayi | 4 | CAT — Caterpillar; GE — GE Aerospace; RTX — RTX; UNP — Union Pacific |
| Enerji | 2 | XOM — ExxonMobil; CVX — Chevron |
| Malzeme | 1 | LIN — Linde |
| Kamu hizmetleri | 1 | NEE — NextEra Energy |
| Gayrimenkul | 1 | AMT — American Tower |

Semboller/hisse sınıfları [iShares OEF'nin 1 Ekim 2026 tarihli yayımlanan varlık listesi](https://www.ishares.com/us/products/239723/ishares-s-p-100-etf/latest-holdings.csv) ile kontrol edildi; seçim ve gruplama asistanın araştırma kapsamı kararıdır. Bu liste OEF'deki ağırlıkları portföye uygulamaz. Aynı şirketin ikinci sınıfı ayrıca sayılmadı; Alphabet için GOOGL seçildi. Berkshire'ın kaynaklarda `BRK B`, `BRK.B` veya `BRK-B` yazımı varlık kaydında aynı B sınıfına eşlenecek. Canlı fiyat, haber ve kotasyon erişimi her varlık için kurulumda sınanacak; bu kontrol üç canlı fiyatın doğrulandığı anlamına gelmez.

Bir hissenin dinamik izleme listesinden çıkarılması **otomatik satış değildir**. Eldeki pozisyon kapanana kadar o cüzdan için fiyat/haber takibi, değerleme ve satış yetkisi devam eder; varlık normal yeni alış listesine kendiliğinden geri girmez. Pozisyon dışında kalan takip ve yeni emir uygunluğu simülatörün ayrıntılı sözleşmesinde tamamlanacak. Haftalık liste değişikliği ile iki günlük işlem oturumu ayrı kararlardır.

### 0.12. Sekiz sanal cüzdan ve yatırımcı karakterleri — kesinleşti

Kullanıcı, zorunlu “şu kadar tut/şu gün sat” süreleri yerine cüzdan sahiplerinin karakter ve amaçlarının tanımlanmasını istedi ve aşağıdaki değerlendirmeyi kabul etti. **Toplam sekiz sanal cüzdan** vardır. Karakterler araştırma deneyi için motivasyon tanımıdır; kanıtlanmış algoritma veya kazanç vaadi değildir.

| Cüzdan kimliği | Kategori | Karakter ve amaç |
|---|---|---|
| US-FAST | ABD hisseleri | Paraya ihtiyaç duyan, mümkün olan en kısa zamanda sermayesini büyütmek isteyen hızlı kazanç odaklı karakter. Fırsata göre hızlı al/sat yapabilir veya pozisyonunu koruyabilir; giriş, çıkış ve tutma süresini model seçer. Önceki trend/momentum cüzdanının yerini bu karakter tanımı alır. |
| US-EVENT | ABD hisseleri | Olay odaklı, dengeli karakter. Bilanço, rehberlik ve önemli haberleri değerlendirir; ne aceleci ne de çok uzun vadeli bir yaklaşım benimser. Olay tezinin gelişimine göre süreyi model seçer. |
| US-PATIENT | ABD hisseleri | Sabırlı, orta–uzun vadeli yatırımcı. Şirketin finansal durumu, değerlemesi ve zaman içinde gelişimine odaklanır. |
| QQQ-FAST | QQQ | Mümkün olan en kısa zamanda sermayesini büyütmek isteyen hızlı kazanç odaklı karakter. |
| QQQ-PATIENT | QQQ | Uzun vadeli yatırımcı karakteri. |
| TQQQ-FAST | TQQQ | Mümkün olan en kısa zamanda sermayesini büyütmek isteyen hızlı kazanç odaklı karakter; ürünün günlük kaldıraç yapısını hesaba katar. |
| TQQQ-PATIENT | TQQQ | Uzun vadeli yatırımcı karakteri; günlük kaldıraç ve çok günlük bileşik getiri etkisini değerlendirerek karar verir. |
| EWC-BALANCED | EWC | Dengeli, standart yatırımcı karakteri. Fırsat ve riski birlikte değerlendirir; belirgin bir acelecilik veya aşırı uzun vade zorlaması yoktur. |

- Karakterlerin vade eğilimi vardır; **sabit veya zorunlu pozisyon tutma süresi yoktur**. US-PATIENT orta–uzun vadeyi, ETF sabırlı kolları uzun vadeyi amaçlar; bunlar otomatik süre dolumu/satış kuralı değildir. Hızlı karakterin bir pozisyonu iki saat veya bir hafta tutabilmesi modelin seçimine bırakılır.
- “En kısa zamanda zengin olma” karakter motivasyonudur; performans masraflar sonrası kazanç ve yaşanan düşüş gibi ölçülerle değerlendirilecek. Kullanıcının son kararıyla ek stratejik risk sınırları konulmayacak; gerçekçi broker/ürün ve gerçekleşme koşulları bölüm 0.14'te ayrılır. Daha fazla risk alma ile daha iyi karar verme değerlendirmede ayrılacak.
- Her karar, neden alındığı, hangi gelişmenin beklendiği ve hangi koşulda çıkılacağı gibi somut gerekçeler taşıyacak. Modelin kendi seçtiği beklenti süresi varsa karar anında kayıtlanabilir; bu, tüm işlemlere dayatılan tutma süresi değildir. Ayrıntılı çıktı şeması sonraki karardır.
- Cüzdanların karar, açık pozisyon, nakit ve sonuç kayıtları ayrıdır; kaynak olguları ortak havuzdan sağlanabilir. Karakter etkisini model sürümü/erişim farklarından ayıracak deney ve kıyas ayrıntıları henüz belirlenmedi.
- İki günlük rutin oturum arasında örneğin iki saat sonra çıkış yapmanın nasıl yürütüleceği henüz belirlenmedi: koşullu emir, ek karar kontrolü, fiyat tetikleyicisi ve bunların fiyat doğrulaması emir mekanizmasında netleştirilecek. Ek haber kontrolü kendiliğinden sınırsız ek işlem yetkisi değildir.
- TQQQ günlük Nasdaq-100 performansının masraf öncesi üç katını hedefler. Çok günlük getiri endeksin aynı dönem getirisinin üç katı olarak varsayılmaz. Sabırlı TQQQ kolu ayrı sanal deneydir; QQQ sabırlı kolunun fiyatı üçle çarpılarak simüle edilmez. Kaynak: [ProShares TQQQ fon açıklaması](https://www.proshares.com/our-etfs/leveraged-and-inverse/tqqq).

### 0.13. Kalan planlama kararları — sırayla tamamlanacak

Kullanıcı aşağıdaki kalan maddelerin de kaydedilmesini istedi. Her konu dört kategori ve ilgili karakterler bakımından konuşulacak; bu liste açık kararları gösterir, aşağıdaki seçenekleri henüz kesinleştirmez.

1. **Sabit 50 hisse ve dinamik liste kuralları:** Seçim ölçütleri ve isimler; modelin ilk 20 seçimi; haftalık değişimin gün/saati; liste uygunluğu ve zorunlu değişiklikler; çıkarılan hissenin açık pozisyonunun durumu.
2. **Başlangıç sermayesi ve cüzdan ayrılığı:** Her cüzdanın sahte para tutarı, nakit kullanımı, dış para girişi/çekimi ve cüzdanlar arası aktarımın kapsamı; karşılaştırmada ortak varsayımlar.
3. **Risk ve izin verilen işlemler:** Pozisyon büyüklüğü, yoğunlaşma, eşzamanlı pozisyon sayısı, kayıp/düşüş sınırları, durdurma davranışı; açığa satış, borçlanma veya başka araçların kapsamı; TQQQ için ürün riskine uygun sınırlar.
4. **Karakterlerin model talimatları ve karar çıktısı:** Nihai karakter metinleri, modele sunulan portföy/veri, al/sat/bekle, miktar, dayanak, beklenti ve çıkış koşulları; gerekli geçmiş fiyat/finansal ölçütlerin okunabilirliği ve eksik verinin davranışı. Zorunlu tutma süresi eklenmeyecek.
5. **Emir ve simülatör kuralları:** Emir türleri/ömrü, gerçekleşme fiyatı seçimi, komisyon, spread, kayma, hacim/likidite sınırı, kısmi gerçekleşme, bölünme ve dağıtım işlemleri; iki rutin oturum arasındaki koşullu çıkışların çalışma biçimi.
6. **Ek olay kontrolleri ve çalışma hataları:** Olay öncesi/sonrası dakika aralıkları, rutin dışı karar yetkisi, zaman aşımı, erişilemeyen kaynaklar, tekrar deneme ve kota yetersizliğinde davranış. Canlı üç kaynaklı fiyat koşulları kurulumda sınanacak.
7. **Hafıza, kanıt saklama ve sonuçtan öğrenme:** Her cüzdanın modele taşınacak geçmişi, belge/kanıt arşivi ve saklama süresi, model/prompt sürümleri, düzeltme geçmişi; sonradan değerlendirme ile canlı karakter/prompt değişikliklerinin ayrılması.
8. **Başarı ölçütleri ve kıyaslar:** Yıllık %10 net hedefinin tam tanımı, masraflar, düşüş ve risk ölçüleri; kategori/karaktere uygun al-tut ve basit strateji kıyasları; hızlı ve sabırlı kolların sonuçlarının nasıl karşılaştırılacağı.
9. **Deney süresi ve değerlendirme disiplini:** Pilotun süresi, raporlama aralığı, geliştirme ile ileri test ayrımı, karakter/model değişikliklerinin kaydı ve devam/durdurma koşulları.
10. **Görüntüleme ve raporlar:** Kullanıcının cüzdanları, açık pozisyonları, gerekçeleri, veri doğrulama durumunu, kazanç/düşüşü ve sistem hatalarını nereden/nasıl izleyeceği.
11. **Teknik altyapı ve kurulum sırası:** Önce GPT abonelik bağlantısı; sunucu/zamanlayıcı, veri ve cüzdan ayrılığı, erişim ve kota testi; Gemini ve model API seçeneklerinin sonraki kapsamı. Kodlama tüm planlama ve genel revizyon sonrası başlayacak.
12. **Son genel revizyon ve uygulama planı:** Bütün belgeleri bu güncel kararlara göre uzlaştırma; eski sekiz ETF/API taslağını güncel kapsamdan ayırma; kabul kontrolleri, iş sırası ve sorumluluklarla kodlama planını çıkarma. Tarihsel ürün B'nin ayrıntılı tasarımı ayrı sonraki kapsamdır.

### 0.14. İlk on madde için en yeni kullanıcı kararları — 5 Ekim 2026

Bu bölüm, önceki açık karar listesiyle çeliştiği yerde önceliklidir. Kullanıcı ilk on başlığı topluca netleştirmeyi; teknik altyapı ve son genel revizyonu daha sonra ele almayı istedi. Henüz kodlama veya gerçek hesap bağlantısı yapılmayacak.

- **Bağımsız sekiz cüzdan:** Ortak veritabanında ortak kaynak olguları/fiyat/haber havuzu bulunur; her cüzdanın özel geçmişi, kararları, notları, karakteri, nakdi ve pozisyonları ayrıdır. Diğer cüzdanın çıktıları, hafızası ve performansı modele verilmez. Aynı GPT erişimiyle sekiz bağımsız model yürütüyormuş gibi ayrı bağlamlar sürdürülecek; sekiz farklı temel model kullanımı kararlaştırılmadı. Özel belleğin teknik yapısı kullanıcı isteğiyle altyapı aşamasına bırakıldı.
- **Kör deney amacı:** Cüzdan modellerine simülasyon bilgisi, diğer cüzdanların varlığı ve araştırmacının %10 başarı ölçütü sunulmayacak. Bu bilgiler kullanıcı/deney kayıtlarında kalır. Modelin göreceği portföy, kaynaklar ve emir sonuçları yalnız kendi hesabıyla ilgilidir. Gerçek broker bağlantısı olmadan kullanıcı ekranında bu hesaplar sanal olarak tanımlanır; modelden gizlenen deney metadatası kaynak fiyatlarını veya gerçekleşme kaydını uydurmaya gerekçe değildir.
- **Sermaye:** Her cüzdan başlangıçta **100.000 USD** nakit ve sıfır açık pozisyonla başlar; sekiz cüzdanın toplam başlangıç sermayesi 800.000 USD'dir. **Dışarıdan para ekleme, para çekme ve cüzdanlar arası aktarım yoktur. Borçlanma ve faiz yoktur.** Yatırım kazancı, zarar, dağıtım ve maliyetler kendi cüzdanında kalır. İşlem bütçesi kendi sermayesi ve kullanılabilir nakdidir; hesaptaki varlığı aşan satın alma gücü üretilmez. 100.000 USD başlangıç sermayesi olarak tutulur; gerçekleşmiş kazanç/kayıp sonrasında güncel bakiye esas alınır.
- **Kategori sınırı:** ABD hisse cüzdanları kendi izleme evrenindeki hisseleri, QQQ cüzdanları yalnız QQQ'yu, TQQQ cüzdanları yalnız TQQQ'yu, EWC cüzdanı yalnız EWC'yi alıp satabilir ve nakitte kalabilir. İşlem kapsamı mevcut parayla hisse/ETF al-sattır. Açığa satış, margin/kredili işlem, ödünç hisse, opsiyon ve vadeli sözleşmeler bu pilotun kapsamında değildir. Model yalnız sahip olduğu miktarı satabilir. TQQQ'nun fon içi günlük kaldıraç yapısı korunur; cüzdanın ayrıca brokerdan borçlanması anlamına gelmez.
- **Stratejik risk serbestliği ve nakit:** Kullanıcı ek portföy yüzdesi, pozisyon sayısı, stop zorunluluğu veya zarar eşiği istemiyor. Model kendi parasının ne kadarını yatıracağını ve **gelecekteki fırsatlar için ne kadar nakit tutacağını** seçer; zorunlu nakit yüzdesi konulmaz. Alış tutarı ve işlem maliyetleri kullanılabilir nakdi aşamaz; açık alış emirlerinde ayrılmış para başka emirlerde yeniden kullanılamaz. Satış geliri ve takası bekleyen tutarlar nakit hesabında ayrı izlenir, kendiliğinden kredi yaratmaz. Piyasa saati, veri doğrulaması ve takas/gerçekleşme koşulları uygulanır.
- **Simülatörün hazırlanması:** Kullanıcı emir/gerçekleşme ve maliyet tasarımını asistana bıraktı. Artık **nakit hesap, yalnız eldeki hisse/ETF al-sat, normal seans** kapsamı esas alınacak. Doğrulanmış güncel fiyat kanıtı, yönü doğru bid/ask kullanımı, zamanında emir, komisyon/spread/kayma ve ilgili dağıtım/stopaj muhasebesi taslağın temelidir. Hesapta borç, borç faizi, ödünç maliyeti veya nakit faiz geliri oluşturulmaz. ETF'nin piyasa fiyatına/NAV'ına zaten yansıyan fon giderleri ikinci kez cüzdandan kesilmez; TQQQ'nun kendi gerçek fiyat serisi kullanılır. Gelecekteki fiyattan veya doğrulanmış son fiyatı otomatik gerçekleşme garantisi yaparak doldurma yok. Kanıtlanamayan emir doldurulmuş sayılmaz; kullanıcıdan her işlem için onay istenmeyecek.
- **Kota ve hata yönetimi:** Abonelik kotasını kullanıcı yönetir; rutin kullanıcı onayı gerekmiyor. Geçersiz/eksik kaynak veya başarısız emir işlemi otomatik tamamlanmış gösterilmez. Olay kontrollerini kullanıcının tek tek yönetmesi gerekmiyor.
- **Emir zamanı:** Modeller iki rutin karar oturumunda emir verir; arada önceden verilmiş koşullu emirler çalışabilir. Yalnız **normal seans** içinde işlem gerçekleşir; seans öncesi/sonrası işlem yapılmaz. Ek olay kontrolleri bilgi günceller; kendi başına üçüncü bir işlem oturumu veya ek model emri oluşturmaz. Koşullu emir de üç kaynaklı fiyat doğrulaması ve gerçekçi gerçekleşme şartlarını geçmelidir; geçmişte tetiklenmiş olabilecek emir sonradan gerçekleşmiş diye yazılmaz. Emir türleri ve süreleri araç kapsamı netleştikten sonra simülatörde tamamlanacak.
- **Karakter ve öğrenme:** Karakter tanımı sabit kalır; model bu karakterin amaçları içinde yöntemini sonuçlarına göre geliştirebilir. Eski karar ve yöntem kayıtları korunur. Öğrenmenin teknik bellek tasarımı altyapı aşamasında konuşulacak.
- **Vergi/maliyet kapsamı:** Broker maliyetleri ve varsa ilgili temettü stopajı işlenecek; kişisel gelir vergisi ayrı tutulacak. Banka transferi yalnız gerçekleşirse maliyet yaratır; para giriş/çıkışı yasak olduğu için başlangıç sonrası banka transferi planlanmıyor. Başlangıç 100.000 USD hesabın içinde hazır nakittir. Stopaj oranı araç/hesap profiline ve kaynağına bağlı olduğundan tek bir oran henüz kilitlenmedi; kişisel vergi sonrası kazanç iddiası yapılmaz.
- **Başarı:** %10 net getiri modele görev, hedef veya başarı eşiği olarak verilmeyecek. Araştırmacının kendi değerlendirme referansıdır. Gerçekleşen ve gerçekleşmemiş sonuç, masraflar ve düşüş araştırmacı için saklanabilir; modele gizli başarı puanı sızdırılmaz.
- **Süre:** Pilot için zorunlu bitiş süresi veya %10'a ulaştığında otomatik durdurma kuralı yoktur; kullanıcı gelişime bakarak karar verecek. Başlangıç zamanı, tüm dönem sonucu ve sonraki kural değişiklikleri korunur; seçilmiş kısa dönem sonucu yıllık başarı garantisi gibi sunulmaz.
- **Ekranın asgari içeriği:** Kullanıcı açtığında her cüzdanın kullanılabilir/ayrılmış/takası bekleyen nakdini, tuttuğu varlıkları ve bunların güncel değerini, toplam özkaynağını görecek. Değerleme zamanı ve güncel olmayan/eksik fiyat durumu görünür; `toplam nakit + eldeki pozisyon değeri + diğer alacaklar − tahakkuk etmiş diğer yükümlülükler` hesabı temel alınır. Aynı satış geliri hem alacak hem nakit olarak iki kez sayılmaz. Kredi ve kısa pozisyon oluşturulmaz. Görsel düzen ve ayrıntılı ekran kapsamı daha sonra konuşulacak.

Maliyet/gerçekçilik için kontrol edilen referanslar: [IBKR ABD hisse/ETF komisyonları](https://www.interactivebrokers.com/en/pricing/commissions-stocks.php), [SEC yatırımcı bülteninde T+1 takas](https://www.investor.gov/introduction-investing/general-resources/news-alerts/alerts-bulletins/investor-bulletins/new-t1-settlement-cycle-what-investors-need-know-investor-bulletin). Bunlar referans araştırmasıdır; gerçek IBKR hesabı açılmadı, ürün/ikamet/hesap profiline özgü bütün oranların doğrulandığı veya birebir broker simülatörü tamamlandığı iddia edilmez. Bir sonraki somut taslak komisyon, dağıtım/stopaj ve nakit/takas muhasebesini tarihli kurallara bağlayacak; eksik değeri tahminle kullanırsa varsayım olduğu işaretlenecek. Önceki borç/teminat araştırması yöntem geçmişidir, nakit hesap pilotunda uygulanmayacak.

### 0.15. İşlem yetkisi netleşti; sıradaki tartışma

Kullanıcının **5 Ekim 2026** yanıtı: faiz ve borçlanma yok; GPT başlangıçta verilen 100.000 USD öz parayla plan yapacak, gerekirse fırsatlar için nakit ayıracak. İşlem kapsamı buna göre **nakit hesapla hisse/ETF al-sat** olarak kaydedildi. İlk on başlık için sorulan işlem yetkisi sorusu kapandı; zorunlu nakit oranı veya yeni stratejik risk sınırı eklenmedi.

Sabit 50, haftalık dinamik liste ve simülatör tasarımı kullanıcı tarafından asistana bırakıldı; bunlar için yeniden onay istenmiyor. Kanıtlanabilir komisyon/stopaj oranları, nakit takas ayrıntıları, koşullu emir gözlem sıklığı ve veri erişimi somut taslak/kurulum kontrollerinde tamamlanacak; kullanıcıdan tek tek teknik değer seçmesi istenmez. Bu ayrıntıların bütününün uygulanmış veya doğrulanmış olduğu iddia edilmez. Sıradaki kullanıcı tartışması ekran kapsamıdır; ardından hafıza teknik tasarımı, altyapı ve son genel revizyon ele alınacak.

### 0.16. Kapsam kontrolünden sonra kabul edilen tamamlamalar — 5 Ekim 2026

Kullanıcı altı değerlendirme maddesine aşağıdaki kararları verdi. Bunlar yeni cüzdan, ek model karar oturumu veya ücretli veri kaynağı eklemez.

1. **Alternatif fiyat kaynağı:** Okunamayan/geciken kaynaklar için yeni ücretsiz site araştırılacak. Yahoo/Nasdaq/Stock Analysis mevcut başlangıç listesi olmakla birlikte, erişim testi sonrası değiştirilebilir. Yeni site henüz seçilmedi; değişiklik ve erişim kanıtı kaydedilmeden “üç kaynak tamam” denmez. Üç kaynak, en fazla %1 açıklık ve 60 saniye zaman/güncellik koşulları korunur.
2. **Aktif koşullu emirlerde fiyat kontrolü:** İki günlük model karar oturumuna ek olarak, normal seansta **aktif koşullu emirlerin ilgili sembolleri için fiyat odaklı ayrı kontrol** çalışacak. Bu kontrol genel 70 hisse haber taraması veya modelin yeni karar oturumu değildir; daha önce verilmiş emir koşullarını gözler. Kontrol sıklığı, erişim/yeniden deneme ve gözlem gecikmesi teknik zamanlayıcıda netleştirilecek. Fiyat kanıtı olmadan kaçırılmış geçmiş tetikleme gerçekleşmiş sayılmaz; gerçekleşme öncesinde mevcut üç kaynaklı doğrulama uygulanır.
3. **Simülatör ayrıntıları sonra:** Kullanıcı sayısal komisyon/spread/kayma, emir ömrü ve nakit takas ayrıntılarının sonraki aşamada somutlaştırılmasını istedi. Bunlar tamamlanmış veya kilitlenmiş sayılmaz; tasarım yetkisi asistandadır.
4. **Listeden çıkan eldeki varlık:** Açık pozisyonun takibi ve satış yetkisi kapanana kadar sürer; listeden çıkarma zorunlu satış oluşturmaz. Kural bölüm 0.11'e işlendi.
5. **Backend emir ve cüzdan tutarlılığı:** Model emir aracına başvurur; yanıt aynı backend muhasebesinin emir/gerçekleşme kaydından üretilir. Emir kimliği, kabul/bekleme/ret/kısmi/tam gerçekleşme durumu, gerçekten gerçekleşen miktar, fiyat, masraf, zaman ve güncel nakit/pozisyon birbiriyle tutarlı olur. Emir alınması tek başına satış başarısı sayılmaz; başarılı gerçekleşme varsa cüzdan hareketi o kayda bağlanır. Modelin kendi cüzdanındaki sonucuna güvenebilmesi için kayıt tutarlılığı sağlanacak. Deney metadatası model bağlamından ayrı kalır; bu tasarım modelin kendiliğinden simülasyonu çıkaramayacağını kanıtlamaz.
6. **Araştırmacıya özel kıyaslar:** Modelden gizli değerlendirmede aynı başlangıç tarihi/sermayesi ve maliyet yaklaşımıyla basit al-tut kıyasları bulunacak. QQQ/TQQQ/EWC için kendi ETF'sinin al-tut kolu; ABD hisseleri için karşılaştırılabilir sabit evren/geniş piyasa referansı ayrıntıları sonradan belirlenir. Getiri, masraf ve düşüş araştırmacı ekranında değerlendirilir; kıyas sonucu veya %10 referansı cüzdan modeline verilmez.

### 0.17. Kendi Go backend'imiz ve web ekranı — 8 Ekim 2026

**Kullanıcının yeni kararı:** Emirler, gerçekleşmeler, nakit ve pozisyonlar kendi backend'imizde tutulacak; bunları gözlemlemek için basit bir web uygulaması yeterli. Hazır trading API/backend ürünü kullanılmayacak. Backend **Go** ile yazılacak; HTTP framework/router seçimi asistana bırakıldı. Kod yazma öncesi planlama ve son genel revizyon koşulu sürüyor.

**Seçilen HTTP katmanı: Go + chi v5.** chi, Go'nun standart `net/http` yapısıyla uyumlu yönlendirme/middleware sağlar; emir motoru, muhasebe veya hazır broker backend'i değildir. İş kuralları ve uygulama API'si bizim tarafımızdan yazılacak. Kaynak: [chi resmi proje belgesi](https://github.com/go-chi/chi).

Aşağıdaki bileşenler **teknik mimari önerisidir**; kullanıcı henüz bütün teknik ayrıntıları kilitlemedi ve çalışır uygulama yoktur:

| Bileşen | Öneri | İşlev |
|---|---|---|
| Backend ve kendi HTTP API'miz | Tek Go uygulaması, chi | Emir doğrulama, nakit/pozisyon muhasebesi, portföy okuma, kullanıcı ve ajan yetkilerinin ayrılması |
| Veritabanı | PostgreSQL | Ortak kaynak verileri; cüzdan bazlı emir, gerçekleşme, para hareketi, özel hafıza ve iş kayıtları |
| Zamanlayıcı/iş yürütme | Go worker; PostgreSQL'de kalıcı iş kuyruğu | İki günlük oturum, haftalık liste ve olay kontrolleri, aktif koşullu emir fiyat gözlemleri; yeniden başlatma sonrası devam |
| GPT bağlantısı | Go'nun çalıştırdığı resmi Codex CLI; ChatGPT abonelik oturumu | Cüzdan başına ayrı çalışma bağlamında karar üretme; Gemini daha sonra |
| Basit web ekranı | Go HTML şablonları, küçük JavaScript; aynı uygulamadan sunum | Sekiz cüzdan özeti, portföy/işlem ayrıntısı, kaynak kanıtları, karar gerekçesi ve araştırmacıya özel kıyaslar |

Başlangıçta tek sunucuda kendi Go uygulaması, PostgreSQL ve Codex runner kurulumu yeterli mimari hedefidir. Sunucu boyutu/işletim sistemi ve dağıtım yöntemi henüz seçilmedi. Redis, Kafka, çok sayıda mikroservis veya hazır portföy backend'i MVP gereksinimi olarak eklenmedi. Web ekranı kullanıcıya ait gözlem yüzeyidir; modele bu ekranın yönetici görünümü açılmaz.

**Karar/işlem akışı önerisi:**

1. Zamanlayıcı o oturumun ortak kaynak veri sürümünü hazırlar ve her cüzdan için ayrı iş kaydı oluşturur.
2. Runner yalnız ilgili cüzdanın karakterini, özel geçmişini ve backend'den alınmış güncel nakit/pozisyonlarını modelin kendi oturumuna verir. Kaynak olguları ortak havuzdan okunur; diğer cüzdanların karar, hafıza ve sonuçları verilmez.
3. Codex oturumu al/sat/bekle ve koşullu emir önerilerini yapılandırılmış çıktıda üretir. Go bağlantı katmanı bunları kimliği o cüzdana bağlı kendi backend emir uçlarına iletir; serbest metin muhasebe işlemi sayılmaz.
4. Backend kapsam, normal seans, miktar, eldeki pay, ayrılmamış nakit, maliyet ve üç kaynaklı güncel fiyat koşullarını denetler. Model uygulama kurallarını atlayarak DB'ye yazamaz.
5. Kabul/bekleme/ret/gerçekleşme sonucu kalıcı kayda girer. Nakit hareketi, pozisyon ve gerçekleşme bir veritabanı transaction'ında tutarlı güncellenir; aynı emrin tekrarlı gönderimi ikinci kez işlem oluşturmaz.
6. Sonuç aynı cüzdanın Codex oturumuna geri verilir; özel not ve karar geçmişi o cüzdana bağlanır. Web ekranı backend kayıtlarını okur; cüzdan toplamını modelin söylediği sayıdan hesaplamaz.

**Abonelik bağlantısının belge dayanağı:** Resmi Codex CLI ChatGPT aboneliğiyle kimlik doğrulamayı destekliyor. `codex exec` kayıtlı CLI kimliğini kullanabiliyor; yapılandırılmış çıktı ve açık session ID ile devam etme belgelenmiş. Mimari başlangıç tercihi ayrı oturumları Go runner ile yürütmek ve çıktı/sonuç döngüsünü kendi kodumuzda kurmaktır. Bu, ChatGPT web ekranına sekiz sohbet açmak veya aboneliği doğrudan OpenAI API anahtarı gibi kullanmak değildir. Hesabın/sunucunun gerçek erişimi kurulumda sınanacak. Kaynaklar: [Codex kimlik doğrulama](https://learn.chatgpt.com/docs/auth), [non-interactive çalışma ve resume](https://learn.chatgpt.com/docs/non-interactive-mode), [CLI komutları](https://learn.chatgpt.com/docs/developer-commands).

**Hafıza ve erişim izolasyonu önerisi:** Her cüzdanın ayrı açık `session_id` eşlemesi, çalışma dizini, özel notları ve yetkili bağlantı bağlamı olacak. Ortak `--last` veya ortak yazılabilir hafıza kullanılmayacak. DB aynı olsa da ajan kendi token/kimliğiyle yalnız kendi hesap uçlarını görecek; yönetici/list-all-cüzdan uçları kullanıcıya ait olacak. Diğer cüzdan dosyaları, backend DB kimliği, deney/kıyas metadatası ve yönetici sırları model çalışma ortamına taşınmayacak. Teknik izolasyon dosya/süreç ve HTTP yetkilendirmesinde kurulacak; sadece promptla “diğerlerini görme” talimatına dayanılmayacak. Aynı cüzdanın iki işi eşzamanlı kendi oturumunu değiştirmeyecek. Oturum kaybında yeni oturum aynı cüzdanın kayıtlı özel geçmişinden başlatılır; ortak veya başka cüzdan hafızası kullanılmaz.

**Veri ve muhasebe önerisi:** Ortak varlık kimlikleri, liste sürümleri, fiyat kanıtları ve haber belgeleri ayrı tablolardır. Cüzdanlar, emirler, gerçekleşmeler, cash ledger, pozisyonlar, karar turları ve özel notlar `wallet_id` ile bağlı tutulur; ilişkinin tutarlılığı backend'de denetlenir. Para/miktar hesapları decimal/sabit noktalı değerlerle yapılır; float yuvarlaması veya modelin ürettiği bakiye kaynak kabul edilmez. Açık emirlere ayrılan nakit izlenir. İstek tekrarına dayanıklı benzersiz emir anahtarı ve işlem sırasında hesap kaydı kilidi çift harcamayı önler. İş kuyrukları kalıcıdır; bir süreç yeniden başlayınca yalnız tamamlanmamış işler yeniden denenir. Kaynak: [PostgreSQL transaction içindeki satır kilitleri](https://www.postgresql.org/docs/current/explicit-locking.html).

**Web araştırması ve fiyat izleme:** Ortak veri toplama ve aktif koşullu emir fiyat işleri Go worker tarafından düzenlenir. Modelin web araması için canlı mod kullanılır; `web_search = "live"` resmen belgelenmiştir. Bir fiyat işinin model çağrısı mı yoksa seçilen ücretsiz sayfanın doğrudan okunması mı gerektirdiği, alternatif kaynak erişim testiyle netleştirilecek; web araması önbelleği bir dakikalık fiyat güncelliğini kanıtlamaz. Fiyat worker'ı kendi başına yeni strateji/alış kararı üretmez; mevcut koşullu emri kanıtla değerlendirir. Kaynak: [Codex canlı web araması](https://learn.chatgpt.com/docs/web-search).

**Basit ekran önerisi:** Ana sayfada sekiz cüzdan için toplam değer, kullanılabilir nakit, pozisyon değeri, net kâr/zarar ve son değerleme zamanı bulunur. Cüzdan ayrıntısında eldeki varlıklar, bekleyen/gerçekleşmiş emirler, masraflar ve karar gerekçeleri gösterilir. Veri/son iş durumu ile kullanıcıya özel al-tut kıyası ayrı görünür. İlk sürümde gözlem ekranı yeterlidir; para aktarma veya elle gerçek broker işlemi gibi yeni özellikler eklenmez.

**Kalıcı özel hafıza ve eşzamanlı abonelik oturumları — 8 Ekim 2026:** Kullanıcının istediği devamlılık için backend her cüzdanın hafızasını ayrı saklayacak. Buradaki hafıza model ağırlıkları veya ChatGPT hesabının genel kişisel hafızası değil, uygulamanın tuttuğu cüzdana özel kayıt ve çalışma bağlamıdır. Sekiz cüzdan aynı temel GPT modelini kullanabilir; sekiz bağımsız oturum/ajan olması sekiz farklı model eğitildiği anlamına gelmez.

| Hafıza katmanı | Tutulacak içerik | Kullanım |
|---|---|---|
| Sabit kimlik | Cüzdan karakteri, varlık kapsamı, talimat sürümü | Her turda değişmeyen rol ve yetki |
| Karar ve işlem geçmişi | Gönderilen girdiler, modelin yayımladığı yanıt/gerekçe, kaynak kimlikleri, yapılandırılmış emir, backend makbuzu, tur zamanları ve model sürümü | Denetlenebilir geçmiş; modelin erişilemeyen iç düşünce sürecini saklama iddiası yok |
| Çalışma hafızası | Önceki kararların kısa özeti, açık tezler, izlenecek olaylar, kendi sonuçlarına ilişkin ders/yöntem notları; sürüm ve dayandığı geçmiş kayıtlar | Sonraki tur için sınırlı, ilgili bağlam |
| Oturum eşlemesi | `wallet_id`, sağlayıcı, açık `session_id`, özel çalışma alanı ve son tamamlanan tur | Doğru oturumu devam ettirme ve toparlanma |

Nakit, pozisyonlar ve emir durumları özel hafıza özetinden hesaplanmaz; her yeni turda backend muhasebesinden güncel olarak alınır. Özetler geçmiş kayıtların yerine geçmez. Modelin dersleri/tezleri yorum olarak işaretlenir; doğrulanmış fiyat/haberle aynı alanlarda tutulmaz. Her tur sonrasında yalnız o cüzdanın yeni kayıtlarıyla sürümlü özet güncellenir; eski sürümler korunur. Tam geçmiş her çağrıya yığılmaz, ilgili kayıtlar gerektiğinde cüzdana yetkili okuma uçlarından alınır. Oturum geçmişinin eksiksiz biçimde her zaman bağlamda kalacağı varsayılmaz.

**Bağlantı akışı:** Güvenilen Go runner, sunucuda ChatGPT hesabıyla giriş yapılmış resmi Codex CLI'yi çalıştırır. İlk çalışmanın JSONL `thread.started` olayındaki kimlik ilgili cüzdana kaydedilir. Sonraki iş açık kimlikle `codex exec resume <SESSION_ID>` üzerinden sürdürülür; girdi/çıktı ve işlem makbuzları kendi backendimizce saklanır. CLI'nin yerel oturum dosyaları da o cüzdanın özel alanında kalıcı tutulur ve erişimi ayrılır; veritabanındaki ID tek başına oturumun tam yedeği değildir. Oturum kaybolursa backend geçmişi ve sürümlü özetiyle aynı cüzdan için yeni oturum oluşturulur, eski-yeni kimlik ve yeniden başlatma nedeni kaydedilir; birebir aynı model bağlamının geri geldiği iddia edilmez. Abonelik kimlik bilgileri yönetilen runner tarafındadır, hafıza tablolarına veya modelin okuyabildiği notlara yazılmaz. Kaynaklar: [abonelik kimlik doğrulama](https://learn.chatgpt.com/docs/auth), [JSONL çıktısı ve açık kimlikle devam](https://learn.chatgpt.com/docs/non-interactive-mode).

**Eşzamanlı yürütme önerisi:** Her günlük oturumda sekiz ayrı iş kuyruğa girer. Başlangıç için iki farklı cüzdanı aynı anda çalıştıran, ayarlanabilir bir Go worker havuzu önerilir; biten işin yerine sıradaki alınır. Bu sayı yeni bir işlem/risk sınırı değildir ve henüz kullanıcı kararı olarak kilitlenmedi. Aynı cüzdan için tek aktif iş/single-writer kuralı bulunur; farklı cüzdanlar paralel ilerleyebilir. Kuyruktaki iş çalışmaya başlarken güncel hesabı alır, emir öncesi fiyat doğrulamasını yeniden yapar; erken çalışan cüzdanın fiyatı sonrakine canlı fiyat diye taşınmaz. Giriş/kota/hizmet hatasında iş hata durumuyla kaydedilir ve uygun yeniden deneme yapılır; uydurma model yanıtı veya gerçekleşme üretilmez. Sekiz ayrı oturum hesabın kullanım hakkını sekize katlamaz. Resmi belgeler plan kullanımının paylaşıldığını belirtir, belirli bir hesabın sekiz eşzamanlı oturum kapasitesini garanti etmez; gerçek paralellik kapasitesi kurulumda ölçülecek. Kaynak: [Codex plan kullanım sınırları](https://learn.chatgpt.com/docs/pricing).

Yalnız farklı klasörler veya `wallet_id` alanları izolasyon kanıtı değildir. Modelin araçları diğer cüzdanların yerel CLI oturum dosyalarını, notlarını veya genel hafıza kaynaklarını okuyamamalıdır; özel çalışma alanı, oturum deposu ve backend yetkilendirmesi birlikte sınırlandırılacak. Ortak ham piyasa verisi paylaşılır; özel not/karar/sonuçlar paylaşılmaz. Bu bölüm planı ayrıntılandırır; kurulum, giriş veya çalışan paralellik testi yapılmadı.

### 0.18. Kapanış durumu

**Veri kapsamı, kaynak öncelikleri, web araştırma yöntemi, fiyat eşikleri, haber doğrulaması, gerekli kayıt alanları, iki rutin oturum, önemli olaylarda ek kontrol ve işlem öncesi yeniden fiyat doğrulaması planlama bakımından kapandı.** Bu, veri toplama uygulamasının tamamlandığı veya canlı kaynak erişiminin kanıtlandığı anlamına gelmez.

ABD hisse evreninin **50 sabit + 20 haftalık dinamik** yapısı, sabit 50'nin başlangıç isimleri ve haftalık güncelleme kuralı, dört kategori için **sekiz sanal cüzdan**, karakterlerin amaçları, **zorunlu tutma süresi olmaması**, cüzdan başına **100.000 USD**, özel bellek ayrılığı ve modele başarı/simülasyon bilgisinin verilmemesi kesinleşti. Önceki üç algoritma başlığı yerine karakter tanımları esas alınır; tutma süresini model seçer. Kendi Go backend'i ve basit web ekranı kabul edildi, chi seçildi; PostgreSQL/runner/ekran teknik mimari önerisi bölüm 0.17'dedir. Alternatif fiyat sitesi, sayısal simülatör/izleme ayrıntıları, kesin model talimatları ve son genel belge revizyonu halen tamamlanacak.

Kaynakların gerçekten okunabildiği ve üç fiyatın zaman koşullarını sağladığı kurulumda sınanacak; henüz bağlantı, çalışan simülatör veya performans sonucu yoktur. Önceki P1 taslağının sekiz ETF evreni, API veri edinimi, sabit cutoff ve deney konfigürasyonları yeni canlı pilot için kesinleşmiş seçim değildir.

Kaynak notları: [Nasdaq fiyat/seans açıklaması](https://www.nasdaq.com/market-activity), [Stock Analysis fiyat erişimi](https://stockanalysis.com/help/faq/stock-price-data/), [EWC fon bilgisi](https://www.ishares.com/us/products/239615/EWC), [NYSE seans/tatil takvimi](https://www.nyse.com/trade/hours-calendars), [ABD saat dilimleri](https://www.nist.gov/pml/time-and-frequency-division/local-time-faqs), [Codex canlı web araması](https://learn.chatgpt.com/docs/web-search). Listeye alınmak her kaynakta anlık erişim veya veri saklama/dağıtım hakkının doğrulandığı anlamına gelmez.

## Önceki P1 sayısal araştırma taslağı — 2 Ekim 2026

Aşağıdaki bölümler yöntem ve uzman değerlendirmesi geçmişini korur. Yeni pilotun bütçesi, evreni, zamanları veya uygulama sırası olarak doğrudan kullanılmayacak; son genel revizyonda güncel kararlarla uzlaştırılacaktır.

## 1. Hipotez, evren ve hedef

İlk soru: aynı bilgi ve tarihlerde **nominal %80 beş-seans getiri aralığı**, güçlü bir referanstan daha düşük uygun kayıp üretiyor mu? Yön üstünlüğü, ortalama tahmini ve kullanıcı değeri ayrı hipotezlerdir.

Önerilen sabit evren V1: SPY, IWM, EFA, EEM, IEF, TLT, GLD, VNQ. Her biri günlük skorda 1/8 ağırlıktadır. Liste veri kalitesi nedeniyle değişirse gerekçe ve V2 **değerlendirme öncesi** kaydedilir; kötü sonuç veren varlık atılmaz. Bugün seçilmiş kohortun sonucu tarihsel yatırım yapılabilir evrene genellenmez; ortak piyasa/kur/vade faktörleri bağımsız gözlem yaratmaz.

America/New_York takviminde cutoff 08:30, issue 08:45 ET; son geçerli yayın 09:00 ET. d1 issue'den sonraki ilk düzenli seans, d5 d1 dahil beşinci seans. OPEN/CLOSE düzenli seansın doğrulanmış ilk/son uygun bar fiyatıdır; resmî ihale veya emir gerçekleşme garantisi değildir.

$$
R_{i,o}=\frac{q_{i,d5}C_{i,d5}+D_{i,(d1,d5]}}{O_{i,d1}}-1,\qquad b_{i,o}=1[R_{i,o}>0].
$$

Split pay adedini değiştirir; brüt cash dividend hak edişi dahil, yeniden yatırım/vergi/komisyon/kayma hariçtir. d1 ex-date dağıtımı dahil değildir; d1'den sonraki ex-date'ler d5 dahil değerlendirilir. Ayrıntılı action, missing-label ve birim sözleşmesi MVP'dedir. Ondalık R, p_up, mean_simple_return, median_simple_return ve q10/q90 etiketleri karıştırılmaz. R=-1 toplam kayıp log hedefte sonlu değildir; simple-return skordan silinmez.

## 2. Ön kayıt ve deney bütçesi

Model karşılaştırması başlamadan protocol, V1, veri manifestleri, split tarihleri, feature schema, skor, baselines, dönüşüm/refit/calibration/purge/stop kuralları ve uygulama sürümleri hash'lenip dondurulur. Şu anda henüz bu hash veya kilit yoktur.

- **En çok 15 toplam kayıtlı konfigürasyon**, aşağıdaki R00–R04 ve C01–C10 dahil. Tablo üst sınırdır; her satırı çalıştırmak zorunlu değildir. Baselines ücretsiz deneme gibi sınırsız genişletilmez.
- Aynı joint çıktı şemasındaki önceden sabit q10/q50/q90 fit'leri bir vektör prosedürüdür; bütün alt-fit artifactleri kaydedilir. Başka quantile seviyesi, model/pencere/feature/calibration yöntemi bir yeni prosedürdür ve bütçeye girer.
- Seçim için denenen alpha, C, tree depth, 1/3/5-year window, soru/prompt, seed seçimi, event tanımı veya preprocessing değişimi de trial'dır. Tablo dışı varyant kullanılacaksa kullanılmamış satırın yerini **holdout açılmadan** kayıtlı gerekçeyle alır; eski deneme sicilden silinmez.
- Önceden tanımlı fold/refit, skor için bootstrap tekrarları ve sabit çıktı bileşenlerinin fit'leri yeni model araması değildir. Seed sabittir; seed taraması yapılmaz.
- İlk 15 girdinin dışına ekleme yeni protocol/konfigürasyon değişikliğidir. Holdout görüldükten sonra bu protokolle yeni araştırma onaylanmaz; o dönem development olur ve yeni ileri test gerekir.

## 3. Girdi şeması — en çok 15

Aşağıdaki fiyat girdileri önceki tamamlanmış günlerden, yalnız o cutoff'ta mevcut sürümden hesaplanır. Tarihsel uzun pencereler için availability bilinmiyorsa backfill kanıt sınıfı ayrıca raporlanır.

| No | Girdi |
|---|---|
| 1–4 | 1/5/20/60 seans geçmiş getiri |
| 5–7 | 5/20/60 seans geçmiş gerçekleşmiş log-return oynaklığı |
| 8 | Önceki tamamlanmış gündeki EWMA günlük volatilite |
| 9 | Önceki sağlayıcı günlük hacminin geçmiş 20-seans medyan hacme oranı |
| 10 | Split uyumlu fiyatın 20-seans ortalamasına göre uzaklığı |
| 11 | SPY'nin geçmiş 5-seans getirisi |
| 12 | SPY ile geçmiş 60-seans günlük log-return beta: sample covariance / SPY sample variance; sonlu ve variance>10^−12 şartı |
| 13–15 | d1–d5'te planlı FOMC kararı / CPI / Employment Situation açıklaması bayrakları |

Geçmiş günlük log getiri, önceki kapanışta elde tutulan bir payın o gün split ve hak edilmiş cash dividend dahil close-close servet oranının log'udur; yalnız cutoff'ta mevcut action sürümü kullanılır. L-seans getiri girdisi exp(sum_L daily_logs)−1, volatilite ddof=1 sample std'dir. Bu günlük-compounded geçmiş özellik, ana hedefin yeniden yatırımsız OPEN→CLOSE servet hesabıyla özdeş değildir. Feature12, feature2 ve feature11'in doğrusal fark kopyası yerine tarihsel beta'dır; nedensel etki diye yorumlanmaz.

Takvim bayrakları ancak ilan edilen takvimin geçmişte bilindiği kayıtla kanıtlanır; bugünkü gerçekleşen yayın tarihleri geçmişe taşınmaz. Hak/availability yoksa ilgili bayrak bütün modeller için pasif olur; bu karar skor görülmeden kaydedilir. Pilotta earnings bayrağı yoktur; şirket deneyinin cheap-event baseline'ına eklenir.

Asset kimliği gruplama/ölçek için metadata'dır; sınırsız learned asset embedding/one-hot etkileşim taraması yapılmaz. Pooled modelde feature sayısı 15'i aşmaz. Standardization, winsorization/imputation ve quantile-crossing onarımı yalnız izinli veride fit edilen/sabit prosedürdür. Eksik girdi sentinel'i fiyatın veya nötr haberin yerine geçmez.

## 4. Tek model kayıt tablosu

Tüm direct modeller R veya b hedefindedir. İlk eğitim penceresi 3 yıl, aylık refit; 3 yıl kullanılabilir olgunlaşmış örnek yoksa o origin warm-up sayılır. Eğitim penceresi sabit kalender aralığıdır; başka pencere seçimi bütçe tüketir.

| ID | Prosedür | Çıktı / kapsam |
|---|---|---|
| R00 | Sıfır basit getiri | Ortalama benchmark; p_up veya aralık üretmez. |
| R01 | Son 3 yıllık yalnız olgunlaşmış aynı-hedef ampirik dağılım | Mean, median, up-rate, q10/q90; rolling referans. |
| R02 | Expanding yalnız olgunlaşmış aynı-hedef ampirik dağılım | Aynı çıktı; nonstationarity kontrolü. |
| R03 | EWMA ölçeği, lambda=.94; aynı hedef geçmiş normalize getirilerin ampirik dağılımı | **Birincil aralık referansı**; aşağıdaki hedef adaptörü zorunlu. |
| R04 | Aynı EWMA/direct-target ölçek adaptörü, normal location-scale kontrolü | Basit parametrik aralık kontrolü; tam Gaussian cumulative log mean varsayılmaz. |
| C01 | Ridge, alpha=1 | Doğrudan simple-return conditional mean. |
| C02 | Ridge, alpha=10 | Tek kayıtlı daha güçlü shrinkage varyantı. |
| C03 | L2 lojistik, C=.1 | p_up; aşağıdaki sabit sigmoid calibration. |
| C04 | L2 lojistik, C=1 | p_up; aynı calibration. |
| C05 | QuantileRegressor, alpha=.001; tau=.10/.50/.90 | Doğrudan R quantile'ları, sabit artan sıralama onarımı. |
| C06 | Constant-mean GJR(1,1), normal innovations; direct-target residual scale adaptörü | Aralık challenger; daily variance conditioning olarak kullanılır. |
| C07 | Aynı GJR, standardized Student-t innovations, nu>2 | Aralık challenger; transformed mean üretilmez. |
| C08 | LightGBM joint quantile .10/.50/.90 | Sığ direct-target nonlinear challenger. |
| C09 | LightGBM binary | p_up; aynı sigmoid calibration. |
| C10 | LightGBM L2 regression | Doğrudan R mean. |

LightGBM ortak sınırı: 200 ağaç, learning_rate=.03, max_depth=3, num_leaves=7, min_data_in_leaf=100, feature/bagging fraction=1, sabit seed. Otomatik early-stopping araması yapılmaz; sürüm ve solver ayrıntıları score açılmadan config artifactinde sabitlenir. C05/C08'de quantile sıralama onarımı her origin'de aynı uygulanır ve crossing sayısı raporlanır.

Ridge/lojistik ilk challenger'dır. GJR/boosting yalnız baseline+harness çalışıp güç raporu çıktıktan sonra denenir. GJR convergence/variance şartı veya hedef adaptörü geçmezse C06/C07 failed/deferred olur, başka uyumsuz hedefle puanlanmaz.

**Hedef adaptörü:** Beş close-close return toplamı OPEN(d1)→CLOSE(d5) hedefi değildir. Origin o'da location μ_i,o, önceki 3 yılın o an uygun/olgunlaşmış direct-target R değerlerinin ortalamasıdır; minimum 252 etiket gerekir. R03/R04'te s_i,o=sqrt(v_EWMA,i,o), C06/C07'de s_i,o=sqrt(sum_k=1…5 h.k_i,o). İkinci ölçek yalnız conditioning girdisidir; direct-target variance diye adlandırılmaz.

Geçmiş origin j'nin sonucu uygun hale gelince z_i,j=(R_i,j−μ_i,j)/s_i,j kaydedilir. μ_i,j ve s_i,j o origin'deki nedensel state/model artifactinden gelir; bugünkü modelle geçmiş yeniden fit edilmez. Son 3 yılda varlık başına en az 250 uygun residual zorunludur. Ölçek sonlu ve >10^−8 değilse origin invalid_scale olur; clipping ile ölçek uydurulmaz.

R03 ve C06/C07: q_tau,i,o=μ_i,o+s_i,o times empirical_quantile_tau(z_i,j). Quantile, sıralı residual'larda inverse-empirical-CDF, indeks ceil(n*tau) (1 tabanlı) olarak sabittir. R04: aynı residual pool'un arithmetic mean ve ddof=1 sample standard deviation'ıyla normal quantile; q=μ+s*(mean_z+std_z*Phi_inverse(tau)). Std_z>10^−8 ve sonlu değilse normal kontrol geçersizdir. Adaptör pool'u yeni olgunlaşmış sonuç geldikçe güncellenir; yalnız past-eligible sonuç girer. Daily EWMA her tamamlanmış önceki gün güncellenir, GJR ayın ilk seansında refit edilir.

τ=.10/.50/.90 için üretilen bütün R quantile'ları önce max(−1,q) support düzeltmesinden, sonra artan sıralama onarımından geçer. R mean tahmininde de sabit −1 alt desteği uygulanır; yapılan düzeltme sayısı raporlanır. Bu dönüşümler config'nin parçasıdır; holdout'ta eklenmez. Gaussian/Student-t günlük log path'den direct R mean/quantile dönüşümü yapılmaz. Açık overnight/intraday path alternatifi yeni deneydir.

Kalman/HMM, AR(1)/ARMA, ElasticNet, conformal, ensemble, analog retrieval, hierarchical Bayes, relative-return ranker ve gelişmiş ağlar bu tablonun parçası değildir. Matematik kataloğunda bulunmaları geliştirme taahhüdü sayılmaz.

## 5. Tarih ayrımları ve erişim koruması

Önerilen tarih planı, izinli/güvenilir 2016 sonrası veri varlığına bağlıdır. Kapsam yetersizse yalnız kalite/coverage bilgisine bakılarak **model skorları açılmadan** bir P2 hazırlanır. Holdout kısaltılıp sonuç güçlüymüş gibi sunulmaz.

| Bölüm | Forecast-origin tarihleri | Kullanım |
|---|---|---|
| Development | 4 Ocak 2016–30 Aralık 2022 | Warm-up sonrası model/özellik seçimi. Yıllık walk-forward skor fold'ları 2020, 2021 ve 2022. |
| Calibration | 3 Ocak–29 Aralık 2023 | Gerçek out-of-sample skor/residual'lardan sabit sigmoid ve hedef/aralık adaptörü doğrulaması. Model menüsü seçmek için kullanılmaz. |
| Kilitli holdout | 2 Ocak 2024–30 Haziran 2026 | Tek açma; ana etki ve ikincil kayıtlı iddialar. Etiketler olgunlaştıktan ve veri yeterliliği doğrulandıktan sonra. |
| Reserve / ileri başlangıç | 1 Temmuz 2026 sonrası backfill ve çalışan günlüğün başlangıcı | Backfill, o tarihte üretilmiş ileri tahmin değildir. Gerçek forward günlüğü sistem hazır olunca başlar; başlangıç manifesti kaydedilir. |

Son holdout origin'i d5 etiketiyle birlikte değerlendirilir; “2026 ortasına kadar” ifadesi label end'i kesmez. Sonuçların 15 Temmuz 2026 08:30 ET'ye kadar olgunlaştığı varsayımı veri aşamasında kontrol edilir; missing action/data hâlâ pending olabilir.

İki veri-zaman modu ayrı tutulur. **Gerçek forward veya kanıtlı historical vintage** modunda matured_at, d5 fiyat/olay verilerinin gerçekten mevcut ve doğrulanmış olduğu zamandır; matured_at ≤ fit_cutoff zorunludur. **Exploratory revised-backfill replay** modunda actual capture/matured_at bugün olarak korunur; ayrı replay_eligible_at_assumed alanı, d5'i izleyen ilk düzenli seansın 08:30 ET cutoff'u olarak atanabilir. Bu konservatif replay varsayımıdır, eski sürüm/erişim kanıtı değildir. Pending action varsa varsayımsal etiket de uygun değildir. Feature replay varsayımı ve label replay varsayımı ayrı manifestte görünür; böyle bir testin p-value/holdout başarısı gerçek point-in-time üstünlük kanıtı sayılamaz.

Aynı origin tarihi bütün varlıklarda aynı bölümdedir. Fit'te moda uygun actual veya açık assumed eligibility zamanı kontrol edilir; d5 geçti diye güncel düzeltme o an mevcut sayılmaz. Eğitim label interval'i validation/calibration/test bilgi sınırıyla kesişirse örnek purge edilir; calibration sonundaki etiketler holdout'a sarkamaz. Panelde gap=5 satır kullanmak beş seans purge değildir; interval/date-aware splitter gerekir. [TimeSeriesSplit tanımı](https://scikit-learn.org/stable/modules/generated/sklearn.model_selection.TimeSeriesSplit.html).

Frozen refit politikası: ayın ilk seansında rolling 3-yıl uygun geçmiş verisiyle refit. Daha önce olgunlaşmış holdout sonuçlarını yalnız bu mekanik sequential kural kullanabilir; araştırmacı skor görüp değiştiremez. C03/C04/C09 için raw decision_function/logit → sigmoid calibration iki katsayısı, yalnız 2023 OOS tahminlerinden fit edilir ve holdout/ilk forward protokolü boyunca **sabit kalır**. Kalibratör refit edilmez; ölçek residual pool'unun nedensel güncellenmesi bununla karıştırılmaz. R01–R04, mean ve direction reference'lar origin'de sadece geçmiş uygun etiketlerle güncellenir. Statik base-model veya farklı calibration/refit kuralı yeni deneydir.

Holdout data/labels uygulamada ayrı erişim alanında tutulur; tuning komutu bu bölümü yükleyemez. Model/protocol/input manifest hash'leriyle tek logged **unseal** adımı değerlendirmeyi açar. Bu klasörde henüz böyle bir yazılım kilidi yoktur; belge tek başına saklı test garantisi değildir. Tarihleri bilmek sızıntı değildir; skor/etiket incelemesiyle tasarım değiştirmek holdout'u tüketir.

## 6. Ölçütler, güç ve istatistik

Birincil günlük kayıp, aynı tarihteki geçerli 8 varlığın eşit ağırlıklı ortalama quantile kaybıdır:

$$
L_o=\frac18\sum_i\frac{\rho_{.10}(R_{i,o}-\hat q_{.10,i,o})+
\rho_{.90}(R_{i,o}-\hat q_{.90,i,o})}{2},\qquad
\rho_\tau(u)=u(\tau-1[u<0]).
$$

Ana karşılaştırma geliştirmede seçilmiş **tek aralık challenger** ile R03'tür. Ayrıca R01/R02/R03/R04'ten geliştirmede en düşük pinball kaybına sahip olan **guard reference** sabitlenir. Challenger her iki referansa karşı raporlanır; daha zayıf R03'ü geçip güçlü guard'ı geçmezse üstünlük kapısı açılmaz. Seçim holdout'tan yapılmaz. Development'ta tek yön ve tek mean finalist de kendi ana kaybıyla seçilir; 2023 calibration veya final skoruyla finalist değiştirilmez.

Göreli iyileşme g=1−mean(L_challenger)/mean(L_reference). %2 önerisi bu revizyonun teknik planning threshold'udur; kabul edilmeden kullanıcı değeri sayılmaz. İstisnai sıfır baseline loss'ta oran tanımsızdır; absolute difference raporlanır ve oran kapısı sonuçsuz olur.

Günlük paired loss farkında bütün varlıklar aynı blokta tutulur. Stationary/moving date-block bootstrap için ana ortalama blok 20 seans, duyarlılık 5/10/40 seans; sabit seed ve en az 5.000 tekrar. Oran CI'da iki loss aynı bloklarla birlikte resample edilir. Sadece günlük satırları bağımsız bootstrap yapmak yasaktır.

Development'ta minimum detectable effect (MDE), effective dependence ve erişilebilir holdout/forward uzunluğu için ilk baseline feasibility raporu **model genişletmeden önce** çıkarılır. Henüz challenger loss'u olmadan onun kesin gücü bilinmez; finalist seçildikten sonra development paired-loss serisiyle güç/MDE yeniden hesaplanıp **holdout açılmadan** dondurulur. %80 hedef güç ve aile düzeyinde %5 hata sınırı için block-resampled paired loss sürecine farklı etki büyüklükleri uygulanır; power curve ve küçük örneklem duyarlılığı gösterilir. Bunu “8 ETF × tarih sayısı” ile hesaplamak geçersizdir. Bootstrap gücü stationarity varsayımına duyarlıdır; bir kesin gelecek garantisi değildir.

Tek seri iid korelasyon örneği için Fisher dönüşümü yaklaşık tanh((1.96+.84)/sqrt(n−3)) verir: 26/150/500/1.000 bağımsız dönemde yaklaşık .53/.23/.125/.089. rho=.05 yaklaşık 3.140 dönem, yılda 50 dönemle 63 yıl gerektirir. Bu, ikinci uzmanın küçük yön etkisi uyarısını destekler; **pinball gücü, gerçek ETF paneli veya evrensel effective-N hesabı değildir**. Beş günlük örtüşme, ortak faktör ve nonstationarity gerçek loss üzerinde ölçülür.

| Çıktı | Ana ölçüt / sınır |
|---|---|
| q10/q90 aralığı | Birincil mean pinball; nominal %80 coverage, width ve interval score birlikte. |
| Medyan | MAE; ortalama model seçiminde ana skor değildir. |
| Ortalama R | MSE/RMSE ve historical rolling mean'e göre OOS R²; transformed Student-t mean yok. |
| p_up | Brier; log loss ve Brier skill ayrıca. Rolling/expanding matured-label up-rate kontrolü. |
| Olasılık kalibrasyonu | Calibration intercept/slope + blok CI, reliability grafiği; accuracy/Brier tek başına kalibrasyon kanıtı değildir. |
| Volatility | Ayrı tanımlı gerçekleşmiş variance hedefinde QLIKE; direct-target aralık kaybıyla karıştırılmaz. |
| Operasyon | Planlı, zamanında, failed, late, fallback ve pending-label oranları. |

Toplu aralık coverage hedef bandı %75–85, asset başına <65% ciddi uyarıdır; CI ile birlikte yorumlanır. Geniş aralık yalnız coverage'ı yükselterek başarı kazanamaz. Önceden seçilmiş non-overlapping d1→d5 alt örneklemde coverage ve bağımlılık kontrolü yapılır. Kupiec/Christoffersen uygun koşullarda diagnostic olabilir; düşük güçte kabul edilen coverage null'ı doğruluk kanıtı değildir.

DM/HAC kayıp farkı testi lag en az 4, duyarlılık 10/20 ve gerektiğinde daha uzun veri-türevli lag ile ek kontrol olabilir; iid test veya sabit lag=4 her bağımlılığı çözmez. Rejim/asset subgroup sonuçları ayrıca ön kayıt ve güç yoksa diagnostics'tir.

Aile: primary vs R03 ve guard, yön Brier ve mean MSE iddiaları birlikte **Holm düzeltmesi**. Her kayıtlı tek-yönlü superiority p değeri için günlük loss farkı d_o=L_reference−L_candidate, centered-null serisi d_o−mean(d) ve aynı date-block resampling kullanılır; p=(1+null_bootstrap_mean≥observed_mean sayısı)/(B+1). Center edilmemiş bootstrap pozitif fark oranı p-value diye sunulmaz. Bu mixing/stationarity altında yaklaşık bir testtir; küçük sample/block sensitivity raporlanır. İddia test edilmediyse family planı sonuca bakarak küçültülmez.

Development search sicili bütün 15 konfigürasyonu içerir; fold/konfigürasyon bazında çok sayıda p-value başarı diye seçilmez. Tek seçilmiş finalist korunmuş testte açılır; Holm, testten sonra seçme veya sızıntıyı onarmaz. Romano–Wolf yeni yöntem olarak ancak önceden tanımlanır. [Backtest selection riski](https://www.davidhbailey.com/dhbpapers/overfitting.pdf).

Olasılık ile quantile aynı dağılımdan geliyorsa tutarlılık zorunludur; örneğin p_up=.80 ve q90<0 birlikte geçerli olamaz. Ayrı modellerin çıktıları otomatik tek dağılım diye birleştirilmez. q-crossing onarımı, R≥−1 desteği ve CDF/mean uyumu kontrol edilir; destek clipping veya joint-distribution seçimi önceden config'de sabitlenir ve kendi etkisi ölçülür.

## 7. Geçiş koşulları ve durdurma

Aşağıdaki sayılar **önerilen pilot koşullarıdır**. Kullanıcı değeri ve development güç raporu sonucu değiştirilmesi gerekirse holdout açılmadan P2 ile değiştirilir. Sonuç görüldükten sonra eşik indirilmez.

| Koşul | Geçiş | Başarısız / sonuçsuz |
|---|---|---|
| G0 — amaç/haklar | Özel araştırma amacı ve her kullanılan kaynakta işleme/arşiv/ML hakkı belgeli; target V1 sabit. | Belirsiz kaynak kolu engellenir; ücretsiz erişim yeterli değil. |
| G1 — veri | Kohortun işlemde olduğu beklenen seanslarda ≥%99,9 kayıt; eksikler açıklanmış; aynı hedefte ikinci kaynakla getiri farkı ≥%99 gün ≤5 bp ve açıklanamayan >50 bp fark yok; dahil tüm action'lar uzlaştırılmış; cutoff/session/dividend boundary testleri geçiyor. | İkinci kaynak veya eşdeğer tanım yoksa doğrulama tamamlanmamış. Çözümsüz material revision/action/timestamp varsa ilgili etiket/kol bloke. |
| G2 — ölçüm/güç | Baselines, exact purge, matured-label fit, input manifests ve holdout erişim testi geçiyor; adaptör frozen; MDE/power raporu tamam. | Güç %2 planning etkisine yetmiyorsa küçük üstünlük iddiası açılmaz; yalnız exploratory/operasyon sürer veya bağımsız yeni hedef tasarlanır. |
| G3 — sayısal aday | Seçilmiş challenger vs R03 nokta iyileşmesi ≥%2, paired-date-block %95 CI altı >0 ve Holm testi p<.05; guard'a karşı da CI altı >0 ve Holm p<.05; nominal coverage %75–85 ve asset <65% uyarıları çözülmüş, numeric sadakat geçiyor. Aynı R03 guard seçildiyse bu tek test sayılır. | CI hem sıfır hem anlamlı faydayı kapsıyorsa **sonuçsuz**. Fayda üst sınırı %2'nin altındaysa planning hedefi desteklenmiyor. Açık zarar/hata: no-go. As-of unknown backfill sonucu yalnız exploratory aday seçimi, confirmatory geçiş değil. |
| G4 — ileri kayıt | 60 seans boyunca bütün scheduled işler kayıtlı ve zamanında başarı ≥%99; hash/cutoff/label kontrolleri geçiyor. En az 26 haftada skor/coverage drift, failures ve holdout CI ile karşılaştırma raporu. | Bu süreler operasyon sağlar; küçük edge'i kanıtlamaz. MDE'nin gerektirdiği yeterli forward kanıt olmadan ürün üstünlüğü iddiası yok. |
| G5 — metin/Jev | Ayrı event-study amaç, kaynak hakkı, sürüm/cutoff kanıtı, power, sabit sorular ve finance annotation seti; metadata/cheap-text kontrollerine karşı ön kayıtlı incremental fayda ve temiz forward kanıt. | 5–10 hisse veya contaminated backfill ile finansal üstünlük geçişi yok. Güçsüz/katkısızsa katman kaldırılır; numerical pilot devam edebilir. |
| G6 — açıklama/ürün | Sayısal kalite ve kullanıcı faydası gerekçesi var; gerekli dış kullanım hakları/düzenleyici inceleme tamam. Template'e karşı LLM testi: otomatik numeric consistency %100, denetlenen iddialarda ≥%95 kaynak desteği ve ölçülebilir comprehension/factuality faydası. | İkna edici yazı tek başına fayda değil; başarısız LLM yerine template. UI/API yalnız somut kullanıcı ihtiyacı varsa. |

G3, “iyileşme sıfırdan farklı ve nokta tahmini en az %2” için aday kanıtıdır; **“en az %2 faydayı kanıtladık” değildir**. İkinci iddia için CI altının %2'yi geçmesi ve buna uygun daha güçlü ön kayıt gerekir. Eşikleri sağlamak veri-vintage bilinmezliğini veya gelecekteki nonstationarity'yi çözmez.

p_up için ayrı devam şartı: ön kayıtlı strongest up-rate baseline'a göre en az %1 relative Brier iyileşme noktası, blok CI altı >0, aile-düzeltilmiş test ve calibration bozulmaması. Mean için kendi MSE/OOS R² iddiası ve güç şartı aranır. Bu ikincil teknik eşikler de kullanıcı değeri değildir; başarısız başlık yalnız labelled reference olarak gösterilebilir.

**Durdurma:** Bütçe dolduğunda yeni model menüsü açılmaz. G3 negative ise ilgili tahmin iddiası reddedilir; inconclusive ise üstünlük söylenmez ve ilave bağımsız veri/gözlem maliyeti değerlendirilir. Yön/mean başarısızsa bu çıktılar için product claim durur; yalnız intervals güçlü ve kullanıcıya yararlıysa ayrı belirsizlik ürünü değerlendirilebilir. G2 güçsüzken daha karmaşık model eklemek çözüm kabul edilmez. Başarısız mutlak yönü göreli getiriye çevirip aynı holdout'ta yeni zafer aranmaz.

İleri operasyon ayda bir incelenir; ≥60 olgunlaşmış origin'den sonra 60-origin rolling loss/coverage güven bandı iki ardışık incelemede baseline'a karşı maddi bozulma gösterirse model askıya alınır ve labelled baseline'a döner. Bu alarm kanıtlanmış istatistiksel üstünlüğün ters testi değildir; bağımlı sequential diagnostics'tir. Material veri/sızıntı/schema hata beklemeksizin askıya alır; yeniden eğitim/rollback yeni sürümle izlenir.

## 8. Yayınlanmayan tahminler ve sonuç kaydı

Bütün scheduled asset/origin'ler ledger'dadır. Structural veri eksikliği ve ortak warm-up maskesi modelden bağımsız olarak score görülmeden belirlenir; minimum 252 label/250 origin residual gereksinimi bütün karşılaştırma adaylarında aynı eligible-score başlangıcına uygulanır. Ana paired karşılaştırmada aynı origin/varlıklar kullanılır; eksik asset olan origin ana eşit-8 serisine girmez ve oranı ayrıca raporlanır. Bir adayın hatalı/missing output'u tarih seçmek için kullanılmaz; G4/başarı oranını düşürür, karşılaştırma eksikse confirmatory sonuç verilmez.

İleride registered fallback'li deployment policy değerlendirilirse bu **ayrı isimli politika** olarak skora girer; raw challenger başarısı gibi sunulmaz. Abstention ayrı konfigürasyondur; coverage-matched baseline, risk–coverage eğrisi ve bütün yayımlanmayan origin sonuçları gerekir. High-volatility günleri elemek bağımsız tahmin kazancı değildir.

Forecast immutable; outcome ayrı sürümlenir. Label ilk ne zaman olgunlaştı, action bekledi mi ve düzeltme hangi skor sürümünü etkiledi kayda girer. Yerel hash zinciri yanında bağımsız gün sonu hash timestamp yoksa tarih kanıtı eksikliği raporlanır. Baseline forward ledger'ı Jev/UI beklemeden G1/G2 hazır olunca başlatılır.

## 9. Sonraki araştırmaların sınırı

Haber/Jev study, event sonrası **ilk erişilebilir fiyat** ile başlayan ayrı abnormal/relative beş-seans hedefini önceden tanımlar. Benchmark beta/event-window seçimi eğitimde yapılır; yalnız gerçekten görülmüş gövde sürümleri, frozen model ve olgunlaşmış sonuçlar kabul edilir. Cheap event/calendar ve cheap text zorunlu kontrollerdir; anonim şirket koşulu sensitivity'dir. Jev finance annotation doğruluğu downstream finansal başarıyla aynı değildir.

Retrieval için numerical/BM25 shortlist → aynı shortlist Jev semantic rerank ablation; fiyat/tarih uzaklığı kodda, outcome-blind seçim ve matured/deduplicated adaylar. Nadir olaylarda Kish weight ESS gerçek bağımsız örnek sayısı sayılamaz.

LLM aynı numerical forecast üzerinde template ile karşılaştırılır. Numbers/source consistency, reviewer audit, comprehension ve over-trust ölçümü yapılır. Forecast değişikliği, model seçimi veya abstention yeni deneydir; unchanged numbers için Brier farkı beklenmez. API/version/prompt/request-response hash ve hata/cache sözleşmesi MVP'dedir.

Ensemble, CQR/ACI/EnbPI, HMM/Kalman, OU/spread, relative-return ranker, foundation modeller ve RL için ayrı hipotez, kaynak, güç ve bütçe gerekir. Bu katalog sonraki fazın otomatik iş listesi değildir. [Matematik araştırması](matematik-algoritmalari.md), [repository araştırması](github-trading-sistemleri.md).
