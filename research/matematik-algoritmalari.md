# ABD hisse ve ETF tahmini için matematiksel algoritmalar

**Güncel MVP geliştirme ve çalışma ortamı: Hetzner (`167.235.158.118`).** Backend/model/veri işleri uzaktaki sunucudadır; Mac gözlem uygulaması ve uzak erişim içindir. Aşağıdaki tarihsel yöntem kataloğu bu yerleşimi değiştirmez; sunucu bağlantısı henüz yapılmadı.

Güncel kapsam notu: 8 Ekim 2026. Bu dosya **2 Ekim sayısal P1 yöntem kataloğudur; güncel canlı trading MVP'nin uygulama sözleşmesi değildir**. Python/Ridge/GARCH, quantile hedefi, konfigürasyon bütçesi, sabit cutoff ve istatistiksel değerlendirme bu MVP'de uygulanmayacak. Kullanıcı değerlendirmeyi sonraya bıraktı. Yeni model/metrik ekleme taahhüdü yoktur.

Güncel matematik ihtiyacı yalnız fiyat açıklığı/zaman kontrolü, decimal para/lot, maliyet, cash rezervi, T+1 ve özkaynak hesabıdır; sözleşmeleri [MVP planında](../MVP-TEKNIK-PLAN.md#4-emir-ve-gerçekleşme-sözleşmesi) bulunur. Karakterler ve kaynak kapsamı [ürün planında](deney-plani.md) tutulur. Önceki belgenin değişmemiş kopyası [arşivdedir](archive-2026-10-08/matematik-algoritmalari.md).

## 2 Ekim 2026 tarihsel çalışma

Bu belge, iki uzman değerlendirmesi sonrasında daraltılan **sayısal araştırma pilotunun** matematiksel sözleşmesini ve ertelenmiş algoritma kataloğunu içerir. İlk teslimat günlük Python batch işi, veri doğrulama, referans modeller, korumalı değerlendirme ve değiştirilemez tahmin kaydıdır. Model üstünlüğü, Jev/LLM katkısı ve ticari yarar henüz gösterilmemiştir. Emir, pozisyon, portföy, FastAPI ve ürün arayüzü ilk kapsamda yoktur.

**Tek deney kaynağı [deney-plani.md](deney-plani.md)'dir.** İlk tur baselines dahil en çok 15 toplam önceden kayıtlı konfigürasyon ve en çok 15 tanımlı model girdisiyle sınırlıdır; aynı ailenin farklı pencere/parametre/kalibrasyon sürümleri de bu bütçede ayrı denemedir. Bu belgede denklem bulunması o algoritmayı uygulama taahhüdü yapmaz. İlk modeller Ridge ortalama ve düzenlileştirilmiş lojistik yön; belirsizlik referansı EWMA-ampirik dağılım, koşullar geçerse ilk oynaklık rakibi GJR-GARCH'tır. Sığ boosting veya doğrudan quantile adayları yalnızca kayıtlı bütçede açılır. Kalman/HMM ikinci tur; haber, Jev, analoglar, derin ağlar ve spread modelleri ayrı geçiş koşullarına bağlıdır.

## Tahmin hedefleri ve bilgi sınırı

Her tahmin origin'i o için yayın saati **08:45 America/New_York (ET)**, bilgi kesimi **08:30 ET**'dir. d1, zamanında yayımlanan tahminden sonraki ilk düzenli seans; d5 dahil beşinci düzenli seanstır. ET sabit UTC ofseti değildir; yaz saati, tatil ve erken kapanış borsa takviminden hesaplanır. Geciken tahmin aynı origin'e sessizce yazılmaz: başarısız/geç koşum olarak kayda girer.

`OPEN(d1)` ve `CLOSE(d5)`, doğrulanmış düzenli seans içindeki ilk/son uygun bar fiyatıdır; resmî açılış/kapanış ihalesi ya da elde edilebilir işlem fiyatı garantisi değildir. Sağlayıcının günlük O/C alanı tek başına seans kanıtı sayılmaz; dakika barları/seans filtresiyle doğrulama gerekir. Günlük hacim alanı uzatılmış seansı içerebilir; kullanılıyorsa önceki tamamlanmış günün **sağlayıcı günlük hacmi** olarak adlandırılır, düzenli seans hacmi diye sunulmaz.

Ana etiket d1 açılışında bir birim satın alıp d5 kapanışında tutmanın, brüt nakit temettü haklarını içeren **basit servet getirisi**dir:

$$
R_{i,o}^{(5)}=\frac{q_{i,d5}\,C_{i,d5}+D_{i,o}^{\mathrm{hak}}}{O_{i,d1}}-1.
$$

O başlangıç açılış fiyatı, C bitiş kapanış fiyatı; q bölünmelerden sonraki birim sayısı, D elde tutma sırasında hak kazanılan brüt nakit dağıtım/alacak toplamıdır. Temettü yeniden yatırılmaz; vergi ve işlem maliyeti hariçtir. d1 ex-date ise açılışta alan yatırımcı o dağıtıma hak kazanmaz; d1'den sonra d5'e kadar ex-date gerçekleşen haklar kapsama girer. Dağıtım tutarı, hak anındaki birim sayısıyla hesaplanır; ödeme tarihi daha sonra olsa bile hak alacak olarak etiketlenir. Spin-off, birleşme ve çözülemeyen kurumsal eylemler `pending/unresolved` outcome olur; sıfır getiriye çevrilmez. Etiket ancak d5 fiyatı ve ilgili eylemler doğrulandıktan sonra olgunlaşır, sonradan gelen düzeltme outcome sürümü üretir.

Log belirsizlik modelinde yardımcı değişken:

$$
Y_{i,o}^{(5)}=\log(1+R_{i,o}^{(5)}),\qquad R=e^Y-1.
$$

Bu tanım, önceki kapanıştan başlayıp beş close-to-close günlük log getiriyi toplamakla aynı değildir: ilk gece fiyat farkını, seans başlangıcını ve nakit temettü yeniden yatırımını farklı ele alır. R=−1 tam kayıp halinde log hedef −∞ olur; bu outcome basit-return değerlendirmesinden atılmaz, log modeli için ayrı atom/hata politikası gerekir. Günlük close-to-close getiriler r, geçmiş momentum ve oynaklık girdileridir; doğrudan etiketi değiştirmez. Dönemlik hedefe taşınan her günlük model için ayrı ufuk adaptörü zorunludur.

| Çıktı | Matematiksel anlamı | Birincil ölçümü ve rolü |
| --- | --- | --- |
| Alt/üst nominal quantile | Q0.10(R), Q0.90(R) | **Birincil pilot sonlanımı:** iki pinball kaybının ortalaması |
| Ortanca | Q0.50(R) | MAE; ortalama yerine gösterilmez |
| Ortalama | E[R \| bilgi_o] | MSE/RMSE; Ridge doğrudan basit R'ye eğitilir, ayrı ikincil hipotez |
| Yön | p = Pr(R > 0 \| bilgi_o) | Brier; lojistik çıktı ve geçmiş yükseliş sıklığına karşı ayrı ikincil hipotez |
| Oynaklık | Tanımlı log/getiri residual varyansı | QLIKE gibi ayrı tanısal ölçüm; yön kanıtı değildir |
| Göreli getiri/sıralama | R_i − R_referans veya rank | İlk pilot dışındaki ayrı hedef ve evren deneyi |
| Rejim | Pr(S_o=j \| bilgi_o) | İkinci tur özellik adayı, kendi başına boğa/ayı olasılığı değildir |

Yön etiketi log ve basit dönüşümde aynıdır. Artan dönüşüm nedeniyle quantile tam dönüştürülür:

$$
Q_\tau(R)=\exp(Q_\tau(Y))-1.
$$

**Ortalamalar bu kurala uymaz:** E[R]=E[e^Y]−1, genellikle e^{E[Y]}−1 değildir. Eğer *beş günlük toplam Y'nin kendisi* gerçekten N(μ,v) ise E[R]=e^{μ+v/2}−1; günlük Gaussian innovations bu toplamın Gaussian veya üstel momentinin sonlu olduğunu tek başına göstermez. Degenerate olmayan Student-t log getiri Y=μ+σTν (σ>0) için polinom sağ kuyruk nedeniyle **E[e^Y]=∞**; ν>2 ile sonlu log varyansı bunu çözmez. Student-t log/GARCH yolundan sonlu Monte Carlo ortalamasını “beklenen basit getiri” diye sunmak yanlıştır. Bu pilot ortalamayı doğrudan R'den öğrenir; Student-t yalnızca uygun log-varyans ve quantile özetlerinde kullanılabilir. Moment sonucu, [SciPy Student-t yoğunluğundan](https://docs.scipy.org/doc/scipy/reference/generated/scipy.stats.t.html) matematiksel çıkarımdır; [arch Student-t standardizasyonu](https://arch.readthedocs.io/en/latest/univariate/generated/arch.univariate.StudentsT.html) üstel moment yaratmaz.

`bilgi_o`, event time yerine **available_at ≤ 08:30 ET** olan doğrulanmış sürümlerle kurulur. Sonradan düzeltilmiş bar, yeniden yazılmış haber, sonraki SEC amendment veya tüm örneklem üzerinde öğrenilmiş dönüşüm geçmişte biliniyormuş gibi kullanılamaz. Kurumsal eylemin sonucu hesaplamada kullanılması, o eylemin duyurusunun tahmin girdisi olarak o an bilindiği anlamına gelmez. İlk turda lisans/erişim zamanı kanıtlanmamış makro ve metin girdisi yoktur. Model ve kalibratör yalnızca fit anına kadar olgunlaşmış etiketleri görür; kaynak eksikliği nötr haber ya da sıfır oynaklık değildir.

## Model ailelerinin karşılaştırılması

Aşağıdaki uygunluk yargıları araştırma önceliğidir; evrensel örnek sayısı veya beklenen performans sıralaması değildir. Etkin kapsam, yalnızca deney kaydında açılan adaylardır.

| Aile | Öğrendiği yapı | Güncellenen kapsam kararı |
| --- | --- | --- |
| Sıfır, olgunlaşmış ortalama/medyan/yükseliş sıklığı | Basit geçmiş referans | Zorunlu; sıfır getirinin kendisi yön olasılığı vermez |
| EWMA + ampirik standardize residual | Koşullu aralık ölçeği | İlk belirsizlik referansı; doğrudan OPEN→CLOSE hedefinde kalibre |
| Ridge ve lojistik | Düzenlileştirilmiş ortalama/yön ilişkisi | İlk sayısal modeller; en çok 15 kayıtlı girdi |
| GJR-GARCH | Asimetrik günlük varyans | İlk ölçek rakibi; günlük varyans doğrudan hedef varyansı sayılmaz |
| AR(1) | Küçük otoregresif kontrol | İlk registry dışında; gerekçeli yeni turda referans olabilir, ARMA/order taraması ertelendi |
| Küçük quantile / sığ boosting | Doğrusal olmayan nominal quantile veya yön | Bütçede önceden tanımlı rakip açılırsa; otomatik kapsam değil |
| ElasticNet | L1/L2 ile değişken seçimi | Ridge yerine serbest geniş grid yok; yeni tur kaydı gerekir |
| Kalman ve HMM/Markov switching | Gizli durum, zamanla değişen beta | İkinci tur; nedensel filtreleme ve yeni test verisi şart |
| ARIMA/ARIMAX, VAR/faktör | Seri veya varlıklar arası dinamik | Gerekçe ve bütçesi olan sonraki araştırma |
| Cointegration ve OU | Spread dengeye dönüşü | Ayrı spread ürünü/etiketi; pilot dışı |
| Analog ve hiyerarşik olay modeli | Olay benzerliği / kısmi pooling | Geniş olay evreni, metin hakları ve kontaminasyon kapısından sonra |
| Kernel/GP, DLinear | Benzerlik ya da doğrusal dizi haritası | Ertelendi; ilk Ridge/lag kontrollerinden farklı değer göstermeli |
| N-BEATS/N-HiTS/LSTM/Transformer | Esnek sıralı yapı | Veri ve güç gerekçesi oluşmadan açılmaz |
| Chronos/TimesFM | Ön eğitimden taşınan örüntü | Kontaminasyon sınırı belli ayrı sanity deneyi; ilk çekirdek değil |
| Ensemble/adaptif conformal | Birleşim / aralık düzeltmesi | Ayrı kayıtlı yöntem; ilk turdan otomatik ek katman değil |
| Reinforcement learning | Eylem ve ödül politikası | Tahmin-only kapsam dışında |

## Referans tahminleri ve momentum regresyonu

Önce sıfır getiri, geçmişte gözlenen yükseliş sıklığı ve EWMA oynaklık gibi referanslar kurulur. Yön referansını otomatik olarak yüzde 50 kabul etmek yerine yalnızca geçmiş olgunlaşmış etiketlerden hesaplamak gerekir.

Momentum bir model ailesinin girdisi veya açık bir hipotez olabilir:

$$
M_{L,t}=\sum_{j=0}^{L-1}r_{t-j},\qquad
R_{i,o}^{(5)}=b+\beta_5M_{5,t}+\beta_{20}M_{20,t}
+\beta_{60}M_{60,t}+\epsilon_t.
$$

Pozitif katsayı devamlılık, negatif katsayı tersine dönüş ilişkisine işaret eder. Bu ilişkiyi varsaymak yerine eğitim döneminde öğreniriz. Momentum, hareketli ortalama, RSI ve MACD gibi göstergeler tek başına tahmin doğruluğu sağlamaz; modele sunulan geçmiş özetleridir.

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** 5, 20 ve 60 işlem günlük momentum; bunları geçmiş oynaklığa bölerek normalize eden ikinci bir sürüm. Katsayıları Ridge ile küçültmek. 252 günlük veriyle hesaplanan özellik kullanılacaksa ilk yıl ısınma dönemi sayılır; üç yıllık arşivde eğitim satırı azalır.

**Başarısızlık koşulları:** Trendin ani kırılması, yatay dönemlerde gürültünün trend gibi yorumlanması, haber kaynaklı sıçramalar. AQR makalesi uzun vadeli momentum için araştırma dayanağı sunar; makalenin futures, çok yıllı dönem ve aylık ufuktaki bulguları beş günlük ETF tahminimize performans kanıtı değildir. [Time Series Momentum](https://www.aqr.com/Insights/Research/Journal-Article/Time-Series-Momentum).

## ARIMA ve ARIMAX

ARIMA, bir serinin geçmiş değerleri ve geçmiş tahmin hatalarıyla gelecekteki değerini açıklar:

$$
\phi(L)(1-L)^d y_t=c+\theta(L)\epsilon_t.
$$

L gecikme operatörüdür. p geçmiş değer sayısını, d fark alma derecesini, q hata gecikmelerini belirtir. Getiriler üzerinde d=0, log fiyat üzerinde fark alma ihtiyacı ayrı ayrı değerlendirilir. Fiyat seviyesinde düşük hata elde etmek, günlük getiri yönünü iyi tahmin ettiğimiz anlamına gelmez; son fiyatı tekrar etmek güçlü bir referanstır.

Dış değişkenli uygulamada statsmodels SARIMAX regresyon ve ARIMA hatalarını şu biçimde temsil eder:

$$
y_t=\beta^\top x_t+u_t,\qquad
\phi(L)(1-L)^d u_t=\theta(L)\epsilon_t.
$$

Bu, aynı sayıları doğrudan gecikmeli y denklemine koyan her ARX uygulamasıyla aynı parametre anlamına gelmez. Kullanılan form ve intercept yorumu kaydedilmelidir. [SARIMAX model tanımı](https://www.statsmodels.org/stable/generated/statsmodels.tsa.statespace.sarimax.SARIMAX.html).

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** Getiride AR(1), AR(2), ARMA(1,1). Küçük modelleri aylık yeniden öğrenme; her gün durumu güncelleme. İç eğitim penceresinde BIC yalnızca aday daraltmada, son seçim kronolojik tahmin hatasında. Günlük haftalık mevsimselliği otomatik olarak eklememek.

**Kritik uygulama konusu:** Çok adımlı ARIMAX tahmininde gelecek x değerleri gerekiyorsa bunları gerçek gelecekteki piyasa getirileri veya sonradan yayımlanmış verilerden veremeyiz. Takvim gibi önceden bilinen değişkenler kullanılabilir; diğerleri ayrıca tahmin edilmeli veya açık senaryo olmalıdır. Beş günlük doğrudan denetimli model R_{i,o}^{(5)} = f(x_{i,o}) bu gereksinimi önleyebilir.

**Veri ve varsayım:** Az sayıda parametre için yüzlerce gözlem başlangıç deneyine yeterli olabilir; başarıyı garanti etmez. Bağımlılıkların pencere içinde yeterince istikrarlı olması gerekir.

**Başarısızlık:** Günlük getiride zayıf otokorelasyon, yapısal kırılma ve yanlış dış değişken zamanı. Haber içeriği ARIMA tarafından kendiliğinden anlaşılmaz; Jev özellikleri x’e eklenebilir.

## Kalman filtresi ve durum uzayı

Genel doğrusal durum modeli:

$$
s_t=A s_{t-1}+w_t,\quad w_t\sim N(0,Q),
\qquad y_t=H_t s_t+v_t,\quad v_t\sim N(0,R).
$$

Gizli durum s, gözlem y’den her gün güncellenir. Filtrenin temel adımları:

$$
\hat s_{t|t-1}=A\hat s_{t-1|t-1},\quad
P_{t|t-1}=AP_{t-1|t-1}A^\top+Q,
$$

$$
K_t=P_{t|t-1}H_t^\top
(H_tP_{t|t-1}H_t^\top+R)^{-1},
\quad
\hat s_{t|t}=\hat s_{t|t-1}+K_t(y_t-H_t\hat s_{t|t-1}).
$$

Kovaryansın güncellemesi, sayısal olarak daha kararlı Joseph biçiminde:

$$
P_{t|t}=(I-K_tH_t)P_{t|t-1}(I-K_tH_t)^\top+K_tRK_t^\top.
$$

Bu biçim yuvarlama altında simetri/pozitif yarı tanımlılık sorunlarını azaltır; toleranslar, boyutlar ve kovaryans kontrolleri uygulamada doğrulanır. [FilterPy resmi uygulaması](https://filterpy.readthedocs.io/en/latest/_modules/filterpy/kalman/kalman_filter.html) aynı güncellemeyi kullanır. Kalman ilk turda yoktur; ikinci turda küçük bir filtrelenmiş beta özelliği ancak ilk pilot ve yeni deney kaydı gerekçe sağlarsa açılır.

K, yeni gözleme verilen ağırlıktır. Q büyükse durum daha hızlı değişebilir; R büyükse gözlemi daha gürültülü kabul ederiz.

Ertelenmiş iki uygulama seçeneği:

- Log fiyatta yerel seviye ve eğim: gizli durum fiyat seviyesi ve trenddir. Beş günlük tahmin eğimi ileri taşır.
- Zamana göre değişen piyasa/ sektör betası: r_i,t = α_t + β_t r_piyasa,t + ε_t. Bu modelin ürettiği güncel beta ve residual, doğrudan beş günlük tahmin modeline özellik olabilir.

Durum uzayı ve Kalman filtreleme resmi statsmodels uygulamalarında mevcut. [Durum uzayı belgeleri](https://www.statsmodels.org/stable/statespace.html).

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** Bir veya iki gizli durum; Q/R’yi küçük bir aday setinde kronolojik olarak seçmek. Gürültü matrislerini sırf tam örneklem hata düşük olsun diye serbestçe büyütmemek.

**Başarısızlık:** Modelin piyasa gürültüsünü gizli trend sanması, doğrusal/Gaussian varsayımının sıçramalarda zayıflaması, Q/R’nin kötü seçilmesi. Filtrenin uyarlanabilir olması trendin öngörülebilir olduğu anlamına gelmez.

**Sızıntı:** Canlı özellik için s_t|t kullanılmalıdır. s_t|T biçimindeki smoothing T>t ise gelecekteki gözlemleri kullanır. Eğitim geçmişini tüm dönem üzerinde smooth edip test girdilerine taşımak geçersizdir.

## Ridge, lojistik ve ertelenmiş ElasticNet

İlk ortalama modeli Ridge, doğrudan basit R etiketi üzerinde kare kaybı ve L2 cezasıyla eğitilir. ElasticNet aşağıdaki daha genel L1/L2 formudur; ilk modelin yerine geçmez:

$$
\min_{\beta,b}\frac{1}{2n}\sum_i
(R_i-b-x_i^\top\beta)^2+
\lambda\rho\|\beta\|_1+
\frac{\lambda(1-\rho)}{2}\|\beta\|_2^2.
$$

ρ=0 Ridge, ρ=1 Lasso ucudur. Standardizasyon ve eksik değer işlemleri yalnızca eğitim bölümünde öğrenilir. Model katsayılarının küçülmesi kısa geçmişte aşırı uyumu azaltabilir. [ElasticNet](https://scikit-learn.org/stable/modules/generated/sklearn.linear_model.ElasticNet.html).

Yön modeli:

$$
p_t=\frac{1}{1+\exp[-(b+x_t^\top\beta)]},\qquad
D_{i,o}=\mathbf1\{R_{i,o}^{(5)}>0\}.
$$

Eğitim kaybı binary log loss ve düzenlileştirme cezasıdır. Çıktı eşiklenmiş karar yerine p olarak saklanır. [Lojistik regresyon](https://scikit-learn.org/stable/modules/generated/sklearn.linear_model.LogisticRegression.html).

**İlk tur sözleşmesi:** Ridge ve düzenlileştirilmiş lojistik için en çok 15 önceden tanımlı girdi; fiyat/oynaklık/hacim ve kesimden önce bilinen az sayıda olay takvimi bayrağı. Kesin listeler ve cezalar [deney kaydındadır](deney-plani.md). Serbest ElasticNet grid'i, 20–40 gösterge, otomatik etkileşimler ve veri sonrasında soru/özellik keşfi bu turda yoktur. Eksik/güncel olmayan girdiler kayıtlı hata/fallback politikasını izler. Her pencere, ceza veya kalibrasyon değişikliği aday bütçesini tüketir.

**Panel model:** Birden fazla varlık ortak modeli besleyebilir. Varlık kimliği ve sektör özellikleri ancak önceden kayıtlı toplam girdi/konfigürasyon bütçesinde yer alıyorsa eklenebilir; ancak aynı günkü ilişkili varlıklar bağımsız kanıt sayılmaz. Gün bazında normalize edilmiş özellikler yalnızca o an kapsamda bulunan varlıkları kullanmalıdır.

**Başarısızlık:** Etkileşimlerin doğrusal modele sığmaması, değişen katsayılar, birbirine benzeyen onlarca teknik gösterge. Katsayı istikrarı ve özellik çıkarma deneyleri kontrol edilmelidir.

**Ertelenmiş Jev bağlantısı:** Metin kapısı geçildikten sonra beklenti düşürme, düzenleyici risk, haber yeniliği ve şirketle ilişki gibi skorlar x’e eklenir. Yalnızca metin özellikleri ile fiyat özelliklerinin katkıları ayrı deneylerde değerlendirilir.

## Gradient boosting ve quantile tahmini

LightGBM ve benzeri boosting yöntemleri birden fazla ağacın katkısını toplar:

$$
F_M(x)=F_0(x)+\eta\sum_{m=1}^{M}h_m(x).
$$

Yeni ağaç, kaybın kalan hatasını azaltır. Yön için log loss; ortalama getiri için L2; ayrı robust location işlevi için Huber (koşullu ortalama veya medyan diye otomatik adlandırılmaz); quantile için pinball loss seçilebilir. LightGBM resmi olarak binary ve quantile hedeflerini destekler. [LightGBM amaçları ve parametreleri](https://lightgbm.readthedocs.io/en/latest/Parameters.html).

Quantile kaybı:

$$
\rho_\tau(u)=u(\tau-\mathbf1\{u<0\}),\qquad
\hat Q_\tau(x)=\arg\min_f\sum_i\rho_\tau(R_i-f(x_i)).
$$

τ=0.1,0.5,0.9 için ayrı modeller Q10, medyan ve Q90 üretebilir. Q10–Q90 aralığı hedef olarak yüzde 80 kapsama taşır; bunu gerçekten sağlayıp sağlamadığını ileri testte ölçmek gerekir. Ayrı modellerde quantile crossing olabilir; sıralama düzeltmesi uygulamak kalibrasyon sorununu tek başına çözmez.

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** Ağaç derinliği 2–4, yaprak sayısı 4–16, min leaf örneği 100–300, öğrenme oranı 0.03–0.1; bunlar küçük panel veri için konservatif deney aralıklarıdır. Ağaç sayısını kronolojik validation ve early stopping belirler. Varlık başına model ile ortak modeli karşılaştırmak.

**Avantaj:** Oynaklık yüksekken momentumun farklı etkisi gibi etkileşimleri elle her çarpım terimini yazmadan öğrenebilir. Fiyat ve Jev özelliklerini aynı modelde birleştirebilir.

**Başarısızlık:** Eğitimde nadir koşulları ezberlemek, piyasalar arası taşınmayan ilişkiler, dağılım dışı örnekler ve çok fazla hiperparametre araması. Tahminleri eğitim aralığının ötesinde doğrusal biçimde extrapolate edeceği varsayılmamalıdır. Feature importance nedensellik değildir.

## EWMA ve GARCH ailesi

EWMA günlük residual varyans referansı:

$$
\sigma_{t+1}^2=\lambda\sigma_t^2+(1-\lambda)\epsilon_t^2.
$$

λ ve lookback seçenekleri yalnızca deney kaydındaki sınırlı adaylardır; 0.94 gibi tarihsel değerler evrensel doğruluk sağlamaz. Günlük ölçek, **doğrudan hedefe** ait olgunlaşmış, gerçek out-of-sample residual dağılımıyla eşleştirilir. Örneğin basit getiride location-scale tahmini:

$$
R_{i,o}^{(5)}=\hat\mu_{i,o}+s_{i,o}Z_{i,o},\quad
z_j=\frac{R_j^{(5)}-\hat\mu_j^{\mathrm{OOS}}}{s_j},\quad
\hat Q_\tau(R_{i,o}^{(5)})=\hat\mu_{i,o}+s_{i,o}\,\hat Q_\tau(z_{\text{olgun}}).
$$

s bir koşullu ölçek tahminidir, doğrulanmadan tam hedef standart sapması diye etiketlenmez. Log Y üzerinde aynı yapı kurulursa Qτ(R)=exp(Qτ(Y))−1 dönüşümü ve R≥−1 desteği korunur; ortalama ayrı Ridge çıktısıdır. Basit-R sürümünde −1 altına inen quantile/ortalama için davranış önceden belirlenir ve dönüştürülmüş predictor ayrı değerlendirilir; sessiz clipping yoktur. Ampirik residual yaklaşımı normal kuyruk varsaymadan güçlü bir belirsizlik referansı sağlar. Üretilen quantile'in gerçekleşen kapsaması yine ölçülür.

GARCH(1,1), ayrı günlük seri r için:

$$
r_t=\mu+\epsilon_t,\quad\epsilon_t=\sigma_tz_t,\quad
\sigma_t^2=\omega+\alpha\epsilon_{t-1}^2+\beta\sigma_{t-1}^2.
$$

ω>0, α≥0, β≥0; standardized innovations ve olağan varsayımlar altında sonlu koşulsuz ikinci moment için α+β<1 gerekir. GJR asimetriyi ekler:

$$
\sigma_t^2=\omega+\alpha\epsilon_{t-1}^2+
\gamma\epsilon_{t-1}^2\mathbf1\{\epsilon_{t-1}<0\}+\beta\sigma_{t-1}^2.
$$

GJR'de pozitif varyans koşulları ile birlikte simetrik, birim varyanslı innovations için α+β+γ/2<1 ikinci moment koşuludur; asimetrik innovations altında γ/2 yerine ilgili negatif-kuyruk ikinci momenti gelir. Student-t variance standardizasyonu ν>2 gerektirir. [arch GARCH/GJR örnekleri](https://arch.readthedocs.io/en/latest/univariate/univariate_volatility_modeling.html).

**Günlük varyansın beş güne taşınması:** Sabit/önceden bilinen ortalamalı GARCH innovations martingale difference ise farklı günlerin ε değerleri koşullu olarak ilişkisizdir. Aynı close-to-close günlük süreçte:

$$
\operatorname{Var}_t\!\left(\sum_{k=1}^5\epsilon_{t+k}\right)
=\sum_{k=1}^5E_t[\sigma_{t+k}^2].
$$

Bu durumda ek kovaryans veya varyans için simülasyon gerekmez. GARCH günlük beklenti yinelemesi v1=σ²t+1 ve vh=ω+(α+β)vh−1'dir; simetrik GJR'de katsayıya γ/2 eklenir. Dinamik AR ortalaması gibi farklı süreçlerde kovaryans/dinamik yayılım ayrıca hesaplanır. Karekök beş ölçeklemesi, future varyanslar sabit değilken doğru değildir.

`arch.forecast(horizon=5)` çıktısındaki `h.5`, **beşinci gözlemin** varyansıdır; beş günlük toplam varyans değildir. İlgili martingale modelinde `h.1`…`h.5` toplanır. `variance` ile `residual_variance` dinamik ortalamada aynı olmayabilir. [arch forecast çıktı sözleşmesi](https://arch.readthedocs.io/en/latest/univariate/forecasting.html).

**Bu pilotun ufuk adaptörü:** Beş close-to-close varyansı, OPEN(d1)→CLOSE(d5) nakit-temettülü etiketinin varyansı değildir. İlk GJR rakibi toplam günlük varyansı yalnızca bir **ölçek girdisi** olarak kullanır; s(·) ve ampirik residual quantile'ları gerçek doğrudan hedefin geçmiş OOS hatalarında öğrenilir/kalibre edilir. Alternatif olarak gece/gün içi ve kurumsal eylem yollarını doğru eşleyen bir model gerekir; bu ayrı kapsam açılmadan varsayılmaz. Böylece variance formülünün doğruluğu ile yanlış etiket adaptasyonu karıştırılmaz.

Toplam log getirinin quantile'i günlük quantile'ların toplamı değildir. Doğru ufuk için dağılım şekli/quantile istenirse bağımlı yollar birlikte simüle edilir, önce yolun toplamı/serveti hesaplanır, sonra quantile alınır. Student-t, beş günlük toplamı otomatik Student-t yapmaz. Simülasyon, olmayan bir momenti var etmez.

**Gaussian-GARCH moment uyarısı:** Tek adımda sabit σ ile Gaussian log getiri lognormal ortalama verir; çok adımda gelecekteki varyans rastlantısaldır. Karışımın üstel momenti ayrıca kanıtlanmalıdır. Örneğin z1,z2 bağımsız N(0,1), σ1²=v ve σ2²=c+αvz1², c=ω+βv ise:

$$
E[e^{\epsilon_1+\epsilon_2}]
=e^{c/2}E\!\left[e^{\sqrt v z_1+(\alpha v/2)z_1^2}\right],
$$

sonlu olması αv<1 gerektirir. Bu, GARCH yinelemesinden yaptığımız matematiksel çıkarımdır; günlük Gaussian varsayımı ve sonlu ikinci momentin yeterli olmadığını gösterir. Çok adımlı log-GARCH Monte Carlo ortalamasını kanıtsız beklenen basit getiri diye adlandırmayız.

EGARCH yalnızca ertelenmiş katalog seçeneğidir:

$$
\log\sigma_t^2=\omega+\beta\log\sigma_{t-1}^2+
\alpha(|z_{t-1}|-E|z|)+\gamma z_{t-1}.
$$

GJR'nin γ katsayısıyla aynı yorumda değildir; çok adımlı analitik forecast her modelde bulunmaz. [EGARCH API](https://arch.readthedocs.io/en/latest/univariate/generated/arch.univariate.EGARCH.html). İlk pilotta EWMA referansı ve tek kayıtlı GJR ailesiyle başlanır; diğer GARCH türleri otomatik eklenmez.

**Başarısızlık:** Haber sıçraması, rejim değişimi, sınırda parametreler, az stres dönemi ve yanlış ufuk/seans eşlemesi. GARCH dalgalanması yön üstünlüğü kanıtlamaz. Volatilite rakibinin başarısı yön ve ortalama hipotezlerinden ayrı raporlanır.

## HMM ve Markov switching

Gizli Markov modelinde rejim S_t bir geçiş matrisiyle değişir:

$$
A_{jk}=\Pr(S_{t+1}=k\mid S_t=j),\qquad
r_t\mid S_t=j\sim N(\mu_j,\sigma_j^2).
$$

Rejimler düşük veya yüksek oynaklık gibi özelliklerle sonradan yorumlanır. İki Gaussian durum kendiliğinden güvenilir boğa ve ayı sınıfı değildir. Eğitim EM ile yapılabilir; başlangıç değerleri yerel optimumu etkileyebilir. [hmmlearn tutorial](https://hmmlearn.readthedocs.io/en/stable/tutorial.html).

Filtrelenmiş durum dağılımı π_t ise beş gün sonraki durum π_t A^5 ile tahmin edilir. Bu günlük emisyon modelinde beş close-to-close log getirinin beklenen toplamı:

$$
E[L_t^{(5)}\mid bilgi_t]=
\sum_{k=1}^{5}\pi_tA^k\mu.
$$

L bu pilotun OPEN→CLOSE basit servet etiketi değildir; ufuk adaptörü gerekir. HMM ikinci turda öncelikle filtrelenmiş durum özelliği olarak araştırılır. Son gündeki durum ortalamasını beşle çarpmak rejim geçişlerini ihmal eder. Alternatif olarak rejim olasılıkları supervised modelin girdisi olur veya ayrı uzmanların ağırlıkları olarak kullanılır.

Markov switching regression, katsayıların ve/veya varyansın rejime göre değiştiği modeli kurar. [statsmodels örneği](https://www.statsmodels.org/stable/examples/notebooks/generated/markov_regression.html).

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** Önce iki rejim, az sayıda feature ve diagonal covariance. Birkaç başlangıç ile eğitim; seçim yalnızca training/validation üzerinde. Üç rejimi ancak istikrarlı katkı görülürse denemek. Etiket değişimine karşı rejimleri geçmişte öğrenilen oynaklık sırasıyla eşleştirmek.

**Başarısızlık:** Üç yılda rejim örneğinin az olması, bir dönemi tamamen ezberlemek, durumun gerçekten değiştiğini gecikmeyle fark etmek ve rejimlere fazla hikâye yüklemek.

**Sızıntı:** π_t = Pr(S_t|y_1…y_t) canlı tahmine uygundur. Tüm seriyle hesaplanan Pr(S_t|y_1…y_T) smoothed olasılıkları geçmiş tahmin girdisi yapmayız. HMM kütüphanesinde verilen posterior fonksiyonunun geleceği kullanıp kullanmadığını kontrol etmek zorundayız.

## Cointegration ve Ornstein Uhlenbeck

Korelasyon, kalıcı denge ilişkisini kanıtlamaz. İki I(1) log fiyatın doğrusal birleşimi durağansa cointegration adayıdır:

$$
X_t=\log P_{A,t}-a-\beta\log P_{B,t}.
$$

Engle–Granger testi bu ilişkinin olup olmadığını değerlendirir; null hipotezi cointegration yoktur. a ve β yalnızca geçmiş eğitim penceresinde öğrenilir. [Cointegration testi](https://www.statsmodels.org/stable/generated/statsmodels.tsa.stattools.coint.html).

Durağan spread’e OU modeli:

$$
dX_t=\kappa(\theta-X_t)\,dt+\sigma\,dW_t,\qquad\kappa>0.
$$

$$
E[X_{t+h}\mid X_t]=\theta+(X_t-\theta)e^{-\kappa h},
$$
$$
Var(X_{t+h}\mid X_t)=
\frac{\sigma^2}{2\kappa}(1-e^{-2\kappa h}).
$$

Yarı ömür ln(2)/κ’dır. Günlük ayrık model X_t = a + b X_t−1 + ε_t şeklinde öğrenilebilir; OU yorumunda 0<b<1, κ=−ln b ve θ=a/(1−b) olur. b≈1 ise dengeye dönüş çok yavaştır ve parametre belirsizliği önemlidir.

Bu yapı PCA veya sektör ETF’leriyle açıklanan getirilerin residual süreçlerine de uygulanabilir. Avellaneda ve Lee bunun araştırma örneğini verir. [Yazarların makalesi](https://math.nyu.edu/~avellane/AvellanedaLeeStatArb071008.pdf).

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** Ekonomik ilişkisi anlamlı az sayıda çift; 126–252 günlük pencere; testleri sadece training’de yapmak; geçersiz veya istikrarsız çift için tahmin üretmemek. Çok sayıda çift tararsak çoklu test kontrolü gerekir. Pencereler bu projeye yönelik öneridir.

**Önemli hedef ayrımı:** Spread’in daralacağını tahmin etmek A fiyatının yükseleceğini söylemez. A ve B birlikte düşerken spread daralabilir. Üründe göreli tahmin olarak gösterilmeli veya piyasa/diğer varlık tahminiyle birleştiği varsayımlar açıkça belirtilmelidir.

**Başarısızlık:** Şirket iş modelinin değişmesi, ETF bileşimi değişmesi, birleşme, yapısal kırılma ve aynı testin çok sayıda çifte uygulanması. Daha önce dönmüş olması tekrar dönme garantisi değildir.

## VAR ve faktör modelleri

K küçük varlığın getiri vektörü için:

$$
r_t=c+\sum_{\ell=1}^{p}B_\ell r_{t-\ell}+\epsilon_t.
$$

K varlık ve p gecikmede yalnızca gecikmeli katsayı sayısı pK²’dir. K=20, p=5 için 2.000 katsayı oluşur; üç yıllık günlük geçmişte düzenlileştirmesiz bu yaklaşımı önermiyorum. K=3–5 ve p=1–2 dahi yalnızca daha sonraki kayıtlı deney adayıdır; ilk pilotta VAR yoktur.

Faktör yaklaşımı:

$$
r_{i,t}=\alpha_i+\beta_i^\top f_t+\epsilon_{i,t}.
$$

f; piyasa, sektör veya PCA bileşenleri olabilir. Bunların zaman bağımlılığı tahmin edilip varlık getirisine taşınabilir; residual ayrıca modellenebilir. PCA yalnızca eğitim bölümünde öğrenilmeli, gelecekteki tüm varlıkların korelasyon matrisi kullanılmamalıdır. PCA boyut azaltma sağlaması nedeniyle otomatik tahmin sinyali değildir.

**Başarısızlık:** Faktör ilişkilerinin değişmesi, yüksek boyut, gelecekteki faktörleri gerçek değerleriyle modele verme ve aynı piyasa bilgisini birçok ETF üzerinden tekrar sayma. Tahminin piyasa yönünden mi varlığa özgü residual’dan mı geldiğini raporlamak yararlıdır.

[statsmodels VAR](https://www.statsmodels.org/stable/vector_ar.html), [dinamik faktör modeli](https://www.statsmodels.org/stable/generated/statsmodels.tsa.statespace.dynamic_factor.DynamicFactor.html).

## Analog arama ve Bayesçi olay modelleri

Bu bölüm ertelenmiş Jev/olay araştırmasının katalog taslağıdır; pilot uygulama kapsamı değildir; aşağıdaki ağırlıklar doğrulanmış bir finansal algoritmanın hazır parametreleri değildir.

Mevcut olayın özellikleri x_t, metin özellikleri z_t olsun. Geçmiş aday j için sayısal uzaklık d(x_t,x_j), Jev ilişki skoru s_j ve olay yaşı hesaplanır. Sonucu olgunlaşmış adayların ağırlığı:

$$
\tilde w_j=
\exp[-d(x_t,x_j)^2/(2b^2)]\,g(s_j)\,
\exp[-\lambda\,yaş_j],\qquad
w_j=\tilde w_j/\sum_k\tilde w_k.
$$

Benzer olayların getiri dağılımı:

$$
\hat F(y\mid x_t,z_t)=
\sum_jw_j\mathbf1\{R_j^{(5)}\leq y\}.
$$

Adayları, gerçekleşen getirilerini görmeden önce mevcut özellikler ve konu benzerliğiyle seçeriz. Kazandıran geçmiş örnekleri seçip bugüne eşlemek geçerli tahmin yöntemi değildir. Jev dış arama veya piyasa verisi sağlayıcısı değildir; önce kod ve arama dizini adayları üretir, Jev sıralar.

Nadir olaylarda dağılımı piyasa/ sektör referansına doğru küçülten bir karışım:

$$
F_{\text{shrink}}=\rho\hat F+(1-\rho)F_0.
$$

ρ validation’da belirlenebilir; ağırlıkların yoğunluğu için n_eff=1/Σw_j² kullanılabilir. Ancak aynı makro olay veya yakın günlerden gelen ilişkili örnekler bu sayının gerçek bağımsız kanıt miktarını abartmasına yol açar. Olay kümeleri deduplicate edilmelidir. Jev skorunu yüzde 80 piyasa yükseliş olasılığı gibi yorumlayamayız.

Basit bağımsız binary olay etiketlerinde Beta prior ve Binomial gözlemlerle:

$$
p\sim Beta(a,b),\quad
p\mid D\sim Beta(a+n_\uparrow,b+n_\downarrow).
$$

Bu form ağırlıklı, ilişkili ve seçilerek getirilmiş haberler için aynen geçerli exact posterior değildir. Gerçek uygulamada seçim süreci, bağımlılık ve küçük örnek belirsizliği ayrıca modellenmelidir.

Daha güçlü seçenek hiyerarşik regresyondur:

$$
R_{i,o}^{(5)}=
\alpha_i+\beta^\top x_{i,t}
+\theta_{sektör(i)}^\top z_{i,t}+\epsilon_{i,t},
$$
$$
\theta_s\sim N(\theta_0,\tau^2I).
$$

Sektör etkileri ortak ortalamaya doğru küçülür; az örnekli sektörlerde uç katsayılar sınırlandırılabilir. Bu yapı benzer haberlerin şirketlere göre farklı etkisini öğrenmeye adaydır. Haber etkisinin nedensel olduğu sonucu yalnızca bu regresyondan çıkmaz.

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** Çok dar 5–10 haber özelliği; tüm geçmişi büyük prompt’a vermeden 20–50 aday; en çok 10–20 analogun ayrı açıklanması. Haber sürümü, yayın anı, kaynak ve model sürümü kaydı. Nadir olaylarda kesin sayı yerine yetersiz örnek bildirmek.

**Başarısızlık:** Metinde benzerliğin ekonomik benzerlik sanılması, fiyatlanmış haber, tekrarlanan haber, yanlış varlık eşleşmesi, yayın zamanı hatası ve çağdaş modellerin geçmiş sonucu önceden biliyor olması. Özellik tasarımını dış test dönemindeki hatalara bakıp tekrar tekrar değiştirmek de test sızıntısıdır.

TypeSafe’in özellik keşfi örneği Jev skorlarını downstream modelde kullanmayı gösterir; örnek wine review verisidir ve finansal performans kanıtı değildir. [Özellik keşfi örneği](https://docs.typesafe.ai/cookbooks/autoresearch_feature_discovery), [model sınırları](https://docs.typesafe.ai/models).

## Kernel modeller ve Gaussian process

Az sayıda özelliğin doğrusal olmayan ilişkisini kontrol etmek için RBF kernel:

$$
K(x,x')=\exp[-\gamma\|x-x'\|^2].
$$

Kernel Ridge, SVR veya Gaussian process bu benzerlik üzerinden tahmin üretebilir. Gaussian process’de latent tahmin ortalaması ve varyansı:

$$
\mu_*=k_*^\top(K+\sigma_n^2I)^{-1}y,
$$
$$
v_*=k_{**}-k_*^\top(K+\sigma_n^2I)^{-1}k_*.
$$

Latent belirsizlik ile gözlem/getiri belirsizliği ayrıdır; yeni getirinin varyansına gözlem gürültüsü de eklenir. Kernel doğru ilişkiyi tarif etmiyorsa varyans piyasa riskini güvenilir biçimde temsil etmez.

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** 5–10 standardize özellikte küçük bir kontrol modeli. Exact Gaussian process için yoğun matris çözümü tipik olarak O(n³) zaman ve O(n²) bellek gerektirir; büyük panelde ilk ana model olmamasının bir nedeni budur. Ölçekler ve kernel parametreleri yalnızca training’de öğrenilmelidir.

**Başarısızlık:** Feature boyutu büyüdükçe benzerliğin zayıflaması, rejim değişimi, örnek dışına güvenilir extrapolation yapamama ve iid hata varsayımının bozulması. [Gaussian process resmi açıklaması](https://scikit-learn.org/stable/modules/gaussian_process.html).

## Modern zaman serisi karşılaştırmaları

### DLinear ve NLinear

Bu alt bölümde Y genel neural forecast çıktısıdır; ilk pilotun etiketini veya kapsamını değiştirmez. DLinear geçmiş diziyi trend ve kalan bileşene ayırıp doğrusal dönüşümlerle geleceğe taşır:

$$
X=T+S,\qquad
\hat Y=W_TT+W_SS.
$$

NLinear son gözleme göre normalize edilmiş diziyi doğrusal modelleyip referans seviyeyi geri ekler. Sabit doğrusal moving-average decomposition ile DLinear girdi penceresinin doğrusal bir haritasıdır; NLinear son-değer çıkarma/ekleme de doğrusal haritada temsil edilebilir. Bu nedenle yeterli lag girdili Ridge ile işlev sınıfı örtüşür; ayrı model adı otomatik yeni nonlinear kapasite sağlamaz. Düşük boyutlu ilk 15-girdi Ridge modeliyle mimari olarak aynı olduğunu iddia etmiyoruz. Bu modeller ancak yeni turda farklı düzenlileştirme/paylaşım iddiasını sınamak için açılır. Yazarların benchmark sonucu kendi veri ve ufuklarına aittir; beş günlük finansal getiriye üstünlük kanıtı değildir. [DLinear makalesi](https://arxiv.org/abs/2205.13504), [yazarların kodu](https://github.com/cure-lab/LTSF-Linear).

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** 20 ve 60 günlük context, beş günlük doğrudan çıktı; aynı hedef, veri kesimi ve test döneminde ElasticNet ve LightGBM ile karşılaştırma. Fiyat seviyesinde sonuç raporlanırsa son fiyat referansı zorunludur. Decomposition işlemi tahmin anından sonraki gözlemleri kullanmamalıdır.

### N BEATS ve N HiTS

N-BEATS residual bloklarıyla geçmişi açıklayan backcast ve geleceği veren forecast üretir. N-HiTS çok ölçekli örnekleme ve interpolation ile farklı frekanslarda tahmin bileşenleri kurar:

$$
\hat Y=\sum_b\hat Y_b,\qquad
\hat Y_b=\operatorname{Interpolate}(\theta_b).
$$

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** Küçük ortak model, 60 günlük context, sınırlı blok sayısı ve parametre bütçesi. Üç yıllık tek seriden büyük ağ eğitmek yerine çok serili panel üzerinde araştırmak. Beş günlük kısa ufukta çok ölçekli yapının ekstra katkısını varsaymamak.

Resmi NeuralForecast deposu bu modellerin uygulamalarını sunar. [N-HiTS makalesi](https://arxiv.org/abs/2201.12886), [NeuralForecast](https://github.com/Nixtla/neuralforecast).

### LSTM ve Transformer

LSTM hücresi geçmiş bilgiyi kapılarla günceller. Transformer attention ilişkisi:

$$
Attention(Q,K,V)=softmax(QK^\top/\sqrt d)V.
$$

Bu mekanizmalar sıralı ilişki öğrenebilir; matematiksel karmaşıklık piyasa öngörülebilirliği üretmez. Geçerli eğitimde attention maskesi, sequence sınırları, ölçekleme, future covariates ve tarih bölmeleri kontrol edilmelidir.

**Ne zaman:** Daha uzun tarih, geniş ama ilişkili olduğu hesaba katılmış panel, birden çok ufuk ve basit modellerin açıklayamadığı istikrarlı yapı varsa. İlk üç yıllık tek varlık deneyi için önceliği düşük. Temporal Fusion Transformer geçmiş ve gelecekte bilinen covariates ayrımını ve çok ufuklu quantile tahminini destekleyen bir araştırma örneğidir. [TFT makalesi](https://arxiv.org/abs/1912.09363).

### Chronos ve TimesFM

Amazon Chronos ve Google TimesFM önceden eğitilmiş zaman serisi modelleridir; eğitmeden veya sınırlı uyarlamayla ayrı referans tahminleri üretilebilir. [Resmi Chronos deposu](https://github.com/amazon-science/chronos-forecasting), [resmi TimesFM deposu](https://github.com/google-research/timesfm).

**Araştırma taslağı (yalnızca kayıtlı kapsam açılırsa):** Sabit checkpoint ile zero-shot karşılaştırma. Aynı target ve holdout döneminde test; fiyat ve getiri hedeflerinin ayrı raporu; model context’inin yalnızca geçmişten oluşturulması. Depodaki en yeni sürümü isimsiz biçimde kullanmak yerine checkpoint ve ağırlık hash’ini kaydetmek.

**Belirsizlik:** Genel zaman serisi başarısı, ABD ETF kısa vadeli getiride üstünlük demek değildir. Ön eğitim veri kapsamı ve zaman sınırı yeterince bilinemiyorsa geçmiş testte kontaminasyon ihtimali kalır. Bu risk ileriye dönük tahmin kayıtlarını gerekli kılar.

## Reinforcement learning neden ayrı bir faz

RL temelinde politika π ve ödül vardır:

$$
Q^*(s,a)=E[r+\gamma\max_{a'}Q^*(s',a')\mid s,a].
$$

State, action, transaction cost ve portföy değeri gibi bileşenler gerekir. Bir sonraki getirinin olasılığını öğrenmek supervised forecasting problemidir; işlem eylemini seçmek farklı hedeftir. İlk faz emir ve portföy yönetmediği için RL geliştirmesi ürün hedefini gereksiz genişletir.

İleride alım satım fazı açılırsa fiyat tahmininden ayrı politika katmanı olarak ele alınabilir. Simülatörün geçmiş veriye aşırı uyumu ve gerçek yürütme farkları ayrıca doğrulanır. [FinRL resmi deposu](https://github.com/AI4Finance-Foundation/FinRL).

## Ensemble, kalibrasyon ve tutarlı çıktılar

Ensemble otomatik ilk tur katmanı değildir. Yeni kayıtlı yöntem olarak açılırsa ağırlıklar yalnızca geçmiş OOS tahminlerde seçilir:

$$
\hat\mu_o=\sum_mw_m\hat\mu_{m,o},\quad w_m\geq0,\quad\sum_mw_m=1.
$$

**Basit-getiri dağılım karışımı** F_R=ΣwmFm için, her bileşenin gerekli momenti varsa:

$$
\mu=\sum_mw_m\mu_m,\qquad
\operatorname{Var}(R)=\sum_mw_m(\sigma_m^2+\mu_m^2)-\mu^2.
$$

Bu modelde bileşen rastgele seçilir; modeller arası location farkı da dağılım varyansına girer. Karışım quantile'i inf{y:ΣwmFm(y)≥τ} ile karışım CDF'inden bulunur. ΣwmQτ(Fm) farklı bir quantile birleşimidir; genellikle karışım quantile'i değildir.

**Ağırlıklı predictor/hata toplamı** e=Σwmem için farklı formül gerekir:

$$
\operatorname{Var}(e)=w^\top\Sigma_e w
=\sum_mw_m^2\operatorname{Var}(e_m)
+2\sum_{m<n}w_mw_n\operatorname{Cov}(e_m,e_n).
$$

Burada ortak hedefteki tahmin hatalarının kovaryansı önemlidir; karışım formülü yerine kullanılamaz. Eğitim fitted prediction'larını meta modele vermek de geçerli OOS ensemble kanıtı değildir.

Lojistik yönün sigmoid kalibrasyonu:

$$
p_{\mathrm{cal}}=\sigma(a\,score+b).
$$

Kalibratör bağımsız kronolojik calibration bölümünün gerçek OOS ve olgunlaşmış tahminlerinde öğrenilir. İlk registry'de sigmoid'in iki katsayısı 2023 bölümünde fit edilir, holdout/ilk forward protokolünde sabittir; base-model aylık refit'i ve scale residual pool güncellemesi ayrıdır. İleride delayed-feedback calibration başka kayıtlı politika olabilir. Küçük, ilişkili örneklemde isotonic otomatik eklenmez. Brier düşmesi kalibrasyonun tek başına iyileştiğini kanıtlamaz; intercept/slope ve reliability belirsizliği ayrıca raporlanır. [Resmî olasılık kalibrasyonu belgeleri](https://scikit-learn.org/stable/modules/calibration.html).

### Signed CQR: yöntem doğru, zaman serisi garantisi sınırlı

Conformalized quantile regression açılırsa standard score, sıfıra kırpılmayan işaretli skordur:

$$
e_j=\max\{\hat Q_{lo,j}-R_j,\ R_j-\hat Q_{hi,j}\}.
$$

Olgunlaşmış n calibration skoru sıralanır; α=0.20 için k=⌈(n+1)(1−α)⌉, c=e(k) ve interval [Qlo−c,Qhi+c]'dir. k>n ise sonlu quantile yoktur; teorik konvansiyon c=+∞'dir, yetersiz calibration ile sonlu “garantili” interval sunulmaz. Genel interpolated yüzde 80 `quantile()` bu finite-sample rank kuralının yerine otomatik geçmez. c negatif olabilir ve aralığı daraltır; e=max(…,0) ancak ayrı **yalnızca genişletme** politikasıdır, standard signed CQR değildir. Ters/boş sınırlar ve eksik calibration açık hata durumudur. [Romano, Patterson ve Candès'in özgün CQR makalesi](https://arxiv.org/abs/1905.03222).

CQR'nin finite-sample garantisi exchangeability altında marjinaldir; örtüşen finansal etiketlerde kronolojik bölme, otokorelasyon ve rejim kayması bu varsayımı sağlamaz. Bu yüzden “her varlık ve rejimde kesin %80” iddiası yoktur. Calibration yalnızca o ana kadar olgunlaşan hatalarla güncellenir. Adaptif yöntemler ayrı politika olarak kaydedilir; uzun dönem coverage sonucu koşullu coverage demek değildir. [Adaptive Conformal Inference](https://arxiv.org/abs/2106.00170).

CQR sonrası interval uçları **kapsama amaçlı sınırlar**dır; otomatik ayrı Q0.10/Q0.90 koşullu quantile'lar veya tam CDF değildir. Ham quantile ile yayımlanan düzeltilmiş sınır ayrı alanlarda saklanır. Kayıt hangi uçları değerlendireceğini önceden belirler; nominal uçlarda pinball kaybı ölçmek yine mümkündür, fakat kalibrasyon garantisiyle karıştırılmaz.

### Yön, quantile ve ortalama aynı dağılımı tarif ediyor mu?

Tek F_R yayımlanıyorsa p=1−F_R(0), Qτ=F_R^{-1}(τ), ortalama ise varsa ∫r dF_R(r) olmalıdır. Quantile'lar artmalı, R desteği −1 altına inmemelidir. Dağılım karışımından quantile almakla component quantile'larını ortalamak aynı işlem değildir.

Ridge ortalama, lojistik yön ve aralık modeli ayrı functional tahminlerdir; ilk pilot bunları ayrı etiketleyip ayrı ölçer. Üç nominal quantile tam CDF tanımlamaz. Sonuçları tek tutarlı dağılım olarak birleştirmek ayrıca kayıtlı kısıtlı birleşim gerektirir. Örneğin Q0.90<0 ile p=0.80 aynı dağılım için bağdaşmaz: Q0.90<0 ise p≤0.10. Genel olarak Qτ<0 ise p≤1−τ; Qτ>0 ise p≥1−τ. Çelişki sayısı kayda/tanısal rapora girer; kullanıcıya ikisi tek “model güveni” diye sunulmaz. Ortalamanın işareti medyan/yönün işaretine eşit olmak zorunda değildir.

Jev/LLM confidence hiçbir durumda finansal probability, coverage veya epistemik belirsizlik yerine kullanılmaz.

## Test tasarımı ve veri miktarı

Üç yıl yaklaşık 750 günlük tek seri, kabaca 150 örtüşmeyen beş-seanslı dönemdir. Bu sayı evrensel etkili örnek sayısı değildir; farklı rejimler ve aynı piyasa şoku bağımsız değildir. Günlük örtüşen etiketler ve çok sayıda benzer ETF daha çok satır verir, aynı oranda bağımsız olay sağlamaz. Seçilmiş bugünkü kohortun geçmiş sonucu bütün ABD piyasasına genellenmez.

Bir seri için bağımsız dönem varsayımıyla korelasyon gücü örneği:

$$
n\simeq\left[\frac{z_{1-\alpha/2}+z_{1-\beta}}{\operatorname{atanh}(\rho)}\right]^2+3.
$$

%5 iki taraflı test ve %80 power altında 500 bağımsız dönemde yaklaşık 0.125, 26 dönemde yaklaşık 0.53–0.55 korelasyon algılanabilir. Bu **basitleştirilmiş örnek**, pinball/Brier için güç hesabı değildir; farklı bağımlılık ve çoklu-test düzeltmesi gereksinimi değiştirir. Altı aylık ileri kayıt küçük edge'i ispatlama süresi sayılmaz. Gerçek karar, geliştirme dönemindeki eşlenik kayıp farklarının tarih-blok varyansı ile hedef etkide power/MDE ölçülerek verilir; güç düşükse sonuç “belirsiz” kalır, satır sayısı artırılıp başarı ilan edilmez.

İşletilecek tek protokol [deney-plani.md](deney-plani.md)'de dondurulur:

1. Bölme forecast-origin tarihiyle yapılır; aynı tarihte bütün varlıklar birlikte kalır. Training → validation → calibration → final holdout zaman sırasıdır. Forward/kanıtlı-vintage fit'te etiketi olgunlaşmamış örnek purge edilir; exploratory backfill replay'de actual capture/matured_at korunup ayrı assumed eligibility kullanılır. Bu varsayım eski sürüm kanıtı veya confirmatory geçiş değildir. Sınırı aşan etiket gerçek başlangıç/bitiş zamanından purge edilir; iki modun sözleşmesi tek deney kaydındadır.
2. Panel için `TimeSeriesSplit(gap=5)` yeterli değildir: gap satır sayar. Özel tarih/seans splitter'ı, beş-seans overlap ve veri availability zamanını denetler. [API](https://scikit-learn.org/stable/modules/generated/sklearn.model_selection.TimeSeriesSplit.html).
3. Ölçekleme, imputation, winsorization, özellik/çift/soru seçimi, calibration, residual havuzu ve varsa PCA yalnızca izin verilen geçmişte öğrenilir. Satır grupları tarih ve varlık kimliğiyle testlenir.
4. Aynı origin/varlıklarda eşlenik loss farkı hesaplanır; önceden tanımlı varlık ağırlıklarıyla önce güne indirgenir. Tarih-blok bootstrap / kayıtlı HAC yöntemi cross-asset dependence ve beş günlük overlap'i korur. En az dört gecikme günlük overlap için başlangıçtır; daha uzun bağımlılık olabilir, blok/gecikme duyarlılığı önceden kaydedilir.
5. Pencere, parametre, özellik, kalibrasyon ve model seçimleri dahil en çok 15 aday log'lanır. Birincil karşılaştırma seçimi final testi görmeden dondurulur. Aday/sonlanım ailesinde kayıtlı Holm veya Romano–Wolf düzeltmesi/simultaneous intervals uygulanır; düzeltmesiz tek iyi p-değeri kabul gerekçesi olmaz.
6. Final holdout tek kayıtlı unseal ile açılır. Açılmış tarihleri etkileyen tasarım değişikliği yeni test verisi ister; eski holdout artık geliştirme verisidir. Önceden kayıtlı refit/gecikmiş-feedback kuralları ile araştırmacının sonucu görüp değiştirmesi ayrıdır.
7. Bütün scheduled origin'ler (başarı, gecikme, failure, fallback, abstention) append-only/hash-chain tahmin defterindedir. Sonuç yeni sürümlü kayıttır; tahmin yeniden yazılmaz. Çekimserlik score'u kolay örnek seçerek düşürebilir: kapsam eşlenmiş kontrol zorunludur.

Çoklu deneme yalnızca kronolojik bölmeyle çözülmez; seçilmiş sonucun rastlantı riski sınırlı arama, correction, korumalı holdout ve yeni ileri kanıtla yönetilir. [Bailey ve arkadaşlarının araştırması](https://www.davidhbailey.com/dhbpapers/overfitting.pdf).

## Ölçümler ve kabul yaklaşımı

Aşağıdaki kayıp–functional eşlemesi zorunludur; model ailesinin adına bakıp medyanı ortalama diye adlandırmayız. [scikit-learn resmî scoring rehberi](https://scikit-learn.org/stable/modules/model_evaluation.html#which-scoring-function-should-i-use) bu ayrımı açıklar.

| Tahmin edilen functional | Model/seçim kaybı | Ek tanılar |
| --- | --- | --- |
| E[R\|x] (Ridge) | MSE, RMSE; geçmiş ortalama referansına karşı OOS R² | MAE yalnızca ikincil robust tanı |
| Q0.50(R) | MAE veya τ=0.5 pinball | Medyan bias |
| Q0.10/Q0.90(R) | İki quantile pinball kaybının ortalaması | Coverage, width, interval score, varlık/dönem tanıları |
| Pr(R>0\|x) | Brier (ikincil hipotez) | Log loss, Brier skill, calibration intercept/slope/curve |
| Tanımlı residual varyansı | QLIKE gibi kayıtlı kayıp | Proxy hassasiyeti, rejim bazlı hata |

Birincil origin kaybı:

$$
\ell_{i,o}^{\mathrm{PB}}=\tfrac12\left[
\rho_{0.10}(R_{i,o}^{(5)}-\hat Q_{0.10,i,o})+
\rho_{0.90}(R_{i,o}^{(5)}-\hat Q_{0.90,i,o})\right].
$$

**Pilotun planlama hedefi**, güçlü kayıtlı EWMA-ampirik referansına göre bu kayıpta **en az %2 nokta tahmini göreli iyileşme**dir. Bu eşik doğrulanmış kullanıcı değeri veya garantili sonuç değildir; final holdout açılmadan gerçek power/MDE ve kullanım kararıyla dondurulur. CI altı sıfırdan büyük olması en az %2 faydayı kanıtlamaz; bütün istatistik/pratik/coverage koşulları tek deney kaydında tanımlanır. Aralığın nominal kapsaması %80'dir; kaybı genişlik ve miss cezasıyla birlikte ölçeriz:

$$
IS_\alpha(L,U;R)=(U-L)+\frac{2}{\alpha}(L-R)\mathbf1\{R<L\}
+\frac{2}{\alpha}(R-U)\mathbf1\{R>U\},\quad\alpha=0.20.
$$

Geniş aralıkla coverage artırmak tek başına yararlı tahmin değildir. Coverage'in güven aralığı ve önceden kayıtlı toleransı; örtüşmeyen alt örneklemler ve tüm tarih-blok sonuçları birlikte raporlanır. Küçük varlık/regime alt gruplarında sabit başarı yüzdesi evrensel sınır sayılmaz.

Yön ölçümleri:

$$
Brier=\frac1n\sum_o(p_o-D_o)^2,\qquad
LogLoss=-\frac1n\sum_o[D_o\log p_o+(1-D_o)\log(1-p_o)].
$$

Brier skill, aynı origin'deki yalnızca geçmiş olgunlaşmış etiketlere dayalı yükseliş sıklığına karşı ölçülür. Sıfır-return referansı yerine otomatik p=0.50 koyulmaz. **Yön ve ortalama ayrı ikincil hipotezlerdir:** kendi anlamlı etki/power ve düzeltilmiş testleri gerekir; interval başarısı bunlara geçmez. Yön için önerilen %1 Brier planning eşiğinin bütün koşulları tek deney kaydındadır; bu evrensel veya ölçülmüş kullanıcı değeri değildir. Yeterli gücü olmayan sonuç, olumsuz sonuçla aynı değildir; doğruluk %60 gibi evrensel başarı eşiği yoktur.

QLIKE=log(v̂)+v/v̂ için v açıkça tanımlı realized variance proxy'sidir. Tek beş günlük kare log toplam veya günlük kare getiriler gürültülü proxy'dir; günlük varyansın toplamıyla hedef varyansı özdeş değildir. Intraday proxy ayrı veri ihtiyacıdır. [Patton'ın volatility proxy araştırması](https://public.econ.duke.edu/~ap172/Patton_vol_proxies_JoE_2011.pdf). Göreli IC ilk pilot sonlanımı değildir; ayrı hedef açılırsa gün bazında rank IC, tarihe göre belirsizlik ve çoklu-test kaydı gerekir.

## İlk deney paketi: tek registry, sınırlı kapsam

İlk sürümde zorunlu referanslar, Ridge ve lojistik; koşullar geçerse EWMA/GJR ölçek karşılaştırması vardır. Yalnızca [deney-plani.md](deney-plani.md)'de açık konfigürasyonlar çalıştırılır. Bu belge artık ikinci 20–30-model deneyi tanımlamaz. Baselines dahil en çok 15 toplam kayıt, en çok 15 girdi, bir beş-seans horizon, aynı nedensel origin ve aynı label sözleşmesi geçerlidir. Bütçe dolarsa yeni aile eklemek için başarısızları sessiz silmek yoktur; yeni tur gerekçe ve yeni test verisi ister.

Geçiş koşulları sıralıdır: veri/hak/availability sözleşmesi → referansların doğruluğu ve gerçek power → sınırlı sayısal adaylar → korumalı final test ve ileri kayıt. Sonuç anlamlı değilse kapsamı büyütmek yerine hipotez reddi/belirsizlik raporlanır. Sadece interval yararı görülürse ürün iddiası da uncertainty yönünde daralır. Kalman/HMM ikinci turdan önce açılmaz; metin, analog, spread, derin model, RL ve ürün UI kendi kapıları geçmeden uygulama işi açılmaz.

## Jev ve LLM araştırmasının ertelenen sınırları

İlk pilotta haber/Jev/LLM forecasting girdisi değildir; numeric static rapor ve şablon açıklama yeterlidir. Daha sonraki metin deneyi, az sayıda hisseye on yıllık Jev backfill yaparak üstünlük kanıtladığını iddia edemez. Geniş, önceden dondurulmuş olay evreni; gerçek metin sürümü/available_at; haklar; model training cutoff; sabit soru/model sürümü ve yeni ileri kanıt gerekir. Cutoff bilinmiyorsa geriye dönük Jev sonucu doğrulayıcı değil keşifseldir. Şirket adı anonimleştirme duyarlılık kontrolüdür, kontaminasyonu temizlediğinin kanıtı değildir. [Glasserman ve Lin araştırması](https://arxiv.org/abs/2309.17322).

Ablation zinciri ayrı kaydedilir: no-text → haber sayısı/zaman/takvim → ucuz dictionary/TF-IDF → sabit Jev soruları. Analog önce numeric/BM25 shortlist'e karşı sınanır; Jev rerank katkısı aynı aday havuzunda ölçülür. Olgunlaşmamış sonuç, duplicate olay ve aynı haber kopyaları efektif kanıt sayısını artırmaz. Olay etiketi sınıflandırma başarısı downstream finansal değer demek değildir.

Jev request/response, tam input/output, hash, sorular, returned model ID, version, prompt, source sürümü, timestamp, error, retry ve fallback saklanır. Model alias'ı tek başına replay sağlamaz; API değişkenliğinin finansal etkisi ayrıca ölçülür. Jev skorunu kalibre Pr(R>0) gibi kullanmayız.

LLM yalnızca daha sonraki açıklama kapısında kaynak paketiyle açılır. Sayısal tahmin önce kaydedilir; LLM aynı sayıları koruyorsa Brier/RMSE/pinball değişemez. Karşılaştırma şablon açıklamaya karşı sayı tutarlılığı, source attribution, factual support ve kullanıcı yararıdır. LLM kaynak hatasında şablon fallback uygulanır. Tahmini değiştirme, model seçme veya “yetersiz kanıt” diyerek abstention ayrı forecasting politikasıdır; kapsam eşlenmiş volatilite/ucuz çekimserlik kontrolüne karşı sınanır.

Katalog seçenekleri tahmin üstünlüğü vaadi değildir. Bu sürümün işi, aynı bilgi sınırı ve gerçek OPEN→CLOSE servet etiketi altında küçük hipotezleri reddedilebilir biçimde sınamaktır.
