# Mac gözlem uygulaması

`npm ci --ignore-scripts && npm run dev` ile `http://127.0.0.1:5173` açılır. Node 22.12+ gerekir. HTML/CSS/vanilla TypeScript; `npm run build` tip ve üretim derlemesini kontrol eder.

`observe.mjs`, yalnız bu uygulama için `127.0.0.1:18883 → root@167.235.158.118:127.0.0.1:18880` SSH tüneli kurar. Vite'ın sunucu tarafındaki proxy'si, `~/.config/trading/operator.env` içindeki `OPERATOR_TOKEN` değerini inject eder. Bu dosya repo dışındadır ve 0600 olmalıdır. Token browser bundle veya URL'ye girmez. Proxy yalnız GET `/api/observation` kabul eder. Tünel 10 saniye arayla yeniden bağlanır, uygulama kapanınca kendi SSH sürecini kapatır. Port çakışmasında mevcut servisi öldürmez.

15 saniyelik okuma üst üste bindirilmez. Bağlantı yoksa son kayıt ve zaman kalır. Eksik fiyat toplamı `—` yapar; eski/doğrulanmamış pozisyon fiyatları etiketlenir. Para backendde decimal string hesaplanır, ekranda BigInt ile cent formatlanır. Arayüz GPT çağrısı veya emir üretmez. `vite preview` statik çıktı kontrolüdür; güvenli SSH proxy yalnız `npm run dev` akışında çalışır.
