# Mac gözlem ekranı

Referans: `observer-concept.png`, ImageGen built-in ile oluşturuldu. Kullanıcı basit, salt okunur HTML/CSS/TypeScript ekran istedi; React eklenmedi.

Tam beyaz zemin, navy #0c173d metin, #536487 yardımcı metin, #0764e9 seçim, #d8e0ec çizgi. Sistem sans fontu, 40/26px başlık, 14–16px tablo ve kontroller, tabular para. 40px dış boşluk, 22px kolon aralığı, 6px çerçeve. Ana ekran: yatay Trading / Cüzdanlar / Emirler / Sistem navigasyonu, üçlü toplam bandı, sekiz satır tablo, seçili hesap detayı, üç detay sekmesi, gerçekleşme/güncellik alt bandı.

İlk tasarımdaki 800.000 ve 100.000 değerleri başlangıç sözleşmesine aittir. Üretim arayüzü yalnız backend decimal string değerlerini gösterir. Eksik fiyat sıfır sayılmaz; doğrulanmamış/eski değerler açık etiketlidir. Kopan bağlantıda son başarılı veri ve zamanı kalır. Hesap sayısı backendden gelir. Detayda karar ve gerçekleşme metinleri kaçışlanır.

Mobilde kolonlar alt alta gelir, tablo yatay kayar. Bağlantı ve hata durumları, uzun gerçek karar metinleri, sistem kayıtları aynı görsel sistemin gerekli işlevsel uzantılarıdır. Ekranda emir verme veya GPT başlatma kontrolü yoktur.

Prompt: Türkçe sekiz bağımsız sanal hesap için 1440x1000 bütün gözlem ekranı; Trading, Cüzdanlar, Emirler, Sistem; nakit/hesap/pozisyon toplamları; seçili US-FAST; tablo ve detay; gerçekleşme kapalı ve kaynak kabulü bekleniyor durumu. Görsel yalnız referanstır, ekran bitmap olarak kullanılmaz.

## Referans karşılaştırması

- Kod ile çizilmiş ekran, bitmap kullanmıyor. IAB'de gerçek backend ile 1440x1000 ve 390x844 sınandı; referans ve son masaüstü/mobil screenshotları `view_image` ile birlikte incelendi.
- Kopya/navigasyon: Trading, Cüzdanlar, Emirler, Sistem, Yenile, tablo ve üç sekme korundu. Gerçek as-of zamanı ve GMT+3, tasarımdaki `—` yerine backendden gelir.
- Kompozisyon: açık üçlü toplam bandı, sekiz satırlı tablo, seçili hesap detayı, kaynak kabulü alt bandı korundu. İlk alfabetik sıralama sabit cüzdan sırasına düzeltildi.
- Tipografi: sistem sans, navy başlık/para, tabular sayı, kontrollü 40/26/14–16px hiyerarşi. İlk dikey boşluklar referansa yaklaşmak için daraltıldı.
- Palet/çerçeve: tam beyaz zemin, mavi seçili satır, ince gri çizgi; gölge/illüstrasyon/pazarlama bölümü eklenmedi.
- İşlevsel uzantı: gerçekleşmiş/gerçekleşmemiş değişim ve temettü alacağı için kapalı 'Diğer hesap bilgileri' kontrolü eklendi; bu bilgiler ürün sözleşmesinin gereğidir. Bağlantı hatası gerçek son kayıtla görünür.
- Mobil: tablo kendi çerçevesinde yatay kayar, detay aşağı taşınır; dış belge genişliği 390px içinde kalır.

Hesap seçimi → QQQ-FAST → Kararlar akışında gerçek HOLD kararı; Sistem görünümünde gerçek işler/makbuz sayısı; tünel kesilmesi ve geri gelmesi kontrol edildi. Başarılı bağlantıda konsol uyarısı/hatası yoktu. Ekran görsel yapısı referansla doğrulandı; canlı fiyat/BBO kabulü bu UI kabulünden ayrı açık kalır. Geçici QA görüntüleri repo dışında tutuldu.
