# BOXEL

Aplikacja desktopowa do **budowania i grania** w kwadratowe światy z klockami (styl OG) i programowaniem jak w Scratchu. Jest multiplayer, czat i znajomi.

## Uruchomienie (aplikacja, nie strona)

Potrzebujesz Node.js i **Go 1.22+**.

```bash
cd boxel
npm install
npm run app
```

To odpala:

1. serwer pokoi w **Go** (`ws://127.0.0.1:8787`) — goroutine na każdego gracza
2. silnik UI
3. **okno Electron** — BOXEL → Nowe okno (drugi gracz) otwiera drugie konto na tym samym PC

Bez Electrona (tylko do testów silnika): `npm run server` i `npm run dev`.

## Multiplayer, znajomi, czat

- Każdy ma **kod znajomego** `BX-XXXXXX`. Wpisz kod kolegi i kliknij Dodaj.
- **Załóż pokój** — dostajesz kod 4-cyfrowy. Znajomy klika Zaproś albo wpisuje kod.
- W pokoju widać innych graczy (kolorowe kostki), wspólne stawianie klocków i czat.
- Czat pokoju + szept tylko do znajomych. Linki, Discord i wulgaryzmy są blokowane. Są gotowe bąbelki: Cześć!, Chodź tu!, Pomocy!…

## Język i wątki

TypeScript w Node jest jednowątkowy — to był zły wybór na **serwer gry**, nie na cały projekt.

| Warstwa | Język | Dlaczego |
| --- | --- | --- |
| Studio + klocki + voxel 24³ | TypeScript + Three.js | Edytor i UI nie są wąskim gardłem. Test `physics speed`: 2000 klatek fizyki w << 250 ms. Przepisywanie tego w Rust/Go teraz nie przyspieszy klikania klocków. |
| Okno aplikacji | Electron | To jest **aplikacja**, nie karta przeglądarki. |
| Pokoje / czat / sync pozycji i klocków | **Go** | Tu TS bolał: jedno pętla eventów na wszystkich graczy. W Go każdy socket ma swoją goroutine (odczyt + osobny write pump). Mutex tylko na stan pokoju. |

**Dlaczego Go, a nie Rust (na teraz):** ten sam protokół JSON, szybciej dowieźć multiplayer, goroutines mapują się 1:1 na graczy. Rust (tokio albo później Bevy/godot-rust) ma sens, gdy świat urośnie i CPU silnika (meshing, fizyka, 50+ graczy) stanie się wąskim gardłem — wtedy przenosimy silnik, nie edytor klocków.

HUD pokazuje **fps** na żywo. Jeśli spada poniżej ~40 przy tej wielkości mapy, wtedy ruszamy meshing (greedy mesh) albo natywny silnik — nie cały stack od zera.

## Testy

```bash
npm test
npm run e2e   # wymaga działającego `npm run server`
npm run build
```
