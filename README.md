# shortify

Сервис сокращения URL в рамках обучения Yandex Practicum.

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Запросы для тестирования

### Запрос создания сокращённого URL

```sh
curl -v -X POST http://localhost:8080/ \
    -H "Content-Type: text/plain" \
    -d "https://ya.ru"
```

```sh
curl -v -X POST http://localhost:8080/api/shorten \
    -H "Content-Type: application/json" \
    -d '{"url":"https://ya.ru"}'
```

### Запрос переадресации на оригинальный URL
```sh
curl -v http://localhost:8080/{xxx}
```

## DB

Для развертывания PostgreSQL используется:
```bash
docker run -d \
  --name local-postgres \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e PGDATA=/var/lib/postgresql/data/pgdata \
  -v ./.db:/var/lib/postgresql/data \
  postgres:15-alpine
```


## Профилирование
Для запуска сервиса с pprof можно использовать флаг `-profiler` или `PROFILER_ENABLED=true`
Снятые профили в `profiles`

Использование hey:
```
hey -n 100000 -c 50 \
    -H "Content-Type: application/json" \
    -m POST \
    -d '{"url":"https://www.ya.ru"}' \
    http://localhost:8080/api/shorten
```

Вывод после перехода на использование другой библиотеки gzip:
```
$ go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof                                                                                                                               [18:26:08]
File: main
Type: inuse_space
Time: 2026-01-27 18:27:12 MSK
Duration: 60.01s, Total samples = 283.25MB 
Showing nodes accounting for -98.56MB, 34.80% of 283.25MB total
Dropped 2 nodes (cum <= 1.42MB)
      flat  flat%   sum%        cum   cum%
 -178.05MB 62.86% 62.86%  -217.48MB 76.78%  compress/flate.NewWriter (inline)
   50.81MB 17.94% 44.92%    50.81MB 17.94%  github.com/klauspost/compress/flate.newFastEnc (inline)
  -38.43MB 13.57% 58.49%   -38.43MB 13.57%  compress/flate.(*compressor).initDeflate (inline)
   36.98MB 13.06% 45.43%    36.98MB 13.06%  github.com/klauspost/compress/flate.(*fastGen).Reset
   35.84MB 12.65% 32.78%    94.09MB 33.22%  github.com/klauspost/compress/flate.NewWriter (inline)
    7.45MB  2.63% 30.15%    58.25MB 20.57%  github.com/klauspost/compress/flate.(*compressor).init
   -4.50MB  1.59% 31.74%    -4.50MB  1.59%  encoding/base64.(*Encoding).EncodeToString
   -3.01MB  1.06% 32.80%    -3.01MB  1.06%  compress/flate.(*huffmanEncoder).generate
   -2.64MB  0.93% 33.73%    -2.64MB  0.93%  github.com/hydra13/shortify/internal/repositories/inmemory_db.(*InMemoryDB).Add
    2.50MB  0.88% 32.85%     2.50MB  0.88%  encoding/json.(*decodeState).literalStore
   -1.50MB  0.53% 33.38%    -1.50MB  0.53%  github.com/google/uuid.UUID.String (inline)
   -0.50MB  0.18% 33.56%    -0.50MB  0.18%  bufio.NewReaderSize (inline)
   -0.50MB  0.18% 33.74%    -0.50MB  0.18%  runtime.allocm
   -0.50MB  0.18% 33.91%       -1MB  0.35%  compress/flate.newHuffmanBitWriter (inline)
    0.50MB  0.18% 33.74%     0.50MB  0.18%  runtime.malg
   -0.50MB  0.18% 33.91%    -0.50MB  0.18%  net/textproto.MIMEHeader.Add (inline)
   -0.50MB  0.18% 34.09%    -0.50MB  0.18%  encoding/json.NewDecoder (inline)
   -0.50MB  0.18% 34.27%       -1MB  0.35%  net/http.(*conn).readRequest
   -0.50MB  0.18% 34.44%    -0.50MB  0.18%  compress/flate.newHuffmanEncoder (inline)
   -0.50MB  0.18% 34.62%    -0.50MB  0.18%  net/url.parse
   -0.50MB  0.18% 34.80%    -0.50MB  0.18%  net/textproto.readMIMEHeader
         0     0% 34.80%    -0.50MB  0.18%  bufio.NewReader (inline)
         0     0% 34.80%    -3.01MB  1.06%  compress/flate.(*Writer).Close (inline)
         0     0% 34.80%    -3.01MB  1.06%  compress/flate.(*compressor).close
         0     0% 34.80%    -3.01MB  1.06%  compress/flate.(*compressor).deflate
         0     0% 34.80%   -39.43MB 13.92%  compress/flate.(*compressor).init
         0     0% 34.80%    -3.01MB  1.06%  compress/flate.(*compressor).writeBlock
         0     0% 34.80%       -1MB  0.35%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0% 34.80%    -3.01MB  1.06%  compress/flate.(*huffmanBitWriter).writeBlock
         0     0% 34.80%    -3.01MB  1.06%  compress/gzip.(*Writer).Close
         0     0% 34.80%  -217.48MB 76.78%  compress/gzip.(*Writer).Write
         0     0% 34.80%     2.50MB  0.88%  encoding/json.(*Decoder).Decode
         0     0% 34.80%  -123.39MB 43.56%  encoding/json.(*Encoder).Encode
         0     0% 34.80%     2.50MB  0.88%  encoding/json.(*decodeState).object
         0     0% 34.80%     2.50MB  0.88%  encoding/json.(*decodeState).unmarshal
         0     0% 34.80%     2.50MB  0.88%  encoding/json.(*decodeState).value
         0     0% 34.80%   -96.55MB 34.09%  github.com/go-chi/chi/v5.(*ChainHandler).ServeHTTP
         0     0% 34.80%  -128.03MB 45.20%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
         0     0% 34.80%   -96.55MB 34.09%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 34.80%   -96.55MB 34.09%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 34.80%  -128.03MB 45.20%  github.com/hydra13/shortify/internal/handlers/get_short_url_by_json_handler.(*Handler).Handle
         0     0% 34.80%    33.98MB 12.00%  github.com/hydra13/shortify/internal/middlewares/compresser.(*compressWriter).Close
         0     0% 34.80%  -123.39MB 43.56%  github.com/hydra13/shortify/internal/middlewares/compresser.(*compressWriter).Write
         0     0% 34.80%   -96.55MB 34.09%  github.com/hydra13/shortify/internal/middlewares/compresser.CompresserMiddleware.func1
         0     0% 34.80%  -123.39MB 43.56%  github.com/hydra13/shortify/internal/middlewares/logger.(*loggingResponseWriter).Write
         0     0% 34.80%    -1.50MB  0.53%  github.com/hydra13/shortify/internal/services/auth.(*AuthService).GetOrCreateUser
         0     0% 34.80%    -0.50MB  0.18%  github.com/hydra13/shortify/internal/services/auth.(*AuthService).SetAuthCookie
         0     0% 34.80%    -1.50MB  0.53%  github.com/hydra13/shortify/internal/services/auth.(*AuthService).generateUserID
         0     0% 34.80%    -4.50MB  1.59%  github.com/hydra13/shortify/internal/services/short_id_generator.Generator.GenerateShortID
         0     0% 34.80%    -7.14MB  2.52%  github.com/hydra13/shortify/internal/services/shorter.Shorter.Create
         0     0% 34.80%    -2.64MB  0.93%  github.com/hydra13/shortify/internal/services/urls_keeper.(*UrlsKeeper).Save
         0     0% 34.80%    36.98MB 13.06%  github.com/klauspost/compress/flate.(*Writer).Close (inline)
         0     0% 34.80%    36.98MB 13.06%  github.com/klauspost/compress/flate.(*compressor).close
         0     0% 34.80%    36.98MB 13.06%  github.com/klauspost/compress/flate.(*compressor).storeFast
         0     0% 34.80%    36.98MB 13.06%  github.com/klauspost/compress/gzip.(*Writer).Close
         0     0% 34.80%    94.09MB 33.22%  github.com/klauspost/compress/gzip.(*Writer).Write
         0     0% 34.80%  -130.53MB 46.08%  main.main.func2.NewAuthMiddleware.3.1
         0     0% 34.80%  -130.53MB 46.08%  main.main.func2.NewLoggerMiddleware.2.1
         0     0% 34.80%   -98.56MB 34.80%  net/http.(*conn).serve
         0     0% 34.80%   -96.55MB 34.09%  net/http.HandlerFunc.ServeHTTP
         0     0% 34.80%    -0.50MB  0.18%  net/http.Header.Add (inline)
         0     0% 34.80%    -0.50MB  0.18%  net/http.SetCookie
         0     0% 34.80%    -0.50MB  0.18%  net/http.newBufioReader
         0     0% 34.80%       -1MB  0.35%  net/http.readRequest
         0     0% 34.80%   -96.55MB 34.09%  net/http.serverHandler.ServeHTTP
         0     0% 34.80%    -0.50MB  0.18%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0% 34.80%    -0.50MB  0.18%  net/url.ParseRequestURI
         0     0% 34.80%    -0.50MB  0.18%  runtime.mstart
         0     0% 34.80%    -0.50MB  0.18%  runtime.mstart0
         0     0% 34.80%    -0.50MB  0.18%  runtime.mstart1
         0     0% 34.80%    -0.50MB  0.18%  runtime.newm
         0     0% 34.80%     0.50MB  0.18%  runtime.newproc.func1
         0     0% 34.80%     0.50MB  0.18%  runtime.newproc1
         0     0% 34.80%    -0.50MB  0.18%  runtime.resetspinning
         0     0% 34.80%    -0.50MB  0.18%  runtime.schedule
         0     0% 34.80%    -0.50MB  0.18%  runtime.startm
         0     0% 34.80%     0.50MB  0.18%  runtime.systemstack
         0     0% 34.80%    -0.50MB  0.18%  runtime.wakep
```

## Форматирование
Установка goimports:
```
go install golang.org/x/tools/cmd/goimports@latest
```
Использование:
```
goimports -local "github.com/hydra13/shortify" -w main.go
```

## Документация
Установка godoc:
```
go install -v golang.org/x/tools/cmd/godoc@latest
```
Запуск:
```
godoc -http=:8080
```
После выполнения этой команды в корне проекта, документация будет доступна по адресу:
http://localhost:8080/pkg/github.com/hydra13/shortify/?m=all