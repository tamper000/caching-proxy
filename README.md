# 🧠 Caching Proxy

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[English](README-en.md)

**Caching HTTP proxy server written in Go**
Простой кеширующий прокси-сервер, который перенаправляет запросы на удалённый сервер и кэширует результат[...]

> Идеально подходит для уменьшения нагрузки на внешние API путём кэширования ответов с возможностью гибкой конфигурации.

---

## 📌 Особенности

- Прозрачное кэширование HTTP-запросов
- Поддержка Redis в качестве хранилища кэша
- Гибкая настройка через YAML-файл
- Возможность указать blacklist URL-путей, которые не нужно кэшировать
- Очистка кэша по секретному ключу
- TODO: Добавить TTL JITTER

---

## ⚙️ Установка и запуск

### 1. Локальный запуск

```bash
go run cmd/main.go
```

Убедись, что конфигурационный файл `config.yaml` находится в корне проекта и имеет правильную конфигурацию.

### 2. Запуск через Docker

Собери образ (если нужно):

```bash
docker build -t caching-proxy .
```

Запусти контейнер:

```bash
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml -v $(pwd)/app.log:/app/app.log caching-proxy
```

### 3. Запуск через Docker Compose

Вы можете использовать пример `docker-compose` из этого репозитория:

```bash
docker-compose up --build
```

---

## 📄 Конфигурация

Конфигурация осуществляется через файл `config.yaml`.

### Пример конфигурации:

```yaml
server:
  origin: https://httpbin.org/ # Оригинальный URL
  # port: 1323
  secret: pls_delete_cache_maboy # Секрет для очистки кеша
  timeout: 10 # Таймаут до origin. В секундах
  ratelimit: # RateLimit осуществляется по IP
    rate: 20
    duration: 60 # в секундах

redis:
  addr: redis                    # Адрес Redis
  port: 6379                     # Порт Redis
  password:                      # Пароль (если используется)
  db:                            # Номер базы данных
  TTL: 5                         # Время жизни кэша в минутах

blacklist:
  - /uuid                        # Эти пути не будут кэшироваться
  - /delay/(.+)                  # Поддерживает regexp

logger:
  level: DEBUG                   # Сейчас поддерживает только DEBUG, INFO, ERROR
  file: app.log                  # Оставьте пустым для вывода в stdout
```

---

## 🧪 Использование

После запуска сервис доступен по адресу:

```
http://localhost:8080/
```

Все запросы перенаправляются на указанный `origin` (`https://httpbin.org/` в данном случае), а результаты кэшируются.

### Пример использования:

```bash
curl http://localhost:8080/ip
```

### Очистка кэша

Для очистки кэша отправь POST-запрос с секретным ключом:

```bash
curl -H "Authorization: Bearer pls_delete_cache_maboy" -X POST http://localhost:8080/clear
```

---

## 📁 Формат черного списка (Blacklist)

Поддерживаются простые строки и регулярные выражения:

```yaml
blacklist:
  - /uuid
  - /delay/(.+)
```

---

## 🏷 Статус кэширования (X-Cache)

При каждом ответе добавляется заголовок `X-Cache`, показывающий статус кэширования:

| Значение     | Описание                                      |
|--------------|-----------------------------------------------|
| `MISS`       | Данные не были закэшированы, запрос выполнен напрямую |
| `HIT`        | Ответ взят из кэша                             |
| `BYPASS`     | Запрос был исключен из кэширования (через blacklist) |


---

## 📦 Технологии

- Golang
- Redis
- YAML для конфигурации
- Docker

---

## 🧾 Лицензия

MIT License — см. [LICENSE](LICENSE) для деталей.
