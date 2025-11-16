# Load Testing

Это директория содержит скрипты нагрузочного тестирования для сервиса назначения ревьюеров PR с использованием [k6](https://k6.io/).

## Обзор

Нагрузочные тесты проверяют соответствие сервиса требованиям по производительности:
- **RPS**: 5 запросов в секунду
- **SLI времени ответа**: 95% запросов должны обрабатываться быстрее 300 мс
- **SLI успешности**: 99.9% успешных запросов

## Предварительные требования

### 1. Установка k6

**macOS:**
```bash
brew install k6
```

**Linux:**
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Windows:**
```powershell
choco install k6
```

Или скачайте бинарный файл с [https://k6.io/docs/get-started/installation/](https://k6.io/docs/get-started/installation/)

### 2. Запуск сервиса

Убедитесь, что сервис запущен:

```bash
# Из корневой директории проекта
docker compose up --build
```

Сервис будет доступен по адресу `http://localhost:8080`

## Структура файлов

```
loadtest/
├── README.md              # Эта документация
├── seed_data.js           # Скрипт наполнения БД тестовыми данными
├── stats_assignments.js   # Тест read-only эндпоинта статистики
├── create_pr.js           # Тест write-heavy эндпоинта создания PR
└── run_all.sh             # Скрипт для запуска всех тестов
```

## Быстрый старт

### Способ 1: Автоматический запуск всех тестов

```bash
cd loadtest
chmod +x run_all.sh
./run_all.sh
```

### Способ 2: Пошаговый запуск

1. **Наполнение БД тестовыми данными**

```bash
cd loadtest
k6 run seed_data.js
```

Этот скрипт создаёт:
- 2 тестовые команды
- 8 тестовых пользователей

2. **Тестирование эндпоинта статистики**

```bash
k6 run stats_assignments.js
```

3. **Тестирование создания PR**

```bash
k6 run create_pr.js
```

## Описание тестов

### 1. `seed_data.js` - Наполнение БД

Создаёт тестовые данные, необходимые для нагрузочного тестирования.

**Создаваемые данные:**
- Команда `load-test-team-1` с 5 пользователями
- Команда `load-test-team-2` с 3 пользователями
- Все пользователи активны (`is_active: true`)

**Использование:**
```bash
k6 run seed_data.js
# или с кастомным URL
BASE_URL=http://localhost:8080 k6 run seed_data.js
```

### 2. `stats_assignments.js` - Тест GET /stats/assignments

Тестирует read-only эндпоинт статистики, который возвращает количество назначений по пользователям и по PR.

**Профиль нагрузки:**
- 0 → 10 VUs за 10 секунд (разгон)
- 10 → 50 VUs за 30 секунд (основная нагрузка)
- 50 → 0 VUs за 10 секунд (спад)

**Проверки:**
- HTTP статус 200
- Наличие тела ответа
- Content-Type: application/json

**Пороги (thresholds):**
- 95-й перцентиль времени ответа < 300 мс
- Уровень ошибок < 0.1% (99.9% успешности)

**Использование:**
```bash
k6 run stats_assignments.js
# или с кастомным URL
BASE_URL=http://localhost:8080 k6 run stats_assignments.js
```

### 3. `create_pr.js` - Тест POST /pullRequest/create

Тестирует write-heavy эндпоинт создания PR с автоматическим назначением ревьюеров.

**Профиль нагрузки:**
- 0 → 5 VUs за 10 секунд (разгон)
- 5 → 20 VUs за 30 секунд (основная нагрузка)
- 20 → 0 VUs за 10 секунд (спад)

**Особенности:**
- Генерирует уникальный `pull_request_id` для каждого запроса
- Использует случайных авторов из тестовых пользователей
- Обрабатывает успешные создания (201) и конфликты (409)

**Проверки:**
- HTTP статус 201 (создан) или 409 (конфликт)
- Наличие тела ответа
- Content-Type: application/json

**Пороги:**
- 95-й перцентиль времени ответа < 300 мс
- Уровень ошибок < 5% (допускаются конфликты при параллельных запросах)

**Предварительные требования:**
Перед запуском необходимо выполнить `seed_data.js`

**Использование:**
```bash
k6 run create_pr.js
# или с кастомным URL
BASE_URL=http://localhost:8080 k6 run create_pr.js
```

## Переменные окружения

Все скрипты поддерживают переменную `BASE_URL`:

```bash
# Для локального запуска (по умолчанию)
BASE_URL=http://localhost:8080 k6 run stats_assignments.js

# Для docker-compose окружения
BASE_URL=http://avito-internship-backend-service-test:8080 k6 run stats_assignments.js
```

## Интерпретация результатов

После выполнения теста k6 выводит подробную статистику:

```
     ✓ status is 200
     ✓ response has body
     ✓ content type is JSON

     checks.........................: 100.00% ✓ 2400      ✗ 0
     data_received..................: 1.2 MB  24 kB/s
     data_sent......................: 240 kB  4.8 kB/s
     http_req_blocked...............: avg=1.23ms   min=1µs     med=3µs     max=45.67ms  p(90)=5µs     p(95)=7µs
     http_req_connecting............: avg=1.15ms   min=0s      med=0s      max=43.21ms  p(90)=0s      p(95)=0s
   ✓ http_req_duration..............: avg=45.32ms  min=12.34ms med=43.21ms max=156.78ms p(90)=67.89ms p(95)=89.12ms
       { expected_response:true }...: avg=45.32ms  min=12.34ms med=43.21ms max=156.78ms p(90)=67.89ms p(95)=89.12ms
   ✓ http_req_failed................: 0.00%   ✓ 0         ✗ 800
     http_req_receiving.............: avg=123µs    min=45µs    med=98µs    max=2.34ms   p(90)=234µs   p(95)=345µs
     http_req_sending...............: avg=23µs     min=8µs     med=18µs    max=456µs    p(90)=34µs    p(95)=45µs
     http_req_tls_handshaking.......: avg=0s       min=0s      med=0s      max=0s       p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=45.17ms  min=12.23ms med=43.09ms max=156.45ms p(90)=67.78ms p(95)=88.98ms
     http_reqs......................: 800     16/s
     iteration_duration.............: avg=145.56ms min=112.45ms med=143.32ms max=256.89ms p(90)=167.89ms p(95)=189.23ms
     iterations.....................: 800     16/s
     vus............................: 1       min=0       max=50
     vus_max........................: 50      min=50      max=50
```

**Ключевые метрики:**

- **checks**: Процент успешных проверок (должен быть близок к 100%)
- **http_req_duration**: Время ответа сервера
  - `p(95)` — 95-й перцентиль (должен быть < 300 мс согласно SLI)
  - `avg` — среднее время ответа
  - `max` — максимальное время ответа
- **http_req_failed**: Процент неуспешных запросов (должен быть < 0.1% для 99.9% SLI)
- **http_reqs**: Количество запросов в секунду (RPS)
- **vus**: Количество виртуальных пользователей

**Критерии успеха:**

✅ Тест пройден, если:
- `http_req_duration p(95) < 300ms`
- `http_req_failed < 0.1%` (для read-only эндпоинтов)
- Все `checks` проходят успешно

❌ Тест провален, если:
- Пороги (thresholds) нарушены
- Высокий процент ошибок
- Время ответа превышает SLI требования

## Запуск тестов в Docker

Если вы хотите запустить тесты внутри Docker-сети (например, для тестирования контейнера):

```bash
# Запустите k6 с доступом к Docker-сети
docker run --rm -i --network avito-internship-network \
  -v $(pwd)/loadtest:/scripts \
  grafana/k6 run \
  -e BASE_URL=http://avito-internship-backend-service-test:8080 \
  /scripts/stats_assignments.js
```

## Рекомендации

1. **Перед тестированием**: Убедитесь, что сервис стабильно работает и БД наполнена тестовыми данными
2. **Очистка данных**: После тестов можно пересоздать контейнеры для очистки БД:
   ```bash
   docker compose down -v
   docker compose up --build
   ```
3. **Масштабирование тестов**: Для более интенсивной нагрузки измените параметры `stages` в скриптах
4. **Мониторинг**: Следите за логами сервиса во время тестирования для выявления проблем:
   ```bash
   docker compose logs -f backend
   ```

## Расширение тестов

Для добавления новых сценариев нагрузочного тестирования:

1. Создайте новый `.js` файл в директории `loadtest/`
2. Используйте существующие скрипты как шаблон
3. Определите профиль нагрузки в `options.stages`
4. Укажите пороги в `options.thresholds`
5. Добавьте скрипт в `run_all.sh` при необходимости

## Troubleshooting

**Проблема**: Тест не может подключиться к сервису
```
WARN[0000] Request Failed   error="Get \"http://localhost:8080/healthz\": dial tcp [::1]:8080: connect: connection refused"
```
**Решение**: Убедитесь, что сервис запущен через `docker compose up`

---

**Проблема**: Ошибка "Test data not seeded"
```
ERROR: Test data not seeded. Run seed_data.js before this test.
```
**Решение**: Запустите скрипт наполнения БД:
```bash
k6 run seed_data.js
```

---

**Проблема**: Высокий процент ошибок 404
```
✗ status is 201 or 409
```
**Решение**: Проверьте, что тестовые пользователи существуют в БД (запустите `seed_data.js`)

---

**Проблема**: Превышение порогов времени ответа
```
✗ http_req_duration..............: p(95)=450ms (threshold: p(95)<300)
```
**Решение**: 
- Проверьте загрузку системы
- Оптимизируйте запросы к БД
- Увеличьте ресурсы контейнера PostgreSQL
- Проверьте индексы в БД

