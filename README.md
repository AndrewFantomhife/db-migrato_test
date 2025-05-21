# SQL Migrator for PostgreSQL
Простой инструмент управления миграциями базы данных, написанный на Go с использованием библиотеки [goose]. Поддерживает стандартные операции: применение, откат, создание миграций.

Как использовать:
Локальный запуск (через Go):

go run main.go -cmd=up -db="host=localhost port=5432 user=postgres password=232003 dbname=mydb sslmode=disable"

Откатить последнюю миграцию:

go run main.go -cmd=down -db="host=localhost port=5432 user=postgres password=232003 dbname=mydb sslmode=disable"

Проверить статус миграций:

go run main.go -cmd=status -db="host=localhost port=5432 user=postgres password=232003 dbname=mydb sslmode=disable"

Создать новую миграцию (SQL):

go run main.go -cmd=create -type=sql create_users_table

Важные файлы:

main.go — основная логика CLI

internal/migrator/*.sql — каталог с миграциями