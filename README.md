# Магазин мерча Avito

Этот проект представляет собой сервис внутреннего магазина мерча для сотрудников Avito. В сервисе реализованы возможности покупки мерча за монеты и передачи монет между сотрудниками.

## Оглавление

- [Описание проекта](#описание-проекта)
- [Особенности](#особенности)
- [Технические детали](#технические-детали)
- [Установка и запуск](#установка-и-запуск)
- [Структура проекта](#структура-проекта)
- [API](#api)
  - [Авторизация](#авторизация)
  - [Покупка мерча](#покупка-мерча)
  - [Передача монет](#передача-монет)
  - [История транзакций](#история-транзакций)
- [Тестирование](#тестирование)
- [Возникшие вопросы](#возникшие-вопросы)
- [Решенные задачи](#решенные-задачи)

## Описание проекта

Сервис позволяет сотрудникам Avito:
- **Покупать мерч**: приобретать товары за монеты.
- **Передавать монеты**: переводить монеты другим сотрудникам в знак благодарности или подарка.
- **Просматривать историю транзакций**: видеть список купленных товаров и детальную историю перемещений монет (кто отправлял и получал монеты).

При первой авторизации сотруднику автоматически создается профиль с 1000 монетами, и все операции проходят с проверкой на недопущение отрицательного баланса.

## Особенности

- **Безопасность**: Использование *JWT* для авторизации и защиты API. Хэширование паролей с помощью *Bcrypt*
- **Валидация транзакций**: Проверка достаточности средств перед совершением операций.
- **Масштабируемость**: Сервис рассчитан на до 100 тыс. сотрудников и 1k RPS, с SLI по времени ответа 50 мс и успешности 99.99%.
- **Тестирование**: Реализованы юнит-тесты и интеграционные/E2E-тесты для основных сценариев (покупка мерча и передача монет). Кроме того, есть возможность запуска нагрузочного теста для проверки требований по *RPS* и *SLI*

## Технические детали

- **Язык сервиса**: Go
- **База данных**: PostgreSQL, Redis
- **Контейнеризация**: Docker Compose для запуска сервиса и всех зависимостей
- **Порт доступа**: 8080 (сервис доступен по адресу [http://localhost:8080](http://localhost:8080))

## Установка и запуск

### Предварительные требования

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

### Шаги по установке

1. **Клонирование репозитория:**

    ```bash
    git clone https://github.com/myacey/avito-shop.git
    cd avito-store
    ```

2. **Запуск сервиса:**

    ```bash
    docker compose up --build
    ```

3. **Доступ к сервису:**

    Сервис будет доступен по адресу: [http://localhost:8080](http://localhost:8080)

## API
### Авторизация
- **POST /api/auth**

  **Описание:** Авторизация пользователя. При первой авторизации сотрудник создается автоматически.

  **Параметры запроса:**
  ```json
  {
    "username": "имя_пользователя",
    "password": "пароль"
  }
  ```
    Ответ: JWT токен для доступа к защищенным эндпоинтам.



### Покупка мерча
- **GET /api/buy/:item**

    **Описание**: Покупка мерча за монеты.

    **Параметры запроса:**
    ```json
    {
        "item_name": "t-shirt"
    }
    ```
    
    `Authorization: Bearer <JWT Token>`

### Передача монет
- **POST /api/sendCoin**

    **Описание**: Передача монет другому сотруднику.

    **Параметры запроса:**

    ```json
    {
        "toUser": "имя_получателя",
        "amount": 100
    }
    ```
    `Authorization: Bearer <JWT Token>`


### Информация о пользователе
- **GET /api/info**

    Описание: Получение истории транзакций, включая входящие и исходящие операции с монетами,
    предметы в инвентаре

    `Authorization: Bearer <JWT Token>`

    Ответ: вся информация о пользователе, включая историю транзакций и купленные предметы

## Тестирование

- **Юнит-тесты:**
    ```bash
    make unit_tests
    ```
    > для просмотра покрытия можно открыть `./tests/coverage.html`

- **Интеграционные тесты:**
    ```bash
    make integration_tests
    ```

- **E2E-тесты:**
    ```bash
    POSTGRES_TEST_DB_URL=<url> REDIS_TEST_DB_URL=<url> STATUS=testing bash run_e2e.bash
    ```
    > Для подсказки: `make e2e`

- **Нагрузочные тесты:**
    ```bash
    make load_test
    ```

- **Линтеры:**
    ```bash
    make lint
    ```

## Возникшие вопросы:
1. **Использование Redis для токенов**

    Большая часть приложения уже использует Postgres. **Так почему бы не добавить Redis**? Я и добавил...

2. **Использование микросервисов**

    Чисто в теории реализация микросервисов будет большим плюсом. Однако конкретно в данном проекте она выглядит чересчур неуместно, **ведь все компоненты плотно связаны между собой**. 
    Можно было бы вынести проверку токенов в некий отдельный api-gateway, но я посчитал, что конкретно в данном тестовом задании куда лучше будет написать и рассказать, почему здесь лучше использовать монолит, а не слепо пилить микросервисы "патаму что так нада, так нынче модно"...

3. **Слабое разделение моделей разных слоев приложения**

    Ввиду ограниченности функционала приложения я решил не сильно разбивать приложение на слои **с помощью использования *DTO*, *Service* и *DB*** моделей, хоть и в более больших проектах это куда более уместно...
    Тем не менее в некоторых частях программы данные модели все равно используются ввиду специфики требований по ответам сервиса.

    > (Разделение на хэндлеры, бизнес логику и репозиторий присутствует)

4. **Запуск тестов**

    Я сделал максимально простой (насколько смог) запуск всех тестов (*E2E*, *юнит-тесты*, *интеграционные* и *нагрузочные*) 
    Не знаю, надо ли это было или нет...

## Решенные задачи

- [x] Реализация API
- [x] RPS-1k, SLI-50ms 100% успешности
- [x] JWT токен авторизации (*Redis*)
- [x] Покрытие бизнес сценариев юнит-тестами
- [x] Интеграционные тесты на покупку предметов, получение информации о пользователе
- [x] E2E тесты на авторизацию и отправку монеток
- [x] Конфигурационный файл линтера (*.golangci-yaml*)