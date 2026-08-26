# BOXEL

Aplikacja desktopowa do **budowania i grania** w kwadratowe światy z klockami (styl OG) i programowaniem jak w Scratchu. Jest multiplayer, czat i znajomi.

## Uruchomienie (aplikacja, nie strona)

```bash
cd boxel
npm install
npm run app
```

To odpala:

1. serwer pokoi (`ws://127.0.0.1:8787`)
2. silnik UI
3. **okno Electron** — BOXEL → Nowe okno (drugi gracz) otwiera drugie konto na tym samym PC

Bez Electrona (tylko do testów silnika): `npm run server` i `npm run dev`.

## Multiplayer, znajomi, czat

- Każdy ma **kod znajomego** `BX-XXXXXX`. Wpisz kod kolegi i kliknij Dodaj.
- **Załóż pokój** — dostajesz kod 4-cyfrowy. Znajomy klika Zaproś albo wpisuje kod.
- W pokoju widać innych graczy (kolorowe kostki), wspólne stawianie klocków i czat.
- Czat pokoju + szept tylko do znajomych. Linki, Discord i wulgaryzmy są blokowane. Są gotowe bąbelki: Cześć!, Chodź tu!, Pomocy!…

## Język i prędkość

Nie przerzucamy całego studia na Pythona ani C++.

| Warstwa | Język | Dlaczego |
| --- | --- | --- |
| Studio + klocki + voxel 24³ | TypeScript | To nie jest wąskie gardło. Test `physics speed`: 2000 klatek fizyki w << 250 ms. |
| Okno aplikacji | Electron (Chromium) | To jest **aplikacja**, nie karta przeglądarki. Godot/C# ma sens przy większych światach i 50+ graczach — wtedy się przenosi silnik, nie teraz. |
| Pokoje / czat / sync | Node WebSocket | 8 graczy w pokoju, 15 Hz pozycji, delty klocków. Rust wszedłby przy tysiącach połączeń. |

HUD pokazuje **fps** na żywo. Jeśli spada poniżej ~40 przy tej wielkości mapy, wtedy ruszamy meshing (greedy mesh) albo Godota — nie sam język skryptów.

## Testy

```bash
npm test
npm run build
```
