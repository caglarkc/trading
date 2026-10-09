# GitHub trading sistemleri araştırması

**Güncel MVP geliştirme ve çalışma ortamı: Hetzner (`167.235.158.118`).** Kendi backend'imiz, model bağlantıları ve veri işleri uzaktaki sunucuda geliştirilip çalıştırılır; Mac yerel web gözlem uygulaması ve uzak erişim içindir. 9 Ekim'de root/22 SSH erişimi salt okunur doğrulandı; sunucu başka projelerle paylaşılır ve trading ayrı kurulacaktır. Mevcut ChatGPT Pro kullanılır; model seçimi sonraya bırakıldı.

Güncel kapsam notu: 8 Ekim 2026. Bu katalog **2 Ekim repo incelemesinin yöntem geçmişidir**, güncel uygulama bağımlılık listesi değildir. Kullanıcı kendi Go backend'ini seçti; hazır trading API/backend, Python batch, Qlib/LEAN/NautilusTrader/FinRL motoru güncel MVP'ye eklenmez. HTTP routing için chi ve depolama için PostgreSQL kullanılması kendi iş kurallarımızı başka bir trading ürünüyle değiştirmez.

Güncel [MVP teknik planı](../MVP-TEKNIK-PLAN.md) ve [ürün/model sözleşmesi](deney-plani.md) önceliklidir. Backend/model işlerimiz Hetzner'de, gözlem uygulaması Mac'tedir; benchmark/değerlendirme ve tarihsel B sonraya bırakıldı. Aşağıdaki sürüm/lisans/commit gözlemleri bugünkü en güncel sürüm iddiası değildir. Önceki belge [arşivde](archive-2026-10-08/github-trading-sistemleri.md) korunur.

## 2 Ekim 2026 tarihsel repo incelemesi

Bu katalog depo araştırmasını ve iki uzman değerlendirmesinden sonra yapılan düzeltmeleri kaydeder. Resmî GitHub kaynakları, lisans metinleri, sürüm ve commit geçmişi ile seçili uygulama dosyaları incelendi. Hiçbir repo klonlanmadı, kurulmadı veya çalıştırılmadı. Kodun statik olarak incelenmesi gerçek hesapta çalışabilirlik veya tahmin başarısı kanıtı değildir. Buradaki katalog, matematik ve deney belgelerinden bağımsız bir performans kanıtı oluşturmaz.

Ana karar: İlk pilot tek bir Python batch iş akışı, sürümlü veri/tahmin kayıtları ve yerel statik araştırma raporuyla kurulur. Qlib, trading motoru, broker, çoklu agent orkestrasyonu, Jev veya LLM ilk pilotun bağımlılığı değildir. Qlib özellik ve deney tasarımı için kaynak olarak tutulur; TradingAgents gerçek Jev çağrısının yanında kaçınılması gereken tekrar üretilebilirlik ve hata davranışlarını da gösterir. LEAN/NautilusTrader ancak ayrıca yetkilendirilmiş bir strateji veya yürütme kapsamında değerlendirilir. FinRL ailesinin portföy ödülü hedefi, kalibre edilmiş getiri tahminiyle aynı görev değildir.

## İlk pilotta kullanım sınırı

| Alan | Şimdiki karar | Yeniden değerlendirme koşulu |
|---|---|---|
| Veri ve değerlendirme | Kendi point-in-time arşivi, deterministik özellikler, basit referanslar, tek deney kaydı ve statik rapor | Kullanım hakları, hedef/zaman sözleşmesi ve güç analizi tamamlanmalı |
| Qlib | Kurulum/fork yok; küçük özellik seti ve rolling deney fikirleri incelenebilir | Kendi beş seanslık hedefini ve etiket dönüşümlerini koruyan somut ihtiyaç |
| TradingAgents/Jev | Tahmin yolunda yok; metin katmanı ayrı koşullu araştırma | Finansal metin doğrulaması, lisanslı ve zamanında yakalanmış metin, önceden kayıtlı ek katkı deneyi |
| LLM açıklaması | İlk rapor şablonla üretilir | Ayrı doğruluk/anlama/maliyet deneyi; değişmeyen sayısal tahminin skorunu iyileştirdiği iddia edilmez |
| Backtest, RL ve yürütme | Kapsam dışında | Yeni strateji kararı, işlem maliyeti sözleşmesi ve bağımsız operasyon doğrulaması |

Uzmanlar arasında başlangıç mimarisini daraltma konusunda esaslı bir çelişki yoktur; ikinci görüşün daha açık veri–referans–değerlendirme başlangıcı seçildi. İlk görüşün FinRL-X yürütme kusuru korunmuştur. TradingAgents hakkında ikinci görüşteki “sessiz hata” ifadesi ise sabit kodla tam örtüşmez: hata logu ve metinsel uyarı vardır. Asıl sorun filtrelenmemiş metnin başarılı Jev özelliği gibi kullanılması ve hareketli model alias'ıdır; aşağıdaki düzeltmeler bunu esas alır.

## Depoların karşılaştırılması

Aşağıdaki tarihler kalite veya kârlılık kanıtı değildir; ilk araştırmada gözlemlenen bakım göstergeleridir. Bunlar yeni bir kurulum için “en güncel sürüm” seçimi değildir. GitHub latest release ile varsayılan dalın commit geçmişi farklı şeylerdir. Özellikle LEAN’in eski release sayfası aktif bakımını doğru yansıtmaz. Lisans adları çekirdek depo içindir; veri, model ağırlığı, plugin, bağımlılık ve bulut hizmeti haklarını kapsamaz. Sabit lisans dosyaları ve bu revizyondaki doğrulama sınırı aşağıda belirtilmiştir.

Sürüm tarihleri sorgu anındaki GitHub `published_at`, commit tarihleri `commit.committer.date` alanlarına dayanır; `pushed_at` kullanılmadı. Kaynak kayıtları aşağıda ayrıca bağlantılanmıştır. Commit bağlantıları içerik snapshot’ını sabitler; release etiketleri taşınabilir. GitHub latest release sorgusunun 404 dönmesi paket sürümü bulunmadığını kanıtlamaz.

| Depo | Kategori ve lisans | Gözlemlenen bakım göstergesi | Proje açısından değerlendirme |
|---|---|---|---|
| [Microsoft Qlib](https://github.com/microsoft/qlib) | Tahmin ve ML araştırma platformu; MIT | v0.9.7: 15 Ağustos 2025; main commit: 16 Eylül 2026 | Özellik ve deney referansı; ilk pilotta bağımlılık yok |
| [TradingAgents](https://github.com/TauricResearch/TradingAgents) | LLM analiz ve karar araştırması; Apache-2.0 | v0.5.2: 29 Eylül 2026 | Gerçek Jev çağrısı ve hata/sürüm sorunları için referans; koşullu sonraki araştırma |
| [QuantConnect LEAN](https://github.com/QuantConnect/Lean) | Olay tabanlı backtest ve canlı işlem motoru; Apache-2.0 | master commit: 1 Ekim 2026; GitHub latest release 2017 | İleride gerçekçi strateji simülasyonu ve ABD piyasası işleyişi |
| [NautilusTrader](https://github.com/nautechsystems/nautilus_trader) | Rust çekirdekli simülasyon ve yürütme; bildirilen LGPL-3.0, sabit dosya yeniden doğrulaması açık | develop commit: 2 Ekim 2026; stable v1.231.0: 2 Ağustos 2026; v2 RC hattı mevcut | Daha yüksek frekanslı veri veya işlem düşünülürse alternatif |
| [FinRL](https://github.com/AI4Finance-Foundation/FinRL) | Finansal reinforcement learning; MIT | v0.3.8: 20 Mart 2026; commit: 28 Eylül 2026 | Araştırma kaynağı; ilk tahmin MVP’sine doğrudan uygun değil |
| [FinRL-X / FinRL-Trading](https://github.com/AI4Finance-Foundation/FinRL-Trading) | Modüler portföy sistemi; Apache-2.0 | v1.0.0: 25 Mart 2026; master commit: 18 Eylül 2026 | Kavramsal mimari referans; sabit yürütme kodunda güvensiz placeholder'lar var |
| [Zipline Reloaded](https://github.com/stefan-jansen/zipline-reloaded) | Python olay tabanlı backtest; Apache-2.0 | 3.1.1: 23 Temmuz 2025; default branch commit: 13 Kasım 2025 | Günlük çoklu varlık araştırması için alternatif |
| [Backtrader](https://github.com/mementum/backtrader) | Python backtest ve canlı işlem; GPL-3.0 | Default branch son commit: 19 Nisan 2023 | Öğretici ve olgun; yeni ürünün çekirdeği için bakım ve dependency riski |
| [Backtesting.py](https://github.com/kernc/backtesting.py) | Hafif strateji backtest; AGPL-3.0 | Default branch commit: 5 Ağustos 2026 | Tek varlıkta hızlı referans strateji deneyi |
| [vectorbt](https://github.com/polakowo/vectorbt) | Vektörleştirilmiş araştırma; Apache-2.0 + Commons Clause, OSI anlamında açık kaynak değil | v1.1.1: 26 Eylül 2026 | Kısıtlı kaynak lisansı; ilk pilotta kullanılmaz |
| [Freqtrade](https://github.com/freqtrade/freqtrade) | Kripto botu; bildirilen GPL-3.0, sabit dosya yeniden doğrulaması açık | 2026.9: 29 Eylül 2026; commit: 2 Ekim 2026 | ABD hisselerine doğrudan uymaz; test araçları ilham kaynağı |
| [Jesse](https://github.com/jesse-ai/jesse) | Kripto araştırma sistemi; çekirdek MIT, live plugin ayrı | Default branch commit: 1 Ekim 2026 | Araştırma araçları yararlı; hazır hisse desteği varsayılmamalı |

Ek araştırma kaynağı: [FinMem](https://github.com/pipiku915/FinMem-LLM-StockTrading), MIT lisanslı akademik LLM hafıza uygulaması. [Son gözlenen commit](https://github.com/pipiku915/FinMem-LLM-StockTrading/commit/be814aa47970de9bf2fdd6a1d5a60ae5cf361b46) 18 Ağustos 2024; çalışan ürün altyapısından çok makale kodu olarak ele alınmalı.

### Bakım göstergelerinin kaynak kayıtları

| Depo | Sürüm kaydı | Commit kaydı |
|---|---|---|
| Qlib | [v0.9.7](https://github.com/microsoft/qlib/releases/tag/v0.9.7) | [16 Eylül 2026](https://github.com/microsoft/qlib/commit/be725493eb1a6bbb42bf11b37aa7669f59610ff1) |
| LEAN | [v2.4.0.1](https://github.com/QuantConnect/Lean/releases/tag/v2.4.0.1) | [1 Ekim 2026](https://github.com/QuantConnect/Lean/commit/33e3945f2faa95308d972dfd3d7b762f7743d1d5) |
| NautilusTrader | [v1.231.0](https://github.com/nautechsystems/nautilus_trader/releases/tag/v1.231.0), [v2.0.0rc5](https://github.com/nautechsystems/nautilus_trader/releases/tag/v2.0.0rc5) | [2 Ekim 2026](https://github.com/nautechsystems/nautilus_trader/commit/2863913442f6b9f847fc0d1448ee94f8e6579886) |
| FinRL | [v0.3.8](https://github.com/AI4Finance-Foundation/FinRL/releases/tag/v0.3.8) | [28 Eylül 2026](https://github.com/AI4Finance-Foundation/FinRL/commit/e60e26e4870f00fbc704c6297edfd9d81816b3d3) |
| FinRL-X | [v1.0.0 — 25 Mart 2026](https://github.com/AI4Finance-Foundation/FinRL-Trading/releases/tag/v1.0.0) | [18 Eylül 2026](https://github.com/AI4Finance-Foundation/FinRL-Trading/commit/4409abe925c904e570be78ebfb5e77ac3491dff8) |
| TradingAgents | [v0.5.2](https://github.com/TauricResearch/TradingAgents/releases/tag/v0.5.2) | [29 Eylül 2026](https://github.com/TauricResearch/TradingAgents/commit/8b22d43d01d9ddda5d686d093d5385884622f3de) |
| vectorbt | [v1.1.1](https://github.com/polakowo/vectorbt/releases/tag/v1.1.1) | [26 Eylül 2026](https://github.com/polakowo/vectorbt/commit/ceffc501f2d37033a79dd86a9f883e69ec6977bd) |
| Backtesting.py | Latest release sorgusu 404 | [5 Ağustos 2026](https://github.com/kernc/backtesting.py/commit/ca2e2611621e472542ba90f7243a1fa06a7d7108) |
| Backtrader | Latest release sorgusu 404 | [19 Nisan 2023](https://github.com/mementum/backtrader/commit/b853d7c90b6721476eb5a5ea3135224e33db1f14) |
| Zipline Reloaded | [3.1.1](https://github.com/stefan-jansen/zipline-reloaded/releases/tag/3.1.1) | [13 Kasım 2025](https://github.com/stefan-jansen/zipline-reloaded/commit/943010b9da848e317fc520de87edade2b884d329) |
| Freqtrade | [2026.9](https://github.com/freqtrade/freqtrade/releases/tag/2026.9) | [2 Ekim 2026](https://github.com/freqtrade/freqtrade/commit/5da169854b20adfce1e893ddac0f5270fb9b1d2d) |
| Jesse | Latest release sorgusu 404 | [1 Ekim 2026](https://github.com/jesse-ai/jesse/commit/417f8765225e3bfc12043d4b712f19fe15a3c078) |

Nautilus release candidate için Atom akışının updated alanı yayın tarihiyle aynı olmak zorunda olmadığından kesin RC yayın tarihi verilmedi.

### Sabit lisans kaynakları ve doğrulama sınırı

| Depo | Bu revizyonda okunan sabit lisans dosyası | Gözlenen lisans |
|---|---|---|
| Qlib | [be72549 / LICENSE](https://github.com/microsoft/qlib/blob/be725493eb1a6bbb42bf11b37aa7669f59610ff1/LICENSE) | MIT |
| TradingAgents | [8b22d43 / LICENSE](https://github.com/TauricResearch/TradingAgents/blob/8b22d43d01d9ddda5d686d093d5385884622f3de/LICENSE) | Apache-2.0 |
| LEAN | [33e3945 / LICENSE](https://github.com/QuantConnect/Lean/blob/33e3945f2faa95308d972dfd3d7b762f7743d1d5/LICENSE) | Apache-2.0 |
| FinRL | [e60e26e / LICENSE](https://github.com/AI4Finance-Foundation/FinRL/blob/e60e26e4870f00fbc704c6297edfd9d81816b3d3/LICENSE) | MIT |
| FinRL-X | [4409abe / LICENSE](https://github.com/AI4Finance-Foundation/FinRL-Trading/blob/4409abe925c904e570be78ebfb5e77ac3491dff8/LICENSE) | Apache-2.0 |
| Zipline Reloaded | [943010b / LICENSE](https://github.com/stefan-jansen/zipline-reloaded/blob/943010b9da848e317fc520de87edade2b884d329/LICENSE) | Apache-2.0 |
| Backtrader | [b853d7c / LICENSE](https://github.com/mementum/backtrader/blob/b853d7c90b6721476eb5a5ea3135224e33db1f14/LICENSE) | GPL-3.0 |
| Backtesting.py | [ca2e261 / LICENSE.md](https://github.com/kernc/backtesting.py/blob/ca2e2611621e472542ba90f7243a1fa06a7d7108/LICENSE.md) | AGPL-3.0 |
| vectorbt | [ceffc50 / LICENSE.md](https://github.com/polakowo/vectorbt/blob/ceffc501f2d37033a79dd86a9f883e69ec6977bd/LICENSE.md) | Apache-2.0 + Commons Clause |
| Jesse | [417f876 / LICENSE](https://github.com/jesse-ai/jesse/blob/417f8765225e3bfc12043d4b712f19fe15a3c078/LICENSE) | MIT çekirdek |
| FinMem | [be814aa / LICENSE](https://github.com/pipiku915/FinMem-LLM-StockTrading/blob/be814aa47970de9bf2fdd6a1d5a60ae5cf361b46/LICENSE) | MIT |

NautilusTrader ve Freqtrade'in yukarıdaki sabit commit'lerindeki lisans dosyaları bu revizyonda web aracıyla alınamadı. Resmî depo sayfaları sırasıyla LGPL-3.0 ve GPL-3.0 gösteriyor; kesin snapshot doğrulaması tamamlanmış sayılmıyor. Bir bağımlılık olarak seçilmeleri halinde sabit dosya, alt bileşen lisansları ve kullanım biçimi tekrar incelenir. Katalogdaki lisans adları bir ticari kullanım onayı veya genel hukuki sonuç değildir. [NautilusTrader](https://github.com/nautechsystems/nautilus_trader), [Freqtrade](https://github.com/freqtrade/freqtrade).

## Qlib ile tahmin araştırması

Python tabanlı Qlib, veri hazırlama, özellik üretimi, model eğitimi, tahmin, deney kaydı ve değerlendirmeyi bir iş akışında birleştiriyor. Bileşenlerinin ayrı kullanılabilmesi önemli: portföy ve emir katmanlarını benimsemeden tahmin katmanından yararlanılabilir. MIT lisanslı. Resmî veri seti README’de geçici olarak devre dışı gösteriliyor; kendi veri kaynağımızı hazırlamak daha sağlam bir başlangıç önerisidir. [README](https://github.com/microsoft/qlib), [veri dönüştürme dokümanı](https://qlib.readthedocs.io/en/latest/component/data.html#converting-csv-format-into-qlib-format).

Somut model desteği LightGBM ile sınırlı değil:

- LightGBM, XGBoost ve CatBoost.
- Linear modeller, MLP ve TabNet.
- LSTM, GRU, ALSTM ve TCN.
- Transformer ve Localformer.
- TRA, AdaRNN, HIST ve IGMTF gibi zaman ve ilişki yapısına yönelik araştırma modelleri.

Bunlar gerçek benchmark konfigürasyonlarıyla sunuluyor. Yayımlanan karşılaştırma tablosu Çin A hisseleri ve CSI300 üzerinden; sonuçlar ABD ETF’lerine aktarılmış başarı kanıtı değildir. [Model benchmarkları](https://github.com/microsoft/qlib/tree/main/examples/benchmarks), [benchmark açıklaması](https://github.com/microsoft/qlib/blob/main/examples/benchmarks/README.md).

Özellik tarafı:

- **Alpha158:** Mum yapısı, geçmiş getiri, ortalama, standart sapma, eğim, regresyon kalıntısı, fiyatın geçmiş aralıktaki konumu ve fiyat–hacim ilişkileri.
- **Alpha360:** Son 60 günlük fiyat ve hacim bileşenlerinin normalize edilmiş dizisi. İsmi 360 bağımsız veya kanıtlanmış algoritma anlamına gelmez.

Gerçek formüller sabit [loader.py](https://github.com/microsoft/qlib/blob/be725493eb1a6bbb42bf11b37aa7669f59610ff1/qlib/contrib/data/loader.py) içinde. Buradan kontrollü bir başlangıç özellik seti çıkarılabilir; 158 özelliği körlemesine eklemek gerekmiyor.

Önemli uyarlama: Sabit snapshot'ta Alpha158/360 etiketi `Ref($close,-2)/Ref($close,-1)-1`, yani t+1 kapanışından t+2 kapanışına tek seanslık getiridir. Varsayılan eğitim işlemcisi etiketleri aynı tarihte varlıklar arasında `CSZScoreNorm` ile standardize eder. Bu skor doğrudan beş seanslık mutlak getiri, beklenen basit getiri veya kalibre yükseliş olasılığı değildir. Ayrıca Qlib'in bu seçimi tek başına hedefin neden seçildiğini veya bizim yayın zamanımızla uyumunu kanıtlamaz. [Sabit handler.py](https://github.com/microsoft/qlib/blob/be725493eb1a6bbb42bf11b37aa7669f59610ff1/qlib/contrib/data/handler.py#L35).

Öneri: İlk pilotta Qlib bağımlılığı kurmadan özellik ve deney fikirlerini kullanmak. İleride kullanılırsa bilgi kesimi, yayın saati, başlangıç fiyatı, beş seanslık bitiş, kurumsal işlem sözleşmesi ve etiket normalizasyonu açıkça değiştirilir. Fiyat/getiri hedefi için çapraz kesit z-skorlu etiket varsayılanı taşınmaz; ayrı sıralama deneyi varsa o deneyin hedefi olarak tanımlanır. Kendi point-in-time arşivimiz ve olasılık kalibrasyonu korunur.

## TradingAgents ile Jev ve LLM entegrasyonu

Python ve LangGraph kullanıyor. Teknik, haber, duygu ve temel analiz rolleri; olumlu ve olumsuz görüş tartışması; trader, risk ve portföy yönetimi rolleri mevcut. Çoklu LLM sağlayıcısı, rapor kaydı, checkpoint ve tarih/sembol üzerinden değerlendirme akışı sunuyor. Araştırma sistemi olarak tanımlanıyor; çok sayıda agent bulunması sayısal tahmin doğruluğunu kanıtlamıyor. [README](https://github.com/TauricResearch/TradingAgents), [v0.5.2](https://github.com/TauricResearch/TradingAgents/releases/tag/v0.5.2).

### Doğrulanan Jev çağrısı

Jev entegrasyonu yalnızca yol haritası maddesi değil; gerçek API çağrısı var. İncelenen sabit snapshot: [8b22d43d01d9ddda5d686d093d5385884622f3de](https://github.com/TauricResearch/TradingAgents/commit/8b22d43d01d9ddda5d686d093d5385884622f3de).

Kod akışı:

- [post_screen.py](https://github.com/TauricResearch/TradingAgents/blob/8b22d43d01d9ddda5d686d093d5385884622f3de/tradingagents/agents/post_screen.py#L26): TypeSafe `/v1/systemone` uç noktasına `requests.post` çağrısı.
- Aynı dosyadaki `QUESTIONS`: Şirketle ilgililik için Noul, paylaşımın yönü için Choice.
- `jev_screen`: `TYPESAFE_API_KEY` varsa aktif; ilgisiz paylaşımları çıkarıyor ve bullish/bearish/neutral/unclear sayılarını üretiyor.
- [sentiment_analyst.py](https://github.com/TauricResearch/TradingAgents/blob/8b22d43d01d9ddda5d686d093d5385884622f3de/tradingagents/agents/analysts/sentiment_analyst.py#L64): Filtreyi StockTwits ve Reddit veri akışlarına geçiriyor.

Soru tasarımı ve sınırlı retry/backoff somut örneklerdir; **sürüm sabitleme örneği değildir**. `_DEFAULT_MODEL = "jev-latest"` hareketli bir alias'tır; ortam değişkeniyle değiştirilebilir. Yanıttaki gerçek model kimliği yalnızca debug loguna yazılır, özellik satırına kaydedilmez. Bir istek başarısız olursa `_unscreened` bütün gönderileri geçirir; warning logu ve metinsel “screening unavailable” notu üretir. Bu nedenle ikinci uzman görüşündeki “sessiz” tanımı harfi harfine doğru değildir, ancak filtrelenmiş ve filtrelenmemiş girdilerin aynı özellik anlamını koruduğu da varsayılamaz. [Sabit çağrı, yanıt ve hata kodu](https://github.com/TauricResearch/TradingAgents/blob/8b22d43d01d9ddda5d686d093d5385884622f3de/tradingagents/agents/post_screen.py#L24).

İlerideki ayrı Jev deneyi için somut sürüm kimliği sabitlenir; istenen ve dönen model kimliği, tam girdi/yanıt, hash, soru şeması, zamanlar ve hata durumu her özellik kaydına yazılır. Bounded retry başarısızsa Jev özelliği `missing/error` kalır; gönderiler başarılı sınıflandırma, nötr duygu veya “haber yok” yerine geçirilemez. Sayısal model kendi tanımlı fallback politikasını uygular. İlgililik/stance eşikleri finansal metin doğrulaması ve eğitim döneminde belirlenir. Bu sosyal ileti sınıflandırması beş günlük fiyat tahmininin doğruluğunu göstermez.

### Tarihsel doğruluk sınırları

“Point-in-time desteği hiç yok” demek yanlış olur. Snapshot’taki [StockTwits](https://github.com/TauricResearch/TradingAgents/blob/8b22d43d01d9ddda5d686d093d5385884622f3de/tradingagents/dataflows/vendors/stocktwits.py#L44) ve [Reddit](https://github.com/TauricResearch/TradingAgents/blob/8b22d43d01d9ddda5d686d093d5385884622f3de/tradingagents/dataflows/vendors/reddit.py#L38) kodları tarih filtresi uyguluyor ve kapsam eksikliğini işaretleyebiliyor. Ancak kaynaklar yakın zamandaki gönderileri sunuyor; filtrelemek üç yıllık tarihsel sosyal arşiv yaratmıyor.

Ayrı değerlendirilmesi gereken üç konu:

1. Tarih filtresinin uygulanması.
2. O tarihte erişilebilir veri sürümünün arşivlenmesi ve geçmiş pencerenin gerçekten kapsanması.
3. Ön eğitimli LLM veya Jev’in geçmiş sonucu biliyor olması.

Güncel README’deki point-in-time iyileştirme iddiaları, kodun tarih filtreleri ve kaynakların gerçek tarihsel kapsamıyla birlikte okunmalıdır. Kapsam açığı, bugünkü gönderilerin kesinlikle geçmişe sızdırıldığı anlamına gelmez.

Öneri: Jev ve çoklu agent sistemi ilk pilotun tahmin yoluna alınmaz. Metin değeri yeterli kapsama, kullanım haklarına, tam sürümün zamanında mevcut olmasına ve ileri yakalamaya dayanan ayrı bir deneyle sınanır. Jev yalnızca metinsel ilgililik/olay türü/ifade yönü sorularında araştırılır; tarih hesabı, sayısal piyasa benzerliği ve analog getiri ağırlıkları kodda kalır. Model eğitim kesimi doğrulanamıyorsa geçmiş LLM/Jev metin deneyleri yalnızca keşifsel kanıttır; finansal etki için belirleyici test ileri yakalanan veridedir.

## LEAN ile piyasa simülasyonu

C#/.NET çekirdek ve Python/C# algoritmaları kullanıyor; olay tabanlı backtest ve canlı işlemi ortak mimaride ele alıyor. ABD hisse/ETF, opsiyon ve diğer varlık sınıfları için kapsamlı referans. Apache-2.0 lisanslı çekirdek, QuantConnect’in bulut hizmetleri ve veri ürünleriyle aynı şey değil. [Repo](https://github.com/QuantConnect/Lean), [motor kavramları](https://www.quantconnect.com/docs/v2/writing-algorithms/key-concepts/algorithm-engine).

Somut alpha modelleri de var. [EmaCrossAlphaModel.py](https://github.com/QuantConnect/Lean/blob/master/Algorithm.Framework/Alphas/EmaCrossAlphaModel.py), EMA kesişimini zaman ufku taşıyan yukarı/aşağı Insight nesnesine çeviriyor. Bu sinyal üretimi ile emir gönderimini ayırmak için yararlı bir tasarım örneğidir; EMA sinyalinin bizim kapsamımızda başarılı olacağını göstermez.

Resmî LEAN CLI başlangıç dokümanı ücretli kuruluş üyeliği koşulu belirtiyor; yerel kullanımda verinin ayrıca edinilmesi gerektiğini söylüyor. Çekirdeğin açık kaynak olması bütün hazır CLI, bulut ve veri iş akışlarının ücretsiz olduğu anlamına gelmiyor. [CLI başlangıç koşulları](https://www.quantconnect.com/docs/v2/lean-cli/key-concepts/getting-started).

Öneri: İlk günlük tahmin MVP’sinin bağımlılığı yapmamak. İleride tahminler kurallı stratejilere dönüştürülürse işlem maliyeti, takvim ve emir davranışı altında sınamak için kısa listede tutmak.

## NautilusTrader ile simülasyon ve yürütme

Rust çekirdek, Python kontrol ve strateji katmanı; bar, tick, order book ve özel veriyle olay tabanlı simülasyon sunuyor. Interactive Brokers ve Databento adaptörleri resmî listede mevcut. Çoklu venue desteği, bütün kaynakların ücretsiz olduğu anlamına gelmez. [Repo](https://github.com/nautechsystems/nautilus_trader), [entegrasyonlar](https://nautilustrader.io/docs/latest/integrations/).

Sürüm ayrımı önemli: Güncel dokümantasyon v2 Rust/PyO3 mimarisini anlatıyor, v2 yayınları release candidate; `--pre` olmadan kurulan stable paket 1.x olabiliyor. Eski örnekler, yeni doküman ve stable paket rastgele birleştirilmemeli. [Kurulum ve sürüm açıklaması](https://nautilustrader.io/docs/latest/getting_started/installation/), [v2.0.0rc5](https://github.com/nautechsystems/nautilus_trader/releases/tag/v2.0.0rc5).

Proje yol haritası yerleşik AI/ML araçlarını ve dağıtık orchestration’ı çekirdek kapsamı dışında tutuyor; hazır tahmin modeli beklenmemeli. [Roadmap](https://github.com/nautechsystems/nautilus_trader/blob/develop/ROADMAP.md).

Öneri: Günlük beş günlük tahmin için ilk faza eklememek; ileride tick, order book veya canlı işlem gerekirse LEAN’e alternatif değerlendirmek. LGPL-3.0 koşulları seçilen entegrasyon şekli üzerinden incelenmeli.

## Diğer sistemlerin somut katkıları

### FinRL

Stable Baselines 3 üzerinden A2C, DDPG, PPO, SAC ve TD3 örnekleri; Yahoo, Alpaca ve WRDS gibi veri işlemcileri bulunuyor. Güncel README kendisini eğitim ve araştırma çerçevesi olarak ayırıp yeni üretim çalışmalarını FinRL-X’e yönlendiriyor. RL hedefi portföy kararının ödülü; kalibre edilmiş beş günlük yükseliş olasılığıyla aynı görev değil. [README](https://github.com/AI4Finance-Foundation/FinRL).

### FinRL-X

FMP, Yahoo ve WRDS veri katmanı, `bt` üzerinden backtest ve Alpaca üzerinden yürütme; stock selection → allocation → timing → risk dönüşümlerini ortak ağırlık formatında birleştiriyor. ETF rotasyonu, momentum/rejim ve ML seçim örnekleri mimari referans olarak incelenebilir. [Repo](https://github.com/AI4Finance-Foundation/FinRL-Trading).

Sabit `4409abe` snapshot'ı gerçek emir akışı içerir, ancak güvenilir yürütme örneği olarak kopyalanmamalıdır. `_get_current_price` quote hatasında `100.0` döndürür; aynı sonuç pay adedi, emir değer limiti ve turnover hesabında kullanılır. `_gvkey_to_ticker` yalnızca beş örnek şirketi eşler; eşleşmeyeni olduğu gibi döndürür. Bu iki placeholder fiyat veya kimlik başarısızlığını geçerli işlem girdisine çevirebilir. [Sabit trade_executor.py](https://github.com/AI4Finance-Foundation/FinRL-Trading/blob/4409abe925c904e570be78ebfb5e77ac3491dff8/src/trading/trade_executor.py#L217).

Karar: FinRL-X yalnızca kavramsal veri/strateji/yürütme ayrımı için tutulur; hiçbir yürütme kodu bu pilotun kapsamına girmez. Gelecekte fiyat alınamıyorsa emri durduran davranış, zamanına göre doğru kalıcı varlık kimliği, quote güncelliği ve bağımsız risk kontrolleri ayrıca doğrulanır. “Production-ready” geliştiricinin iddiasıdır; somut placeholder'lar bu iddiayı bağımsız güvenilirlik kanıtı olarak kullanmamamızı gerektirir. `v1.0.0` sürümünün varlığı doğrulanmıştır; önceki “sürüm tarihi raporlanmadı” kaydı düzeltilmiştir. [Sürüm kaydı](https://github.com/AI4Finance-Foundation/FinRL-Trading/releases/tag/v1.0.0).

### Zipline Reloaded

Çoklu varlık ve günlük olay tabanlı araştırma, takvim, bundle yapıları ve maliyet simülasyonu için uygun. Hazır ML tahmin platformu değil. README’deki örnek Nasdaq/Quandl verisini kullanıyor; güncel veri kapsamı garantisi saymadan kendi ingestion hattımız kurulmalı. [Repo](https://github.com/stefan-jansen/zipline-reloaded).

### Backtrader

CSV/Pandas veri, çoklu feed/strateji, göstergeler ve emir simülasyonu açısından geniş. README’deki IB entegrasyonu eski IbPy bağımlılığına dayanıyor; eski canlı adaptörlerin güncel çalışabilirliği ayrıca kontrol edilmeli. Düşük son commit etkinliği nedeniyle yeni ürünün çekirdeği için ilk tercih önerilmiyor. [README](https://github.com/mementum/backtrader).

### Backtesting.py

OHLCV verilen tek varlıkta SMA kesişimi gibi referanslar ve parametre denemeleri için kolay. SAMBO optimizasyonu ve grafik çıktıları var; genel amaçlı çoklu varlık ve broker platformu gibi sunulmamalı. AGPL-3.0 koşulları ürün entegrasyonunda dikkate alınmalı. [Doküman](https://kernc.github.io/backtesting.py/), [repo](https://github.com/kernc/backtesting.py).

### vectorbt

NumPy/Numba ve güncel sürümde opsiyonel Rust ile çok sayıda parametre ve sinyal kombinasyonunu hızlı inceliyor. PRO, ek veri ve backtest özelliklerini ayrıca sunuyor. [Repo](https://github.com/polakowo/vectorbt), [PRO farkları](https://vectorbt.dev/getting-started/upgrade/).

Sabit `ceffc50` lisansı Apache-2.0 **ile birlikte Commons Clause** koşulunu içerir. Değerinin tamamı veya önemli kısmı yazılım işlevinden gelen ücretli ürün/hizmeti `Sell` tanımıyla kısıtlar; yalnızca “Apache-2.0” diye kaydedilemez. Ticari kullanım alanını sınırladığı için OSI'nin Açık Kaynak Tanımı'na uygun açık kaynak olarak sınıflandırılmaz; kaynak kodu erişilebilir, ek kullanım kısıtlı yazılım olarak kaydedilir. Bu, her türlü ticari kullanımın otomatik yasak olduğu sonucu değildir; somut kullanımın `Sell` kapsamına girip girmediği ayrı değerlendirilir. İlk pilotun bağımlılığı değildir. [Sabit LICENSE.md](https://github.com/polakowo/vectorbt/blob/ceffc501f2d37033a79dd86a9f883e69ec6977bd/LICENSE.md), [OSI tanımı, madde 6](https://opensource.org/osd).

### Freqtrade

Aktif kripto botu. FreqAI adaptif model eğitimi ve lookahead-analysis/recursive-analysis araçları öğretici. CCXT ve kripto market varsayımları ABD hisselerine hazır destek olarak yorumlanmamalı. [Repo](https://github.com/freqtrade/freqtrade), [FreqAI](https://www.freqtrade.io/en/stable/freqai/), [lookahead analizi](https://www.freqtrade.io/en/stable/lookahead-analysis/).

### Jesse

Momentum ve GoldenCross, Optuna, Monte Carlo, kural anlamlılığı ve scikit-learn uyumlu ML araştırma akışı bulunuyor. Çekirdek MIT olsa da live/paper işlem resmî lisanslı plugin üzerinden. [Repo](https://github.com/jesse-ai/jesse), [live plugin](https://docs.jesse.trade/docs/livetrade).

Resmî entegrasyon sayfası Alpaca hisse sürücüsünü geliştirme aşamasında gösteriyor; bugün ABD hisse çekirdeği seçmek için yeterli gerekçe değil. [Desteklenen piyasalar](https://docs.jesse.trade/docs/supported-exchanges/).

### FinMem

Katmanlı hafıza, önem/zaman/ilgili olma sıralaması ve geçmiş karar üzerinden reflection fikirleri olay arşivine uyarlanabilir. Train modunun esas işi agent hafızasını doldurmak; sayısal fiyat modelinin ağırlıklarını eğitmekle karıştırılmamalı. Broker ve üretim veri altyapısı yerine akademik referanstır. [Repo](https://github.com/pipiku915/FinMem-LLM-StockTrading), [makale](https://arxiv.org/abs/2311.13743).

## Referansların önceliği ve kapsam kapıları

Bu liste bağımlılık veya ürün yol haritası taahhüdü değildir. Herhangi bir repo benchmark'ı, çalışan API çağrısı veya bakım göstergesi kendi tahmin başarımızın kanıtı sayılmaz. Deney kimlikleri, hedefler ve geçiş eşikleri [deney planındaki tek kayıt](deney-plani.md) tarafından yönetilir.

1. **Kendi küçük Python pilotumuz:** İzinli veri, sabit hedef, kronolojik referanslar, güç analizi, değiştirilemez tahmin ve sürümlü sonuç kayıtları, statik rapor.
2. **Qlib:** İhtiyaç olduğunda özellik ve rolling deney fikirleri; varsayılan etiket/normalizasyon ürün sözleşmesine taşınmaz.
3. **TradingAgents:** Ayrı sonraki metin araştırması için soru ve çağrı örnekleri; hareketli alias ve filtrelenmemiş fallback kopyalanmaz.
4. **LEAN / NautilusTrader / FinRL-X:** Mevcut pilot kapsamı dışında kavramsal kaynaklar; yeni işlem kapsamı ve bağımsız doğrulama olmadıkça kurulum veya entegrasyon yok.

NautilusTrader, ileride düşük gecikmeli yürütme gerekirse LEAN’e alternatif. FinMem’in olay hafızası ve Freqtrade’in lookahead/recursive kontrol mantığı bağımsız fikir kaynakları olarak incelenebilir.

İlk fazda bu katalogdaki sistemler kurulmaz veya fork edilmez. Sayısal katkı yeterli güçle ölçülemiyorsa “başarılı” veya “başarısız” ürün iddiası yerine sonuçsuz araştırma kararı verilir. Metin katmanı ve yürütme ancak kendi kapsam kapıları geçildiğinde ayrı çalışma olur.

Bu araştırma repo işlevlerini ve tekrar kullanım adaylarını inceler. Hazır stratejilerin ABD hisse/ETF kapsamımızda gelecekte başarılı olacağını göstermez. LLM demo getirileri, örnek benchmarklar ve geliştiricilerin üretim iddiaları için kendi kronolojik testlerimiz ve ileriye dönük tahmin kayıtlarımız gerekli.

İlgili belgeler: [Matematik algoritmaları](matematik-algoritmalari.md), [deney planı](deney-plani.md), [MVP teknik planı](../MVP-TEKNIK-PLAN.md).
