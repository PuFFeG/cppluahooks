#!/bin/bash

# Жестко заданный путь к файлу конфигурации
CONFIG_PATH="/pal/cool.json"

# Парсинг аргументов командной строки
while [[ $# -gt 0 ]]; do
    key="$1"

    case $key in
        -playerid)
        PLAYER_ID="$2"
        shift
        shift
        ;;
        *)
        echo "Неизвестный аргумент: $1"
        exit 1
        ;;
    esac
done

# Проверка обязательного аргумента -playerid
if [ -z "$PLAYER_ID" ]; then
    echo "Ошибка: не передан обязательный параметр -playerid"
    exit 1
fi

# Проверка существования файла конфигурации
if [ ! -f "$CONFIG_PATH" ]; then
    echo "Ошибка: файл конфигурации не найден"
    exit 1
fi

# Перебор элементов конфигурации и выполнение команды ARRCON для каждого элемента
for item in $(jq -c '.items[]' "$CONFIG_PATH"); do
    ITEM=$(echo "$item" | jq -r '.item')
    QUANTITY=$(echo "$item" | jq -r '.quantity')

    # Подготовка команды ARRCON для текущего элемента
    COMMAND="./ARRCON -H 192.168.31.109 -P 25575 -p 236006 \"give $PLAYER_ID $ITEM $QUANTITY\""
    echo "Выполняем команду: $COMMAND"

    # Выполнение команды ARRCON и захват вывода
    OUTPUT=$(eval "$COMMAND" 2>&1)
    if [ $? -ne 0 ]; then
        echo "Ошибка выполнения команды ARRCON: $OUTPUT"
        exit 1
    fi

    echo "Команда выполнена успешно: $COMMAND"
done

