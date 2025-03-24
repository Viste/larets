# Larets

Larets - это менеджер репозиториев, аналог Nexus Repository Manager, написанный на Go. Larets позволяет создавать,
хранить и управлять Docker, Git и Helm репозиториями.

## Возможности

- **Docker репозитории**: хранение и проксирование Docker образов
- **Git репозитории**: хранение и проксирование Git репозиториев
- **Helm репозитории**: хранение и проксирование Helm чартов
- **Типы репозиториев**:
    - Hosted (хостинг): для хранения собственных артефактов
    - Proxy (прокси): для проксирования удаленных репозиториев
    - Group (группа): для объединения нескольких репозиториев (в разработке)

## Требования

- Go 1.22+
- PostgreSQL 12+
- Git
- Helm (для работы с Helm репозиториями)

## Установка

### Из исходного кода

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/Viste/larets.git
   cd larets
   ```

2. Создайте `.env` файл на основе примера:
   ```bash
   cp .env.example .env
   ```

3. Настройте переменные окружения в `.env` файле:
   ```
   DATABASE_URL=postgres://username:password@localhost:5432/larets?sslmode=disable
   ```

4. Соберите и запустите приложение:
   ```bash
   go build -o larets .
   ./larets
   ```

### С использованием Docker

1. Создайте `.env` файл на основе примера:
   ```bash
   cp .env.example .env
   ```

2. Настройте переменные окружения в `.env` файле.

3. Соберите и запустите Docker контейнер:
   ```bash
   docker build -t larets .
   docker run -d --name larets -p 8080:8080 --env-file .env -v larets-storage:/app/storage larets
   ```

## Конфигурация

Конфигурация осуществляется через переменные окружения или файл `.env`:

| Переменная        | Описание                                  | Значение по умолчанию |
|-------------------|-------------------------------------------|-----------------------|
| DATABASE_URL      | URL подключения к PostgreSQL              | -                     |
| ENABLE_DOCKER     | Включить поддержку Docker репозиториев    | true                  |
| ENABLE_GIT        | Включить поддержку Git репозиториев       | true                  |
| ENABLE_HELM       | Включить поддержку Helm репозиториев      | true                  |
| SERVER_PORT       | Порт HTTP сервера                         | 8080                  |
| BASE_URL          | Базовый URL для доступа к репозиториям    | http://localhost:8080 |
| STORAGE_PATH      | Путь к директории для хранения артефактов | ./storage             |
| DEFAULT_CACHE_TTL | TTL кеша для прокси-репозиториев (минуты) | 1440 (24 часа)        |
| ENABLE_AUTH       | Включить аутентификацию                   | false                 |
| ADMIN_USER        | Имя пользователя администратора           | admin                 |
| ADMIN_PASSWORD    | Пароль администратора                     | admin                 |

## API

### Общие эндпоинты

- `GET /api/health` - Проверка состояния сервера

### Docker репозитории

- `GET /api/docker/repositories` - Список Docker репозиториев
- `POST /api/docker/repositories` - Создание Docker репозитория
- `GET /api/docker/repositories/{name}` - Информация о Docker репозитории
- `GET /api/docker/images?repository={name}` - Список образов в репозитории
- `POST /api/docker/images?repository={name}&name={image}&tag={tag}` - Загрузка образа

### Git репозитории

- `GET /api/git/repositories` - Список Git репозиториев
- `POST /api/git/repositories` - Создание Git репозитория
- `GET /api/git/repositories/{name}` - Информация о Git репозитории
- `POST /api/git/sync/{name}` - Синхронизация прокси-репозитория

### Helm репозитории

- `GET /api/helm/repositories` - Список Helm репозиториев
- `POST /api/helm/repositories` - Создание Helm репозитория
- `GET /api/helm/repositories/{name}` - Информация о Helm репозитории
- `GET /api/helm/charts?repository={name}` - Список чартов в репозитории
- `POST /api/helm/charts?repository={name}&filename={filename}` - Загрузка чарта
- `POST /api/helm/sync/{name}` - Синхронизация прокси-репозитория

## Примеры использования

### Создание Docker репозитория

```bash
# Создание хостового репозитория
curl -X POST http://localhost:8080/api/docker/repositories \
  -H "Content-Type: application/json" \
  -d '{"name":"docker-local","description":"Локальный Docker репозиторий","type":"hosted"}'

# Создание прокси-репозитория
curl -X POST http://localhost:8080/api/docker/repositories \
  -H "Content-Type: application/json" \
  -d '{"name":"docker-proxy","description":"Прокси Docker Hub","type":"proxy","url":"https://registry-1.docker.io"}'
```

### Создание Git репозитория

```bash
# Создание хостового репозитория
curl -X POST http://localhost:8080/api/git/repositories \
  -H "Content-Type: application/json" \
  -d '{"name":"git-local","description":"Локальный Git репозиторий","type":"hosted"}'

# Создание прокси-репозитория
curl -X POST http://localhost:8080/api/git/repositories \
  -H "Content-Type: application/json" \
  -d '{"name":"git-proxy","description":"Прокси GitHub","type":"proxy","url":"https://github.com/Viste/larets.git"}'
```

### Создание Helm репозитория

```bash
# Создание хостового репозитория
curl -X POST http://localhost:8080/api/helm/repositories \
  -H "Content-Type: application/json" \
  -d '{"name":"helm-local","description":"Локальный Helm репозиторий","type":"hosted"}'

# Создание прокси-репозитория
curl -X POST http://localhost:8080/api/helm/repositories \
  -H "Content-Type: application/json" \
  -d '{"name":"helm-proxy","description":"Прокси Bitnami","type":"proxy","url":"https://charts.bitnami.com/bitnami"}'
```

### Загрузка Helm чарта

```bash
curl -X POST http://localhost:8080/api/helm/charts?repository=helm-local&filename=my-chart-0.1.0.tgz \
  --data-binary @my-chart-0.1.0.tgz
```

### Использование Helm репозитория в Helm CLI

```bash
# Добавление репозитория
helm repo add larets-repo http://localhost:8080/helm/helm-local

# Обновление репозитория
helm repo update

# Поиск чартов
helm search repo larets-repo/
```
