# BOXEL

MVP studia do **grania i tworzenia** kwadratowych gier. Wygląd OG (grube klocki, ostre krawędzie), programowanie jak w Scratchu.

Otwórz folder `boxel`, zainstaluj zależności i odpal studio:

```bash
cd boxel
npm install
npm run dev
```

Wejdź na http://localhost:5173

## Co już jest

- Edytor świata 24×16×24 z klockami: trawa, lawa, lód, sprężyna, monety, meta…
- Tryb **Graj**: WASD / strzałki, spacja, pad na ekranie
- Tryb **Klocki**: zdarzenia i akcje (punkty, wiadomość, wygrana, pętla, „jeśli punkty ≥”)
- Szablony: Parkour monetkowy, Piaskownica, Arena
- Zapis w przeglądarce + plik `.boxel.json`
- Cofnij (Ctrl+Z), limity skryptów (żeby pętla dziecka nie zawiesiła gry)

## Sterowanie

| Akcja | Jak |
| --- | --- |
| Stawianie | lewy przycisk myszy |
| Kasowanie | prawy przycisk albo gumka |
| Obrót kamery | przeciągnij lewy przycisk |
| Skok | spacja / przycisk SKOK |
| Klocki 1–9 | klawisze 1–9, `0` lub `X` = gumka |

Szablon **Parkour monetkowy** ma gotowe skrypty: zbierz 3 monety (30 pkt) i dobiegnij do złotej flagi.

## Testy

```bash
npm test
npm run build
```
