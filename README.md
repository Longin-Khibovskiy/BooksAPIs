# Books API

### authorize on website https://developer.nytimes.com/accounts/login and get apikey

```https://api.nytimes.com/svc/books/v3/lists/overview.json?api-key=yourkey```

```apikey in .env file```

#### Добавленные паки:
```terminaloutput
go get github.com/gorilla/mux
go get github.com/joho/godotenv
go get github.com/lib/pq
go get github.com/a-h/templ
```
### Запуск проекта
```terminaloutput
# 1 раз при первом запуске
echo "ENV=local" >> .env
echo "API_KEY=your_api_key_NYT" >> .env
echo "DB_PORT=5432" >> .env
echo "DB_USER=user" >> .env
echo "DB_PASSWORD=password" >> .env
echo "DB_NAME=booksdb" >> .env
docker compose build

# запуск проекта
docker compose up
```

### Дополнительные команды
```terminaloutput
# Остановить контейнеры
docker compose down
# Показать логи всех сервисов
docker compose logs
# Посмотреть статус контейнеров
docker compose ps
# Перезапустить контейнеры
docker compose restart
```