# Bietrunde

Programm für eine Bietrunde von [Baarfood](https://baarfood.de/)


## Starten

Das Programm ist in Binary. Es enthält alle erforderlichen Daten.

Zum bauen des Binary wird die Programmiersprache [Go](https://go.dev/) benötigt. Führe anschließend aus:

```bash
CGO_ENABLED=0 go build
```

Anschließend kann die Binary ausgeführt werden:

```bash
./bietrunde
```

Beim ausführen wird eine Datei `config.toml` angelegt. Diese kann bearbeitet
werden, um die Anwendung zu konfigurieren. Anschließend muss die Anwendung neu
gestartet werden.

Außerdem wird die Datei `db.jsonl` angelegt. Hierbei handelt es sich um die
Datenbank.

Wenn das Programm hinter einem Proxy läuft, dann achte darauf, dass die Ausgabe
nicht gebuffert wird. Zum Beispiel in nginx:

```nginx
proxy_buffering off;
proxy_pass http://localhost:9600;
```


# Entwicklung

Um die Templates anpassen zu können muss [templ](https://templ.guide/)
installiert sein. Anschließend können die Templates mit

```bash
go generate ./...
```

neu gebaut werden.
