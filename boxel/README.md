# BOXEL

Natywna aplikacja **Go** do budowania i grania w kwadratowe światy z klockami (styl OG) i programowaniem jak w Scratchu. Multiplayer, czat i znajomi działają w tym samym binariu.

TypeScript / Electron / Node zostały usunięte. Silnik, studio 3D, edytor klocków i serwer pokoi są w Go (goroutine na gracza).

## Uruchomienie

Potrzebujesz **Go 1.22+** (przy pierwszym `go run` toolchain może pobrać nowszą wersję).

Na Linuxie do okna gry:

```bash
sudo apt install gcc pkg-config libgl1-mesa-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev libxxf86vm-dev libasound2-dev
```

```bash
cd boxel
go test ./...
go run ./cmd/boxel
```

To otwiera okno BOXEL i lokalny serwer `ws://127.0.0.1:8787`. Drugi gracz na tym samym PC: uruchom `go run ./cmd/boxel` jeszcze raz.

Tylko serwer (np. na Render):

```bash
go run ./cmd/server
# HOST=0.0.0.0 PORT=8787
```

## Sterowanie

- **Buduj** — LPM stawia, PPM kasuje, przeciąganie obraca kamerę, kółko przybliża. 1–9 klocki, 0 gumka, Ctrl+Z cofnij.
- **Klocki** — żółte zdarzenia i akcje jak w Scratchu.
- **Graj** — WASD / strzałki, spacja, albo d-pad i SKOK na ekranie.
- **Znajomi** — kod `BX-XXXXXX`, pokój 4-cyfrowy, czat w pokoju, szept tylko do znajomych. Linki i wulgaryzmy są blokowane.

Zapisy lądują w katalogu konfiguracyjnym systemu (`save.json`, `profile.json`). Serwer na innym hoście: `BOXEL_WS=ws://adres:8787 go run ./cmd/boxel`.

## Klocki kodu — lokalnie i na serwerze

- **Lokalnie (solo / ten komputer):** klocki są w pamięci studia i w `save.json`. GRAJ odpala interpreter na Twoim PC.
- **W pokoju:** gospodarz wysyła klocki razem ze światem. Dołączający dostają tę samą listę skryptów. Zmiana klocków w edytorze (gospodarz) idzie do wszystkich. Zbieranie monety kasuje klocek w świecie pokoju i pokazuje efekt (wiadomość) innym graczom.
- **Dedykowany serwer** (`cmd/server`) trzyma te same skrypty co lokalny hub w oknie BOXEL.

Tylko gospodarz edytuje klocki w pokoju. Gdy wyjdzie, następna osoba w pokoju przejmuje rolę gospodarza.

## Układ kodu

| Pakiet | Co robi |
| --- | --- |
| `cmd/boxel` | okno gry (Ebitengine) + lokalny hub |
| `cmd/server` | sam serwer WebSocket |
| `internal/engine` | świat, fizyka, interpreter klocków, szablony |
| `internal/studio` | tryby Buduj / Graj / Klocki |
| `internal/render` | voxel 3D (painter’s algorithm) |
| `internal/hub` | pokoje, znajomi, czat |
| `internal/app` | UI |

## Testy

```bash
go test ./...
```

Silnik i hub nie potrzebują okna. Aplikacja desktopowa to `go run ./cmd/boxel`.
