#!/bin/bash

sudo apt install -y uhubctl

# ========== НАСТРОЙКИ ==========
OFF_DURATION=120           # на сколько секунд выключать USB (сек)
CYCLE_INTERVAL_MIN=180     # периодичность циклов (минуты)
# ===============================

INTERVAL_SEC=$((CYCLE_INTERVAL_MIN * 60))
WAIT_AFTER_ON=$((INTERVAL_SEC - OFF_DURATION))
[ $WAIT_AFTER_ON -lt 0 ] && WAIT_AFTER_ON=0

echo "=== Управление питанием USB (Raspberry Pi 5) ==="
echo "Отключение на $OFF_DURATION сек, повтор каждые $CYCLE_INTERVAL_MIN мин"
echo "Остановить: Ctrl+C (USB будет включён обратно)"
echo "================================================"

# Проверка uhubctl
if ! command -v uhubctl &> /dev/null; then
    echo "ОШИБКА: утилита uhubctl не установлена."
    echo "Установите: sudo apt update && sudo apt install -y uhubctl"
    exit 1
fi

# Проверка прав root
if [ "$EUID" -ne 0 ]; then
    echo "ОШИБКА: запустите с sudo: sudo $0"
    exit 1
fi

# Функция отключения всех 4 хабов
usb_off_all() {
    echo "$(date '+%H:%M:%S') - Выключаем питание USB (хабы 1-4)..."
    for hub in 1 2 3 4; do
        echo "  Отключаю хаб $hub..."
        uhubctl -l $hub -a 0 2>/dev/null
        if [ $? -ne 0 ]; then
            echo "    (хаб $hub не поддерживает отключение или уже выключен)"
        fi
    done
}

# Функция включения всех 4 хабов
usb_on_all() {
    echo "$(date '+%H:%M:%S') - Включаем питание USB (хабы 1-4)..."
    for hub in 1 2 3 4; do
        echo "  Включаю хаб $hub..."
        uhubctl -l $hub -a 1 2>/dev/null
        if [ $? -ne 0 ]; then
            echo "    (хаб $hub не поддерживает включение или уже включён)"
        fi
    done
}

# Перехват Ctrl+C для корректного выхода с включением USB
trap 'echo ""; echo "Сигнал остановки. Включаем USB..."; usb_on_all; echo "Скрипт завершён."; exit 0' INT TERM

# Основной цикл
while true; do
    usb_off_all
    sleep $OFF_DURATION
    usb_on_all
    if [ $WAIT_AFTER_ON -gt 0 ]; then
        echo "Ожидание $WAIT_AFTER_ON сек до следующего цикла..."
        sleep $WAIT_AFTER_ON
    fi
done