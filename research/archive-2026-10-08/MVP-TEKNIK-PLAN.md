# ABD Hisse ve ETF Tahmini — Araştırma Pilotu ve Teknik Plan

Güncelleme: 8 Ekim 2026. Durum: teknik mimari planlanıyor; kendi Go backend'i ve basit web ekranı seçildi, kodlama ve kurulum henüz başlamadı.

## Güncel kararların önceliği

Güncel kullanıcı kararları [deney planının bölüm 0'ında](research/deney-plani.md#0-güncel-kullanıcı-kararları--canlı-pilot) tutulur ve aşağıdaki 2 Ekim taslağıyla çelişirse önceliklidir. Yeni canlı pilot ABD hisseleri, QQQ, TQQQ ve EWC kategorilerini kapsar; ana erişim GPT aboneliği, fiyat edinimi ücretsiz canlı web araması/ürün sayfası okumasıdır. Veri sağlayıcısı API anahtarı veya ek veri aboneliği kullanılmayacak. Alım/satım adayı üç farklı siteden doğrulanacak; en yüksek/en düşük fiyat açıklığı ortanca fiyata göre en fazla **%1**, kaynak fiyat zamanlarının açıklığı en fazla **60 saniye**, piyasa açıkken her fiyatın kontrol anındaki yaşı en fazla **60 saniye** olacak. Ayrıntılı kaynak listesi, hesap ve hata kuralları bölüm 0'da tek kaynakta tutulur.

Fiyat, haber ve diğer bilgi toplama kapsamı kesinleşti. Kategori bazında kaynak öncelikleri, arama sorguları, iki bağımsız kaynak/tek kaynak/çelişki kuralları, beklenti–gerçekleşen ayrımı ve asgari kanıt alanları deney planının 0.6–0.9 bölümlerine kaydedildi. Modelin hafızası doğrulama kaynağı değildir; aynı bültenin iki kopyası iki bağımsız kanıt sayılmaz. Dört kategoride işlem günü başına iki rutin araştırma/karar oturumu vardır: **açılıştan 30 dakika sonra** (kullanıcının 15–30 dakika aralığından seçildi) ve **kapanıştan 1 saat önce**. Saatler gerçek borsa takviminden türetilir ve GMT+3 gösterilir; önemli olaylarda ek bilgi kontrolü ve her alım/satım öncesi yeniden üç kaynaklı fiyat doğrulaması yapılır. Ayrıntılar bölüm 0.10'dadır. Canlı kaynak erişimi kurulumda sınanacak; sıradaki planlama başlığı evren/ufuktur. Diğer maddeler dört kategori için ayrı ayrı netleştirilecek. Kodlama öncesi son genel revizyon yapılacak. Aşağıdaki sekiz ETF/API tabanlı sayısal taslak yöntem geçmişidir; yeni pilotun onaylanmış uygulama planı değildir.

ABD hisse kategorisinde **50 sabit + sabit liste dışından modelin seçtiği 20 dinamik hisse** izlenecek. Dinamik liste haftalık güncellenebilir, günlük değiştirilemez. ABD'de ortak 70 hisselik evrenle **üç ayrı sanal cüzdan** çalışacak: hızlı kazanç arayan, olay odaklı dengeli ve orta–uzun vadeli sabırlı karakter. QQQ ve TQQQ'da ayrı ayrı hızlı ve uzun vadeli iki karakter, EWC'de bir dengeli karakter vardır: **toplam sekiz cüzdan**. Zorunlu tutma/satış süresi konulmaz; model karakterin amacıyla süreyi seçer. Başlangıç sermayesi ve maliyet varsayımları karşılaştırmada aynı tutulacak; risk sınırları ayrıca belirlenecek. Ortak veri havuzundan genel tarama yapılır, önemli gelişmeler ve işlem adaylarında derinleşilir. Evren ve karakter ayrıntıları deney planının 0.11–0.12, kalan karar listesi 0.13 bölümlerindedir. Sabit 50'nin isimleri/seçim ölçütleri ve haftalık değişim zamanı henüz belirlenmedi. TQQQ'nun günlük üç kat hedefi çok günlük getirinin üç katı gibi simüle edilmeyecek.

En yeni kararlar deney planının **0.14–0.15** bölümlerindedir ve yukarıdaki önceki açık varsayımların önüne geçer: her cüzdan **100.000 USD** ile başlar; ortak kaynak veritabanına rağmen **sekiz özel hafıza/bağlam birbirinden ayrıdır**. Modeller diğer cüzdanlar, simülasyon ve araştırmacının %10 referansı hakkında bilgilendirilmez. **Yalnız öz parayla nakit hesap işlemi yapılır; borç, faiz, margin, açığa satış ve türev sözleşmeler kapsam dışıdır.** Kullanılabilir nakit işlem ve masrafları karşılamalıdır; model fırsatlar için nakit ayırmayı kendi seçer, zorunlu nakit yüzdesi yoktur. Ek stratejik risk limitleri konulmayacak; maliyet, nakit/takas, işlem saati ve veri doğrulama koşulları uygulanacak. **Para ekleme, çekme ve cüzdanlar arası aktarım yoktur. ETF cüzdanları yalnız kendi ETF'lerinde ve nakitte kalır.** İki karar oturumunda emir verilir; arada koşullu emirler yalnız normal seansta çalışabilir. Karakter sabit, karakter içindeki yöntem geliştirmeye açıktır. Broker maliyetleri ve ilgili temettü stopajı dahil, kişisel gelir vergisi ayrıdır; stopaj oranı henüz kilitlenmedi. Pilotun sabit bitiş süresi yoktur. Sabit 50 hisse ve ilk işlem günü 09:00 New York haftalık güncelleme kuralı bölüm 0.11'de somutlaştırıldı. İşlem yetkisi sorusu kapandı; ekran ayrıntıları, hafıza teknik yapısı, altyapı ve son genel revizyon sonra konuşulacak.

Kapsam kontrolü sonrası tamamlamalar deney planının **0.16** bölümüne kaydedildi: alternatif ücretsiz fiyat sitesi araştırılacak; aktif koşullu emirler için normal seansta yalnız ilgili sembollerin fiyatını kontrol eden ayrı izleme kurulacak; dinamik listeden çıkan eldeki pozisyon izlenip satılabilecek; emir yanıtları backend gerçekleşme/cüzdan kayıtlarıyla tutarlı olacak. Al-tut kıyasları yalnız araştırmacı değerlendirmesinde yer alacak. Yeni fiyat sitesi ile izleme sıklığı ve sayısal simülatör kuralları henüz belirlenmedi; kullanıcı simülatör ayrıntılarını sonraya bıraktı.

## Güncel teknik mimari önerisi — 8 Ekim 2026

Kullanıcının kararıyla backend **Go ile tarafımızdan yazılacak**, hazır trading API/backend ürünü kullanılmayacak. HTTP katmanı için **chi v5** seçildi. Önerilen başlangıç yapısı tek Go uygulaması + PostgreSQL + resmi Codex CLI runner'dır; web gözlem ekranı Go HTML şablonları ve küçük JavaScript ile aynı uygulamadan sunulabilir. PostgreSQL, dağıtım ve diğer teknik ayrıntılar bu aşamada mimari önerisidir. Ayrıntılar ve kaynaklar [deney planının 0.17 bölümündedir](research/deney-plani.md#017-kendi-go-backendimiz-ve-web-ekranı--8-ekim-2026).

Go uygulamasının modülleri: kendi HTTP API'si, cüzdan muhasebesi, emir/doğrulama ve gerçekleşme motoru, kalıcı iş kuyruğu/zamanlayıcı, ortak web veri toplama, koşullu emir fiyat kontrolü, Codex çalışma bağlamlarının yürütülmesi ve kullanıcı ekranı. İş kuralları kendi kodumuzdadır. Para/pozisyon muhasebesi backend'in kaynağıdır; modele DB veya yönetici erişimi verilmez.

Her cüzdanın ayrı session ID, çalışma alanı ve özel hafızası bulunur; ortak veri olguları paylaşılır. Go runner ilgili özel bağlamla Codex oturumunu çalıştırır, yapılandırılmış kararları cüzdana yetkili kendi emir API'mize iletir ve sonuçları aynı oturuma döndürür. CLI'yi açık session ID ile sürdürme ve ChatGPT abonelik kimliğiyle çalıştırma resmen belgelenmiştir; kendi sunucumuzda gerçek bağlantı/izolasyon henüz sınanmadı. Paylaşılan `--last` oturumu kullanılmayacak. Tekrar gönderilen emir aynı kalıcı sonuçla eşlenir; para hareketi, pozisyon ve gerçekleşme transaction içinde tutarlı tutulur.

Web ekranının ilk kapsamı gözlemdir: sekiz cüzdanın toplam değeri, nakit/pozisyonları, kâr/zararı, son fiyat zamanı; cüzdan bazında emir geçmişi ve gerekçeler; kullanıcıya özel al-tut kıyasları. Kaynak seçimi ve kontrol sıklığı, simülatör sayısal kuralları, kesin prompt/çıktı şeması, sunucu/dağıtım ve son genel revizyon halen tamamlanacak. Aşağıdaki eski taslak yeni Go mimarisinin uygulama planı değildir.

### Cüzdan hafızası ve paralel GPT oturumları

Backend sekiz cüzdanın özel hafızasını ayrı tutar: sabit karakter/talimat sürümü, gönderilen girdiler ve yayımlanan yanıtlar, karar/emir/makbuz geçmişi, dayandığı kaynaklar, sürümlü çalışma özeti ve oturum kimliği. Model ağırlıkları veya hesabın genel kişisel hafızası bu uygulama hafızası değildir. Bakiyeler ve pozisyonlar her turda backend muhasebesinden alınır; model özeti hesap kaynağı olmaz. Ders/tez notları yorum olarak korunur; geçmiş kayıtları silmez. Ayrıntı ve belge kaynakları deney planının 0.17 bölümündedir.

Go runner resmi Codex CLI'nin abonelikle kayıtlı kimliğini kullanır. İlk JSONL çıktısındaki `thread_id` cüzdana bağlanır, sonraki turlar açık `session_id` ile devam eder. Her cüzdanın yerel oturum deposu ve çalışma alanı diğerlerinden erişim bakımından ayrılır; yalnız ID saklamak tam oturum yedeği değildir. Oturum kaybında aynı cüzdanın backend geçmişi/özetiyle yeni oturum oluşturulur, değişim kaydedilir. Kimlik bilgileri model notlarından ve hafıza tablolarından ayrı yönetilir. Gerçek sunucu entegrasyonu henüz sınanmadı.

Sekiz ajan aynı temel modeli kullanabilir, ayrı karakter ve geçmişlerle bağımsız çalışır. Günlük sekiz iş için başlangıçta **iki paralel cüzdan işi** yürüten ayarlanabilir worker havuzu önerilir; sayı henüz kilitlenmedi. Aynı cüzdanda tek aktif iş olur, farklı cüzdanlar paralel çalışabilir. Her iş başladığında güncel hesabı alır, her emir öncesi fiyatı yeniden doğrular. Hesabın kullanım hakkı ortaktır; sekiz oturum sekiz ayrı kota yaratmaz. Belgeler sekiz eşzamanlı çalışmayı hesap bazında garanti etmez; kurulumda kapasite ölçülür. Hatalar kaydedilir ve işler uygun biçimde yeniden denenir; aynı emir idempotency anahtarıyla ikinci kez gerçekleşmez.

## Önceki teknik taslak — 2 Ekim 2026

İlk teslimat küçük, yanlışlanabilir bir **sayısal araştırma pilotudur**: veri, güçlü referanslar, korunan değerlendirme ve ileri tahmin günlüğü. Hibrit ürün veya tahmin üstünlüğü varsayılmaz. İki uzman görüşünden yapılan seçimler [karar raporunda](KAPSAMLI-ARASTIRMA-RAPORU.md); deney kimlikleri, tarihler ve kabul koşulları için tek kaynak [deney protokolüdür](research/deney-plani.md).

## 1. Karar ve kullanıcı sözleşmesi

| Alan | Pilot kararı ve sınırı |
|---|---|
| Kullanıcı | Varsayılan: proje sahibi, özel ve kişisel araştırma. İç şirket kullanımı veya müşteriye sunum farklı sözleşme gerektirir. |
| Desteklenen karar | Beş seanslık getiri belirsizliği hakkında referanstan daha yararlı bilgi üretilebiliyor mu ve ölçüm sistemi güvenilir mi? Otomatik yatırım kararı verilmez. |
| Piyasa/evren | SPY, IWM, EFA, EEM, IEF, TLT, GLD, VNQ: kaldıraçsız 8 ETF başlangıç önerisi. Kimlik ve uygunluk veri aşamasında doğrulanır. |
| Evren sınırı | Liste performans taramasıyla seçilmedi; optimum veya bağımsız evren değildir. Bugün seçilmiş kohorttan tüm geçmiş ETF evrenine/kapanmış fonlara genellenmez. Sekiz ETF sekiz bağımsız piyasa deneyi değildir. |
| Birincil araştırma çıktısı | Nominal %80 getiri aralığı: .10 ve .90 quantile tahminleri. |
| İkincil çıktılar | Medyan; ayrı modelden basit getiri ortalaması; kalibre yükseliş olasılığı. Her biri kendi referansı ve ölçütüyle değerlendirilir. |
| Kullanım değeri | Kullanıcının değişen kararı, yanlış tahminin maliyeti ve ödemeye değer fayda bilinmiyor. Ürün aşamasından önce kullanıcı görüşmesi ve fayda testi gerekir. |
| İlk teknik etki eşiği | Birincil referansa göre en az %2 göreli pinball kaybı azalması önerisi; kullanıcı değeri veya kanıtlanmış fayda değildir. İstatistiksel yorum deney protokolünde. |
| İşlem kapsamı | Emir, broker hesabı, strateji/PnL optimizasyonu ve portföy yönetimi ilk fazda yok. |

Yön tahmini korunur, ancak düşük güçlü araştırma başlık başarı iddiası yapılmaz. Göreli getiri, geniş hisse evreni ve spread ayrı karar problemleridir; mevcut hedef başarısız olunca sonuçlara bakarak hedef değiştirme yapılmaz. Bunlar yeni protokol gerektiren sonraki seçeneklerdir. Sekiz ETF, %2 eşik ve aşağıdaki saatler uzmanlar veya kullanıcı tarafından verilmiş kesin gereksinimler değil, bu revizyonun açık plan varsayımlarıdır.

## 2. Zaman, hedef ve birim sözleşmesi

Takvim America/New_York ve sürümü sabitlenmiş ABD seans takvimidir. Kayıtlar UTC zaman damgasıyla saklanır; sunumda saat dilimi gösterilir. İstanbul saati sabit kaydırmayla türetilmez; ABD yaz saati değişir.

- İşlem günü **08:30 ET bilgi kesimi**, **08:45 ET tahmin kaydı**. 09:00 ET'ye kadar kaydedilmeyen iş late/failed olur; açılıştan sonra aynı tahmin geçmişe yazılmaz.
- Fiyat özellikleri en son tamamlanmış önceki New York gününden gelir; tahmin gününün tamamlanmamış günlük barı alınmaz. Önceki günün sağlayıcı günlük hacmi uzatılmış seansı içerebilir; açıkça provider_daily_volume adını taşır.
- d1, kayıt zamanından sonraki ilk düzenli seans; d5, d1 dahil beşinci seans. Hedef **OPEN(d1) → CLOSE(d5)**. Tatil/erken kapanışlar takvimden gelir; takvim günü eklenmez.
- OPEN/CLOSE doğrulanmış düzenli seansın ilk/son uygun bar fiyatıdır. Alpaca bar O/C'si resmî açılış/kapanış ihalesi veya gerçekleşebilir emir fiyatı garantisi değildir. Günlük barla seans eşleşmesi örnek dakika verisiyle sınanır; gerekirse seans adaptörü kullanılır. Eşleşme doğrulanamazsa etiket kabul edilmez. [Bar toplama kuralları](https://docs.alpaca.markets/us/docs/market-data-faq).
- 08:30 sonrası haber/fiyat bilgisi girdi değildir. Tahmin yayımlanırken başlangıç fiyatı henüz bilinmez; yalnız sonuç etiketi oluştuğunda kullanılır.

Başlangıçta bir birim payın alındığı varsayımsal servet hesabı:

$$
R_{i,o}=\frac{q_{i,d5}C_{i,d5}+D_{i,(d1,d5]}}{O_{i,d1}}-1,\qquad b_{i,o}=1[R_{i,o}>0].
$$

q, başlangıçtaki bir birimin bölünmeler sonrası pay adedidir. D, elde tutulan payların hak ettiği brüt nakit dağıtımı alacaklarıdır: temettü yeniden yatırılmaz; vergi, komisyon ve kayma hariçtir. d1 açılışında ex-dividend olmuş payı alan kişi o dağıtıma hak kazanmaz; ex-date'i d1'den sonra ve d5'e kadar olan hak edişler dahil edilir. Pay başına dağıtım o tarihteki pay adediyle eşlenir. Ödeme tarihi daha sonra olabilir; hedef nakit bakiye değil servet/hak edişidir.

Ham O/C, olay kaydı ve hesap sürümü birlikte saklanır. Spin-off, birleşme, kapanan fon ve olağan dışı dağıtımda değerleme belirsizse etiket pending_action_review olur; sıfır sonuç verilmez. Sonucu hesaplamak için sonradan öğrenilen olay ile tahmin özelliği olarak önceden bilinen olay ayrıdır.

Getiriler ondalıktır: 0.01 = %1. mean_simple_return, median_simple_return, p_up, q10, q90 ayrı alanlardır. Y=log(1+R) yardımcı modelde kullanılabilir; quantile dönüşümü tamdır, ortalama dönüşümü dağılıma bağlıdır. Student-t log getirinin üstel momenti sonsuz olduğundan bu modelden beklenen basit getiri üretilemez; Ridge ortalamayı doğrudan R üzerinde tahmin eder. Bu nakit yeniden yatırımsız hedef, beş close-to-close log getirinin toplamı değildir. [Matematik sözleşmesi](research/matematik-algoritmalari.md).

## 3. Veri kaynağı ve kullanım hakları

Ücretsiz erişim ML/non-display işleme, arşivleme, türetilmiş tahmin paylaşımı veya ticari kullanım hakkının kanıtı değildir. **İlk veri ediniminden önce** kaynak başına amaç, kişisel/şirket kullanımı, ML işleme, saklama süresi, ham/türetilmiş çıktı paylaşımı, haber görüntüleme ve dış Jev/LLM işleme hakları kayıt altına alınır. Hesaba özel anlaşma ve koşul sürümü manifestte tutulur. Belirsiz kaynak kolu kullanılmaz; izinler dağıtım aşamasına ertelenmez.

| Kaynak | Teknik olanak | Plan kararı |
|---|---|---|
| Alpaca | Belgelerde 2016'dan beri geçmiş, Basic'te 200 tarihsel istek/dakika; en az 15 dakika gecikmiş SIP sorguları. | İlk sayısal veri **adayı**. Gerçek hesap erişimi, olay kapsamı ve amaç için haklar doğrulanmadı. |
| Massive | Bireysel 2/5/10 yıl geçmiş için $0/$29/$79 aylık seviyeler; standard adjusted bars temettü yerine split düzeltir. | İkinci kaynak adayı; uygun hak ve aynı hedef tanımı gerekir. |
| Alpha Vantage | Daily Adjusted premium; ham OHLCV, adjusted close, split/dividend olayları; varlığa göre 25+ yıl. | Uzun geçmiş ve olay karşılaştırma adayı; kapsam ve haklar kontrol edilir. |
| Benzinga/Alpaca haber | 2015'e uzanan arşiv ilanı; seçilen varlık ve eski metin sürümleri belirsiz. | İlk fazda edinilmez; ayrı metin protokolü ve haklar gerekir. |
| SEC EDGAR | Anahtarsız submissions/XBRL. | Sonraki şirket araştırması; ETF pilotuna temel özellik değildir. |
| FRED/ALFRED | Vintage tekniği var; güncel koşullar ALFRED'i kapsıyor ve ML geliştirme/eğitim ile API arşivlemeyi kısıtlıyor. | Başlangıçtan çıkarıldı. Uygun yazılı izin veya doğrudan kaynağından bağımsız lisanslı vintage arşiv olursa yeni protokol. |

Kaynaklar: [Alpaca planları](https://docs.alpaca.markets/us/docs/about-market-data-api), [Massive fiyatları](https://massive.com/pricing), [Massive düzeltmesi](https://massive.com/knowledge-base/article/is-massives-stock-data-adjusted-for-splits-or-dividends), [Alpha Vantage API](https://www.alphavantage.co/documentation/), [haber geçmişi](https://docs.alpaca.markets/us/docs/historical-news-data), [SEC API](https://www.sec.gov/search-filings/edgar-application-programming-interfaces), [FRED koşulları — API (k)/(l)](https://fred.stlouisfed.org/legal/). Kontrol: 2 Ekim 2026; gerçek hesap testi yok.

Alpaca'nın standart destek açıklaması API verisinin yeniden dağıtımına izin vermiyor; türetilmiş tahmin hakları ayrıca incelenir. Massive non-display ve Derived Works için de lisans koşulları getiriyor; bireysel fiyat ürün bütçesi değildir. Alpha Vantage varsayılan lisansı kişisel/ticari olmayan kullanımla sınırlıyor. [Alpaca dağıtım](https://alpaca.markets/support/redistribute-alpaca-api), [Massive koşulları](https://massive.com/legal/market-data-terms-of-service), [Alpha Vantage koşulları](https://www.alphavantage.co/terms_of_service/).

Başkalarına menkul kıymet bazlı tahmin sunulacaksa ülke, kişiselleştirme, ücret ve kullanım şekline göre düzenleyici inceleme ürün aşamasının koşuludur. Her tahmin yayınının kesinlikle yatırım danışmanlığı olduğu genellemesi kabul edilmez; bu plan hukuki statüyü tespit etmiş değildir.

## 4. Point-in-time veri modeli ve kalite

Parquet değişmez ham snapshot, DuckDB sorgu/manifest katmanıdır. Kaynak revizyonu eski dosyanın üzerine yazılmaz; aynı manifestle deney tekrarlanır.

| Kayıt | Zorunlu alanlar / kural |
|---|---|
| Tümü | source, hak/koşul sürümü, stable_asset_id, geçerli sembol aralığı, event_time, source_published_at, vendor_available_at, first_observed_at, retrieved_at, version_id, raw hash, availability_evidence ve revision ilişkisi. Bilinmeyen zaman null + neden; uydurulmaz. |
| Fiyat | Feed, bar/seans aralığı, para birimi, ham OHLCV, adjustment=raw, asof tarihi, tam request parametreleri ve sayfalama. |
| Kurumsal olay | Announcement, ex/effective/record/pay zamanları, oran/tutar, para birimi; bilgi zamanı ile ekonomik etki ayrı. |
| Etiket | Forecast ID, d1/d5 ve gerçek fiyat zamanları, hesap/olay sürümü, actual matured_at, yeterlilik durumu, outcome sürümü. Historical exploratory replay için ayrı replay_eligible_at_assumed ve varsayım kaydı. |
| Haber — sonraki faz | Article ID, created_at, updated_at, tam gövde/hash ve ilk gözlem; endpoint filtre semantiği örnekle test edilir. |
| SEC — sonraki faz | CIK, accession, form, acceptance ve gerçek dissemination, amendment, original fact context/unit/period; hesap dönemi bilgi zamanı değildir. |
| Makro — izinli sonraki faz | Observation period, release zamanı, vintage, vendor availability/capture; vintage tarihi tek başına intraday kanıt değildir. |

Canlı özellikte olay tamamlanmış, sürüm kesime kadar gözlenmiş ve varsa yayın/erişim zamanları kesimden önce olmalıdır. Bugün indirilen tarihsel backfill'in first_observed_at/matured_at değeri geçmişe çekilmez; eski availability bağımsız arşiv kanıtıyla kurulur. Eski sürümü kanıtlanamayan fiyat revizyonları exploratory historical sınıfıdır. Böyle bir replay'de varsayılan bilgi/etiket uygunluk zamanı ayrı assumed alanında tutulur; gerçek capture zamanı korunur. Kanıtlı vintage veya yeterli forward değerlendirme olmadan historical skor confirmatory geçiş sağlamaz. Canlı-cutoff snapshotla backfill farkı ayrıca ölçülür. Split faktörünün oranlarda iptal olması bütün veriyi otomatik güvenli veya sızıntılı yapmaz; seviye/hacim/revizyon ayrı incelenir.

Alpaca sorgusunda feed=sip, timeframe=1Day, adjustment=raw ve açık start/end/asof zorunludur; bütün next_page_token sayfaları alınır. Varsayılan sembol eşlemesine güvenilmez, stabil kimlik ve yeniden adlandırma kaydı kullanılır. [Bars API](https://docs.alpaca.markets/us/reference/stockbars).

Haberin created_at filtresi bugünkü gövdenin eskiden bilindiğini kanıtlamaz. Endpoint eski body-version selector belgelemiyor; availability kanıtı yoksa tarihsel metin doğrulayıcı değildir, ileri capture gerekir. [Haber API](https://docs.alpaca.markets/us/reference/news-3). SEC'de bazı 17:30 ET sonrası başvurular ertesi iş günü yayılır; form istisnaları vardır. frames as-of değildir; companyfacts ve orijinal filing bağlamı actual dissemination/sürüm filtresiyle kullanılır, filed tarihi tek başına yetmez. SEC User-Agent ve erişim sınırları adaptörde uygulanır. [EDGAR dissemination](https://www.sec.gov/search-filings/edgar-search-assistance/accessing-edgar-data).

Sayısal kabul toleransları deney protokolündeki G1'dir. Uygulama kontrolleri:

- DST, tatil, erken kapanış, başlangıç tarihi, sembol değişimi, halt ve eksik gün.
- Eksik hacim ile sıfır hacim ayrımı; ileri doldurarak sahte seans/etiket oluşturma yasağı.
- Split, dividend, özel dağıtım, spin-off, birleşme/delisting ve yeni/eski sürüm uzlaştırması.
- O/C seans eşleşmesi, extended-hours hacim ve corrected-close revizyonu.
- İki kaynağı aynı hedef/olay/yeniden yatırım tanımında karşılaştırma; farklı adjusted close serilerini eşit varsaymama.

## 5. Sayısal mimari ve araştırma sırası

~~~mermaid
flowchart LR
    A[İzinli fiyat ve olay kaynakları] --> B[Sürümlü ham snapshotlar]
    B --> C[As-of özellik ve olgunlaşmış etiket]
    C --> D[Referans ve küçük kayıtlı model seti]
    D --> E[Önce değişmez sayısal forecast]
    E --> F[Ayrı sürümlü outcome]
    E --> G[Statik araştırma raporu]
    F --> G
~~~

Tek Python paketi ve günlük batch yeterlidir. İlk bağımlılıklar pandas **veya** polars, pyarrow, DuckDB, exchange_calendars, scikit-learn; kayıtlı GJR için arch, kayıtlı sığ challenger için LightGBM. statsmodels yalnız sonraki hipotez gerektirirse eklenir. Sürümler/lockfile uygulama aşamasında sabitlenir; belge kurulum değildir.

Model, özellik, parametre ve pencere seçimleri yalnız [deney kayıt tablosundan](research/deney-plani.md) okunur. İlk sıra rolling/expanding referanslar ve EWMA → Ridge/lojistik → doğrudan hedef quantile/location-scale → GJR veya sığ boosting. GJR günlük close-close ölçeği hedef aralığına doğrudan çevrilmez; gerçek OPEN→CLOSE hedef kalibrasyonu zorunludur.

Kalman/HMM sonraki kayıtlı araştırma; ARMA/ARIMAX, VAR, GP, OU/cointegration, derin ağlar, foundation model taraması, analog ensemble ve RL ilk fazda yok. Qlib/Freqtrade referanstır. FastAPI, UI, trading engine, Qlib dependency ve tam TradingAgents mimarisi pilot gereksinimi değildir. [Kod ve lisans incelemesi](research/github-trading-sistemleri.md).

## 6. Tahmin günlüğü ve operasyon sözleşmesi

Her planlanan (seans, varlık, model) işi scheduled → issued_on_time / late / failed / invalid_input / baseline_fallback durumuyla kayda girer. Fallback'in model kimliği açıkça değişir; başarısız challenger yerine sessizce referans geçirilip challenger skoru verilmez.

Forecast en az stable ID, universe/protocol/config ID, cutoff, issued_at, d1/d5, session/feed/target/units, immutable input manifest, feature code hash, model artifact hash, parametre/eğitim/refit/kalibrasyon sürümü, çıktılar, durum ve hata kodunu içerir. Sayısal kayıt açıklamadan **önce** yapılır. İleri bilgi yalnız outcome kaydına girer.

- Append-only JSONL, kanonik serileştirme ve SHA256(previous_hash || payload) zinciri; ilk hash/manifest ve gün sonu kökü saklanır. Doğrulayıcı kayıt kaybı ve oynamayı raporlar.
- Yerel zincir geçmiş yeniden zincirlenebildiği için bağımsız değişmezlik kanıtı değildir. Günlük kökün bağımsız zaman damgalı sabitlemesi gerekir; yalnız uygun hash paylaşımı ve haklar değerlendirilir. Bu revizyonda remote repo yaratılmaz/veri gönderilmez. Sabitleme yoksa rapor “yerel değişiklik tespiti, bağımsız zaman kanıtı yok” der.
- Forecast satırı güncellenmez. Düzeltme yeni supersedes kaydıdır; ilk forecast ana skorda kalır. Outcome revizyonu eski/yeni etiketi ve skor sürümünü korur.
- Eksik/stale fiyat veya kaynak hatası nötr/0 fiyat/haber yok yapılmaz. Gerekli fiyat/olay verisi yoksa varlık forecast'i yayımlanmaz. İsteğe bağlı eksik girdi politikası eğitimde tanımlanır.
- Ağ/429 hatasında en fazla 3 deneme, jitter/backoff ve Retry-After; cutoff/son yayın sınırı değişmez. Yetki/schema/kalıcı 4xx hata retry ile gizlenmez. Cutoff'a yetişmeyen cevap o forecast'e eklenmez.
- Günlük iş/varlık sayısı, gecikme, staleness, hatalar, feature drift, olgunlaşmış skor ve coverage izlenir. Kayıtlı drift uyarısı inceleme tetikler; otomatik parametre avı başlatmaz.
- Schema/zaman sızıntısı/doğrulama hatası modeli askıya alır. Son doğrulanmış artifact veya etiketli referans fallback olur; onarım yeni sürümdür. İstatistiksel retirement protokole bağlıdır.

## 7. Haber, Jev, analog arama ve LLM — koşullu kapsam

Sayısal pilot geçiş koşulları sağlanmadan entegrasyon işi değildir. Sonraki tasarım:

1. Metadata/takvim → sözlük veya TF-IDF → sabit dar Jev soruları matched ablation. 5–10 hisse işleyiş gösterimi olabilir, küçük finansal katkı kanıtı sayılamaz. Geniş hisse event-study için ayrı evren, tarihsel uygunluk, güç ve bütçe protokolü gerekir.
2. Jev yalnız olay türü, guidance yönü, şirket ilgisi gibi metinsel görev yapar. Tarih, fiyat hesabı ve sayısal benzerlik kodun işidir. Confidence finansal doğruluk/piyasa olasılığı değildir; Noul için ayrı confidence alanı varsayılmaz.
3. Eğitim cutoff'u, sürüm tutma, finansal kullanım ve dış işleme hakları öğrenilir. Cutoff bilinmiyorsa eski Jev testi exploratory; temiz doğrulama frozen modelle ileri capture ister. Anonimleştirme duyarlılık testidir, kontaminasyonu çözmez.
4. Alias yerine sabit model; tam input/output, istenen/dönen ID, hash, soru/prompt/schema, request zamanı, maliyet/retry/hata saklanır. Model/prompt değişimi yeni deneydir. Tekrar aynı çıktı garanti etmez; cevap cache'i deney girdisidir. Hata numerik fallback olur, eksik metin nötr sayılmaz.
5. Analog arama yalnız numerical/BM25 retrieval kontrolünü geçerse açılır. Aynı haber kümesi/örtüşen olaylar deduplicate edilir, yalnız sonucu olgunlaşmış analoglar seçilir. Jev aynı shortlist'te semantik rerank yapar; sonuca göre seçim yok.
6. LLM sabit sayıları/kaynakları açıklar. Template karşısında sayısal sadakat, kaynak desteği, kullanıcı anlayışı ve maliyet sınanır; değişmeyen forecast'in Brier/MAE'sini iyileştirdiği iddia edilmez. Abstention/weight selection/tahmin değişimi ayrı coverage-matched deneydir.

[TypeSafe modeller](https://docs.typesafe.ai/models), [Jev sınırlamaları](https://docs.typesafe.ai/model-jaggedness/jev-1.13), [metin kontaminasyonu araştırması](https://arxiv.org/abs/2309.17322). Açılma koşulları deney protokolündeki G5/G6'dır.

## 8. Aşamalar, kaynak ve açık kararlar

Bir deneyimli tam zamanlı geliştirici için tahmin; süreler söz verilmiş teslim tarihleri değildir.

| Aşama | Süre | Çıktı/koşul |
|---|---:|---|
| 0 — sözleşme | 2–4 iş günü | Kullanıcı/amaç, haklar, hedef ve ön kayıt; G0. |
| 1 — veri | 1–2 hafta | Raw+olay+takvim, seans adaptörü, kalite/revizyon raporu; G1. |
| 2 — ölçüm | 1,5–3 hafta | Date purge, referanslar, güç/MDE, holdout kilidi; G2. İleri referans günlüğü burada başlar. |
| 3 — sayısal deney | 1–2 hafta | Bütçe içi aday, tek seçilmiş challenger, calibration, tek holdout açma; G3. |
| 4 — ileri operasyon | Yaklaşık 1 hafta kurulum, 3 ile paralel | Batch/hash/outcome; 60 seans operasyon ve en az 26 hafta izlem; G4. |
| 5 — haber/Jev | Koşul geçerse ek 3–6 hafta | Ayrı güçlü event-study/ablation; temiz kanıtın takvim süresi ayrıca. |
| 6 — açıklama/ürün | Koşul geçerse ek 2–4 hafta | Template/LLM kullanıcı testi, haklar/düzenleyici inceleme, gerekirse UI/API; G6. |

0–4 yaklaşık **5–9 mühendis-haftası**; ileri gözlem buna ek takvim süresidir. 6–12 ay küçük üstünlüğü otomatik kanıtlamaz; gereken süre geliştirme loss serisinden hesaplanır, yetmezse sonuç “sonuçsuz”dur.

Kişisel/izinli veri için **$0–160/ay** veri+küçük compute/storage planlama zarfıdır, satın alma onayı değildir. Müşteri ürünü, haber display/archive ve bağımsız timestamp maliyeti teklif bekler. LLM seçilmedi; iş yüküyle sonraki aşamada bütçe çıkarılır. Jev'in ilan edilen $0.042/milyon input token, ücretsiz output fiyatında 5–75 milyon token ≈ $0.21–$3.15, retry hariç. Düşük çağrı fiyatı hak, kontaminasyon veya gücü çözmez. [TypeSafe fiyatı](https://docs.typesafe.ai/models).

Açık kullanıcı kararları: gerçek hedef kullanıcı/karar, kişisel–şirket–müşteri kullanım, dağıtım ülkeleri, minimum kullanıcı değeri, satın alma bütçesi, vazgeçilmez varlık ve ileri bekleme süresi. Mevcut varsayımlar revizyonu tamamlamak içindir; tercihler belli olunca sözleşme/protokol **holdout görülmeden** güncellenir. İlk uygulama yalnız izinli veri, referanslar, güç raporu ve ileri günlük olacaktır.
