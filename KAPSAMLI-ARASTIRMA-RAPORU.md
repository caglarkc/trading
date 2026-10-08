# Trading Araştırması — Güncel Kapsam ve Uzman Görüşlerinin Karar Kaydı

**Güncel geliştirme ve çalışma ortamı: Hetzner (`167.235.158.118`).** Backend, model bağlantıları ve veri işleri bu uzaktaki sunucuda geliştirilip çalıştırılır; Mac yerel web gözlem uygulaması ve uzak erişim içindir. Sunucu bağlantısı henüz yapılmadı.

Güncelleme: 8 Ekim 2026. Güncel ürün, sekiz ayrı karakterin kendi nakdiyle ABD hisseleri/QQQ/TQQQ/EWC üzerinde çalıştığı canlı sanal pilot A'dır. Ürün sözleşmesi [deney planında](research/deney-plani.md), tamamlanmış teknik kararlar ve uygulama sırası [MVP planındadır](MVP-TEKNIK-PLAN.md). Kodlama, sunucu bağlantısı ve normal seans kaynak kabul testi yapılmadı. Aşağıdaki 2 Ekim uzman sentezi yöntem geçmişidir; ikinci bir aktif uygulama planı değildir. Revizyon öncesi belge [arşivdedir](research/archive-2026-10-08/KAPSAMLI-ARASTIRMA-RAPORU.md).

## Güncel kapsam revizyonu ve ayrılıklar — 8 Ekim 2026

Teknik uygulama tasarımı [MVP planının](MVP-TEKNIK-PLAN.md) 12. bölümünde ayrıntılandırıldı: tek Go kod tabanında API/worker/runner süreçleri, typed SQL/PG queue, atomik muhasebe, özel CLI session/auth sınırları, kaynak kapasite testi ve deploy/restore düzeni. Ürün tercihleri korundu. Abonelik türü ve sunucu/SSH bilgisi beklenir; oturum/auth ve ücretsiz canlı veri erişimi tasarım metniyle kanıtlanmış değildir.

Kullanıcının son yetkisiyle emir/simülatör, ücretsiz fiyat kaynağı havuzu, kontrol sıklıkları ve model talimatları somutlaştırıldı. Backend Go/chi, PostgreSQL, model bağlantıları ve işler Hetzner'de; gözlem için yerel web uygulaması Mac'tedir. Eski Python sayısal pilot ve anahtarlı piyasa API başlangıcı uygulanmaz. Model özel hafızaları ayrılır; iki paralel worker başlangıcı seçildi. Gemini, tarihsel B ve gerçek broker daha sonraya bırakıldı.

| Konu | Önceki görüş / varsayım | Güncel seçim ve gerekçe |
|---|---|---|
| Araştırma başlangıcı | Uzman sentezi sayısal tahmin ve baseline/evaluator'ı öne koymuştu | Kullanıcı canlı karakter/cüzdan deneyini seçti, değerlendirmeyi şimdilik erteledi. Bu ürün tercihi uygulanır; yöntem üstünlüğünün kanıtlandığı söylenmez. |
| Mimaride LLM/emir/UI | Önceki sentez ilk fazdan çıkarmıştı | Güncel görev LLM'nin kendi hesabında karar vermesini gözlemlemek; LLM, kendi emir motoru ve basit UI bunun gereğidir. Hazır trading backend kullanılmaz. |
| Evren/hedef | 8 ETF, beş seanslık quantile ve sabit cutoff asistan varsayımıydı | 50+20 ABD hissesi, QQQ/TQQQ/EWC, 8 cüzdan ve iki günlük oturum kullanıcı kararıdır. Eski sayısal hedefler aktarılmaz. |
| Başarı/kıyas | Önce baseline ve güç analizi önerilmişti | Kullanıcı 8 Ekim'de değerlendirmeyi erteledi. Kıyas botu, başarı paneli, %10 görevi veya model sıralaması MVP'de yok. Hesap/işlem gözlemi ve kayıtlar korunur. |
| Kaynak bağımsızlığı | İki site veya üç fiyat sitesini bağımsız feed sayma riski vardı | Menşe/feed bilgisi ayrılır. Üç site fiyat uyumudur; haber kopyaları iki bağımsız kanıt değildir. Bu uzman düzeltmesi korundu. |
| Fiyat kaynağı | Yahoo/Nasdaq/Stock Analysis zorunlu ilk üçlüydü | Stock Analysis/Yahoo/Webull havuzu seçildi; Nasdaq yedek. Üçü de %1 ve 60 saniye koşulunu geçmeli. Erişim araştırması başarılı canlı üçlü testi sayılmaz. |
| Gerçekçilik | Fiyatı görmekten gerçek fill sonucu çıkarma riski vardı | Güncel bid/ask, miktar ve zaman kanıtı, açık maliyet modeli ve partial/timeout şartı konuldu. Kanıt yoksa fill yok. |
| Bellek/sızıntı | Sistem promptunun modeli tarihsel sonucu unutmaya zorlayabileceği varsayımı | Böyle garanti yok. Ana sınav canlı ileri kayıt; B sonraki kapsam. Özel cüzdan belleği ve kaynak olguları ayrılır. |

Her iki uzmanın veri zamanı, kaynak kanıtı, maliyet ve geçmiş bilgi sızıntısı düzeltmeleri geçerlidir. Ancak eski ilk faz mimarisi güncel kullanıcı görevini yerine getirmediği için uygulama planı olarak seçilmedi. Kullanıcı değerlendirmeyi ertelediğinden bu MVP sonuçlarından istatistiksel üstünlük veya gelecekte kâr garantisi çıkarılamaz; şimdi bir değerlendirme modülü eklenmez.

Fiyat havuzu seçimi ile canlı erişim başarısı farklıdır: 8 Ekim sayfa araştırması Stock Analysis/Webull erişimini gösterdi; Yahoo 429, Nasdaq boş fiyat gösterdi. Normal seans kapalı olduğundan 60 saniyelik üçlü şartı kanıtlanmadı. Hetzner'de bu kabul kontrolü geçmeden sanal execution etkinleşmez. Sıradaki iş uygulama kabul planında belirtilen gerçek abonelik ve izolasyon prototipidir.

Son uygulama sırası yönlendirmesi: üç güncel fiyat ve likidite alanları veri çekme prototipi aşamasında sınanacak; bütün yapı bitince yapılacak ayrı bir kontrol olarak bekletilmeyecek. Hafıza ayrılığı kendi modülü içinde iki-cüzdan işlev/yetki kontrolüyle doğrulanacak; ayrı uzun deney fazı eklenmeyecek.

## 2 Ekim 2026 uzman sentezi — tarihsel yöntem kaydı

Aşağıdaki hedefler, konfigürasyon bütçeleri, modeller, veri/API ve eski ilk teslimat seçimleri **güncel MVP gereksinimi değildir**. Kaynak bağlantıları ve matematik düzeltmeleri geçmişte yapılan değerlendirmenin kanıt sınırını korumak için tutuldu.

Bu belge iki uzman değerlendirmesinin karar kaydıdır. Önceki sürümün diğer dosyaları kopyalayan bölümleri kaldırıldı; ayrıntılar tek kaynakta korunur. Kopyalar bağımsız doğrulama değildi ve ileride farklı planlar oluşmasına yol açabilirdi.

## 1. Değerlendirilen görüşler ve kanıt sınırı

- **A:** “Overall verdict: proceed with a smaller, explicitly falsifiable research pilot…” başlıklı ilk yapıştırılan değerlendirme. [Orijinal metin](</Users/caglarkc/.codex/attachments/b443c5f2-801d-4765-8ae9-d2b4eb2095bd/Yapıştırılan metin.txt>).
- **B:** “Verdict — I don't recommend building the architecture as written…” başlıklı ikinci değerlendirme. [Orijinal metin](</Users/caglarkc/.codex/attachments/26621f31-6f15-4d9e-ac19-2c32e6a6dcf6/Yapıştırılan metin.txt>).

A ve B aynı beş plan dosyasını ele alıyor. İki ayrı eleştiri değerli çapraz kontroldür; ortak kaynaklar ve önceki kopya rapor bağımsız performans kanıtı değildir. Bu revizyonda resmî veri koşulları, API belgeleri, matematik kaynakları ve kritik sabit repository kodları kontrol edildi. Hesap erişimi, gerçek dataset, çalışma zamanı, eğitilmiş model veya forward forecast yok.

Her iki görüşün ortak, kabul edilen sonucu: veri ve baseline/evaluation önce; fiyat, bilgi kesimi ve label uyumu zorunlu; gerçek güç analizi ve korunan test gerekli; Jev/LLM'nin katkısı bilinmiyor; ücretsiz veri dağıtım hakkı demek değil; emir motoru gerekmiyor.

## 2. Gerçek ayrılıklar ve hangi yaklaşım seçildi?

| Konu | A'nın yaklaşımı | B'nin yaklaşımı | Seçim ve gerekçe |
|---|---|---|---|
| Ana hedef | Küçük, ekonomik olarak farklı 6–10 ETF ile beş günlük yön olasılığı pilotunu koru. | Ana hedefi göreli getiri/oynaklığa yönlendir; az sayıda ilişkili ETF'de küçük yön etkisi ölçülemez. | **A'nın küçük pilot kapsamı + B'nin güç önceliği.** Aralık araştırması primary, mutlak yön/mean secondary ve ayrı güç koşullu. Göreli hedefe otomatik geçiş yok; o ayrı kullanıcı problemi ve evren ister. |
| Güç iddiası | Gerçek paired-loss varyansından güç ve gerekli süreyi hesapla. | Tek-seri korelasyon hesabıyla mevcut tasarımın çoğunlukla yetersiz olduğunu savunuyor. | **A'nın hedefe özgü hesabı seçildi; B'nin düşük güç uyarısı kabul edildi.** B'nin yaklaşık sayıları belirli varsayımlarda makul, fakat bütün interval/panel araştırmalarının imkânsızlığını veya yalnız bir hedefin mümkün olduğunu kanıtlamaz. |
| Deneme bütçesi | Yaklaşık 8–12 ön kayıtlı konfigürasyon. | En çok 15; daha geniş başlangıç listesini daralt. | **En çok 15 toplam kayıt, baselines dahil; bütün satırları koşmak zorunlu değil.** Her pencere/parametre/özellik/calibration seçimi sayılır. Bu sentez tercihi, daha çok modelin daha iyi olduğu iddiası değil. |
| Log→simple mean | Student-t üstel momenti sonsuz; simülasyon ortalaması geçerli mean değil. | Log iç model, quantile'ı tam, mean'i dağılımdan dönüştür. | **A seçildi.** B'nin önerisi gerekli moment koşulu olmadan eksik. Ridge doğrudan simple R'yi öğrenir; Student-t uncertainty için kullanılabilir. [Student-t yoğunluğu](https://docs.scipy.org/doc/scipy/reference/generated/scipy.stats.t.html). |
| GARCH beş günlük variance | Günlük horizon'ların toplamını ayrı ele al; h.5 cumulative değil. | Constant-mean martingale innovations altında variance analitik toplamdır; simülasyon variance için zorunlu değil. | **B'nin daha kesin matematik düzeltmesi seçildi.** Simülasyon distribution shape/quantile için gerekebilir. Ayrıca yeni OPEN→CLOSE label beş close-close toplamı değildir; direct-target adaptör şart. [arch forecasting](https://arch.readthedocs.io/en/latest/univariate/forecasting.html). |
| ALFRED | ML geliştirme/eğitim ve API arşivleme kısıtını tespit ediyor. | Vintage önerisini koruyor; bu koşulu doğrulayamamış. | **A seçildi; başlangıçtan çıkarıldı.** Güncel koşullar ALFRED'i kapsıyor, API (k)/(l) planlanan işlemlere kısıt getiriyor. İzinli alternatif veya yazılı izin gerekir. [FRED koşulları](https://fred.stlouisfed.org/legal/). |
| LLM/UI | Şablon veya optional açıklama daha sonra olabilir. | İlk mimariden LLM, Jev, UI ve API'yi çıkar. | **İlk fazda B; sonraki açıklama testinde A.** Batch+statik rapor; LLM template kontrolünü ve kullanıcı fayda testini geçerse eklenir. |
| Tarihsel haber sürümü | Görülen API eski body-version selector belgelemiyor; availability kanıtlanmazsa exploratory. | Orijinal gövde geri alınamaz diye daha kesin ifade ediyor. | **A'nın kanıt sınırı seçildi.** Endpointte özellik bulunmaması bütün sağlayıcılarda geri alımın imkânsızlığını kanıtlamaz; exact-version kanıtı olmadan eski test confirmatory değildir. [News API](https://docs.alpaca.markets/us/reference/news-3). |
| Ticari maliyet | Veri, display/archive ve derived rights fiyatı anlaşma bekliyor. | Kullanıcı ürünü için yüzler/binler USD tahmini. | **A seçildi.** Ticari veri bütçesi teklif olmadan kesinleştirilmez; kişisel plan fiyatı ticari maliyet sayılmaz. |
| GPL/AGPL/Commons Clause | Kullanım biçimi ve ayrı lisans koşullarını incele. | Hosted üründe bu lisanslardan genel olarak kaçın. | **A'nın ayrımı seçildi.** Genel yasak hükmü doğru değil; ilk çekirdeğin bu motorlara ihtiyacı yok. vectorbt ek Commons Clause nedeniyle unrestricted OSS diye sunulmaz. |
| SEC yayın zamanı | Acceptance/dissemination, amendment ve original context ayrılmalı. | 17:30 sonrası filings ertesi gün diye genelleştiriyor. | **A'nın ayrımı + resmî belgenin daha dar ifadesi.** Bazı submissions ve form istisnaları var; filed tarihi tek başına as-of değildir. [SEC erişim](https://www.sec.gov/search-filings/edgar-search-assistance/accessing-edgar-data). |

İki görüş bütünüyle birbirinin karşıtı değildir. A “küçük ve yanlışlanabilir pilot”, B “mimarinin mevcut hali yerine veri/ölçüm ve hedefi tekrar değerlendir” diyor. İkisini birden sınırsız özellik listesine dönüştürmek yerine ortak küçük çekirdek seçildi.

**Bu revizyonun yeni plan varsayımları:** sekiz isimli ETF; 08:30 ET cutoff/08:45 issue; OPEN(d1)→CLOSE(d5) gross cash-rights return; primary .10/.90 ortalama pinball ve %2 teknik planning threshold. Bunlar uzmanların üzerinde uzlaştığı optimumlar veya kullanıcının kesin tercihleri değildir. Kullanıcı değeri ve development MDE ile holdout açılmadan değiştirilebilir; sonuçtan sonra başarı eşiği indirilmez.

## 3. Ortak bulguların kapsama uygulanması

| Alan | Revize kapsam |
|---|---|
| İlk teslimat | İzinli fiyat/olay snapshotları, exact hedef, güçlü rolling/expanding/EWMA referans, chronological evaluator, güç/MDE ve forward ledger. |
| Aktif modeller | Registry içi Ridge/lojistik; doğrudan R quantile; gerekli koşullarda GJR scale veya sığ boosting. Mean, median, probability ve interval ayrı functional/score. |
| Ertelenen modeller | Kalman/HMM ikinci kayıtlı tur; ARMA/ARIMAX, VAR/GP, spread/OU, hiyerarşi, foundation/derin modeller, ensemble/conformal ve analog için bağımsız gerekçe/test gerekir. |
| Veri | Raw fiyat + kendi açık action hesabı; feed/adjustment/asof açık. Exact event/availability/capture/version alanları. Eski snapshot kanıtı bilinmiyorsa exploratory; ileri capture ayrı. |
| Hedef zamanı | Kapanıştan sonra öğrenilen bilgiyle önceki kapanıştan getiri iddiası kaldırıldı. Beş seans başlangıcı issue sonrasında; resmî auction fiyatı iddiası yok. |
| Değerlendirme | Date-group split, actual interval purge, yalnız fit zamanında olgunlaşmış labels, OOS calibration, tek holdout unseal, bütün trial sicili. |
| İstatistik | Primary paired loss date-block inference; MDE gerçek loss'tan; Holm kontrollü ayrı iddialar, coverage/width/interval score ve non-overlap diagnostics. Yetersiz güç “fayda yok” sayılmaz. |
| Tahmin kaydı | Her scheduled origin/asset; append-only hash chain, ayrı outcome/correction, independent hash timestamp kanıt sınırı, bounded retry/stale/fallback/rollback. |
| Haber/Jev | Ayrı event-study; metadata/calendar ve cheap text kontrolü, finance annotation seti, fixed version/questions ve tam request-response cache. Numeric similarity kodda. |
| LLM | Sabit forecast + template vs LLM explanation testi; source/number consistency ve user comprehension. Abstention/model seçimi ayrı coverage-matched policy. |
| Dağıtım | Şimdilik özel araştırma varsayımı. Amaç için veri hakları ilk edinimden önce; dış kullanıcı/ticari aşamada kullanım biçimine uygun hukuk/veri incelemesi. |

**Matematikte B'den kabul edilen ek düzeltmeler:** Joseph Kalman covariance update; signed CQR score ve finite-sample rank; mean→MSE/RMSE, median→MAE, quantile→pinball. A'nın eklediği mixture variance ile predictor-error covariance ayrımı, mixture CDF quantile ve probability/quantile coherence de eklendi. CQR'nin iid coverage garantisi finansal zaman serisine otomatik taşınmaz; signed correction sonlu örnekte negatif de olabilir.

**Repo incelemesinde iki somut düzeltme:** FinRL-X quote hatasında 100.0 fiyat ve beş sembollük mapping içeriyor; execution güvenilirliği kanıtı değil. TradingAgents jev-latest kullanıyor ve hata halinde unscreened ileti geçiriyor. B'nin “sessiz” ifadesi harfi harfine doğru değil: kod warning ve metinsel not üretir; yine de başarısız screening başarılı feature gibi kullanılamaz. Qlib label ve cross-sectional normalization birlikte uyarlanmalıdır. Kanıtlar [repository raporunda](research/github-trading-sistemleri.md).

## 4. İlerleme, başarısızlık ve sonuçsuzluk

Bütün eşikler ve test ayrıntıları [deney protokolündeki G0–G6](research/deney-plani.md) kaydında tutulur; burada ikinci bir eşik seti oluşturulmaz.

- Veri/izin/zaman sözleşmesi yoksa ilgili kol çalışmaz.
- Güç yetersizse operasyon veya keşif sürebilir; küçük üstünlük/product claim açılamaz.
- Tek seçilen interval challenger, güçlü referanslar ve kayıtlı testte değerlendirilecek. Mean/direction aynı başarıyı miras almaz.
- Nokta faydası ≥%2 ve CI altı >0, “en az %2 fayda kanıtlandı” demek değildir. İkinci iddia CI altının %2'yi geçmesini ve daha güçlü test tasarımını gerektirir.
- Sonuç hem faydayla hem faydasızlıkla uyumluysa **sonuçsuz**; yeterince dar CI anlamlı faydayı dışlıyorsa planning hedefi desteklenmiyor. Bütçe tükenince yeni modeller ekleyerek sonuç avı yapılmaz.
- 60 seans veya 26 hafta forward kayıt operasyonu ve drift'i sınar. Küçük edge'i otomatik kanıtlamaz; gerçek güçle gerekenden kısa süre başarı kanıtı sayılamaz.
- Jev/LLM/UI sadece kendi veri, incremental fayda ve kullanıcı değeri koşullarını geçince açılır. Relative-return veya volatility ürünü başarısız mutlak yönün aynı testte yeniden adlandırılması değildir.

## 5. Süre, bütçe ve henüz belirlenmemiş tercihler

İzinli küçük pilot için 5–9 mühendis-haftası; ayrı haber/Jev için koşul geçerse ek 3–6, açıklama/ürün için ek 2–4 hafta. İleri doğrulama ek takvim süresidir ve MDE'ye bağlıdır. Kişisel veri+compute/storage zarfı $0–160/ay mühendislik tahminidir; ticari display, derived, archive/news hakları ve LLM ürünü fiyatlandırılmadı. B'nin doğrulanmamış ticari/LLM fiyatları taahhüde dönüştürülmedi. Ayrıntılı aşamalar [MVP planında](MVP-TEKNIK-PLAN.md).

Kullanıcı/karar, kişisel–şirket–müşteri kullanım, ülke, minimum fayda, bütçe, vazgeçilmez varlık ve bekleme süresi açık kalıyor. Belge güncellemesi için özel araştırma varsayımı kullanıldı; bu alanlarda bir tercihin verildiği iddia edilmedi. Bu revizyonda ücretli veya authenticated çağrı, kurulum, remote paylaşım, model eğitimi veya uygulama yapılmadı.

## 6. Belgeler ve bulgu izlenebilirliği

| Tek kaynak | Sorumluluk |
|---|---|
| [MVP-TEKNIK-PLAN.md](MVP-TEKNIK-PLAN.md) | Kullanıcı/kapsam, hedef/zaman/action, veri hakları ve PIT şeması, batch/ledger/operasyon, koşullu Jev/LLM, aşamalar/bütçe. |
| [research/deney-plani.md](research/deney-plani.md) | Tek registry, trial/feature bütçesi, tarihler/purge/refit/calibration, primary/secondary skor, MDE/test/gate/stop. |
| [research/matematik-algoritmalari.md](research/matematik-algoritmalari.md) | Düzeltilmiş denklemler, distribution/functional uyumu ve aktif/ertelenmiş algoritma kataloğu. |
| [research/github-trading-sistemleri.md](research/github-trading-sistemleri.md) | 13 repo kataloğu, sabit code/license kanıtları, somut kusurlar ve reuse sınırları. |
| Bu karar raporu | A/B seçim gerekçesi, kapsam sentezi ve doğrulama sınırı; diğer dosyaların kopyası değil. |

| Uzman bulguları | Karşılık |
|---|---|
| A C1; B D1/D5/D6 | MVP amaç/haklar ve G0/G6; ALFRED özel engeli ve hukuk statüsünün bilinmezliği. |
| A C2/C3; B C6/C7/C8/C9/C10 | Math moment/functional/GARCH/Kalman/CQR; experiment scores ve direct target adapter. |
| A C4; B C4/C11 | Repo FinRL-X/TradingAgents/Qlib/release düzeltmeleri; tek registry. |
| A C5; B D11 | Sabit numerical output üzerinde explanation testi; ayrı abstention/policy. |
| A D zaman/availability/operations; B C2/D7/D12/D14 | MVP zaman/PIT/ledger; experiment matured-label purge, failures ve versioned outcome. |
| A D güç/endpoint/acceptance; B C1/D2/D3/D8 | Experiment primary/date split/MDE/correction/gates; küçük ETF kohort sınırı ve inconclusive. |
| A D text validation; B C3/D4/D9/D10/D13 | Finance annotation, cheap text/event controls, geniş ayrı event-study, fixed requests/cutoff/anonymization, semantic-only Jev. |
| B C5 ve A consolidation uyarısı | Bu raporun kopya derleme yerine karar kaydı yapılması. |

Doğrulama sınırları: tarihsel fiyat/metin sürümleri ve hesap hakları henüz gerçek kaynak hesabında test edilmedi. Repo kodlarının çalışır olduğu sertifikalanmadı; kesin snapshot erişimi tamamlanmayan lisans/release ayrıntıları repository raporunda açıkça işaretli. Plan varsayımları, kaynakta doğrulanan olgular ve henüz ölçülmemiş hipotezler ayrı tutuldu.
