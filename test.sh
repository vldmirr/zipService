#!/bin/bash

# Тестовый скрипт для проверки zip-сервиса (без jq)

# 1. Проверяем, что сервер запущен
if ! curl -s http://localhost:8080 > /dev/null; then
    echo "Ошибка: Сервер не запущен на localhost:8080"
    exit 1
fi

echo "=== Тестирование создания задачи ==="
# Создаем новую задачу и извлекаем task_id (без jq)
response=$(curl -s -X POST http://localhost:8080/task/create)
task_id=$(echo "$response" | grep -oP '"task_id":"\K[^"]+')

if [ -z "$task_id" ]; then
    echo "Ошибка: Не удалось создать задачу"
    echo "Ответ сервера: $response"
    exit 1
fi

echo "Создана задача с ID: $task_id"
echo

echo "=== Тестирование добавления ссылок ==="
# Добавляем первую ссылку
echo "Добавляем PDF-файл 1..."
curl -v -X POST \
  -H "Content-Type: application/json" \
  -d '{"link":"https://zhjwpku.com/assets/pdf/AnIntroductionToProgrammingInGo.pdf"}' \
  "http://localhost:8080/links/add?task=$task_id"

# Добавляем вторую ссылку
echo "Добавляем PDF-файл 2..."
curl -v -X POST \
  -H "Content-Type: application/json" \
  -d '{"link":"https://www.cs.cmu.edu/~dst/LispBook/book.pdf"}' \
  "http://localhost:8080/links/add?task=$task_id"

# Даем серверу время обработать запросы
sleep 2

echo
echo "=== Проверка статуса задачи ==="
curl "http://localhost:8080/task/status?task=$task_id"
echo

echo "=== Тестирование загрузки архива ==="
echo "Загружаем архив..."
curl -v -o result.zip "http://localhost:8080/task/download-archive?task=$task_id"

# Проверяем, что архив создан
if [ -f "result.zip" ]; then
    file_size=$(du -h "result.zip" | cut -f1)
    echo "Архив успешно создан. Размер: $file_size"
    
    # Простая проверка архива
    if file result.zip | grep -q "Zip archive"; then
        echo "Архив валиден (Zip archive)"
        echo "Список файлов:"
        unzip -l result.zip || echo "Не удалось прочитать содержимое архива"
    else
        echo "Ошибка: это не ZIP архив"
        echo "Тип файла:"
        file result.zip
        exit 1
    fi
  
else
    echo "Ошибка: архив не был создан"
    exit 1
fi

echo
echo "=== Тестирование завершено успешно ==="