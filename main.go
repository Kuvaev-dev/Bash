package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func checkUtil(name string) {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("ОШИБКА: утилита %s не установлена.\n", name)
		fmt.Printf("Установите: sudo apt update && sudo apt install -y %s\n", name)
		os.Exit(1)
	}
}

func checkRoot() {
	if os.Geteuid() != 0 {
		fmt.Printf("ОШИБКА: запустите с sudo: sudo %s\n", os.Args[0])
		os.Exit(1)
	}
}

func usb_off_all() {
	fmt.Printf("%s Выключаем питание USB (хабы 1-4)...", time.Now().Format("2006-01-02 15:00:00"))
	for i := 0; i < 4; i++ {
		fmt.Printf("Отключаю хаб %d", i)
		cmd := exec.Command("uhubctl", "-l", fmt.Sprintf("%d", i), "-a", "0")
		cmd.Stderr = io.Discard // 2>/dev/null

		if err := cmd.Run(); err != nil {
			fmt.Printf(" (хаб %d не поддерживает отключение или уже выключен)\n", i)
		} else {
			fmt.Println("OK")
		}
	}
}

func usb_on_all() {
	fmt.Printf("%s Включаем питание USB (хабы 1-4)...", time.Now().Format("2006-01-02 15:00:00"))
	for i := 0; i < 4; i++ {
		fmt.Printf("Включаю хаб %d", i)
		cmd := exec.Command("uhubctl", "-l", fmt.Sprintf("%d", i), "-a", "1")
		cmd.Stderr = io.Discard // 2>/dev/null

		if err := cmd.Run(); err != nil {
			fmt.Printf(" (хаб %d не поддерживает отключение или уже выключен)\n", i)
		} else {
			fmt.Println("OK")
		}
	}
}

func main() {
	// Constants
	OFF_DURATION := 120 * time.Second
	CYCLE_INTERVAL_MIN := 180 * time.Minute

	// Calc
	waitAfterOn := CYCLE_INTERVAL_MIN - OFF_DURATION

	if waitAfterOn <= 0 {
		fmt.Println("Error")
	}

	fmt.Println("=== Управление питанием USB (Raspberry Pi 5) ===")
	fmt.Printf("Отключение на %.0f сек, повтор каждые %.0f мин\n", OFF_DURATION.Seconds(), CYCLE_INTERVAL_MIN.Minutes())
	fmt.Println("Остановить: Ctrl+C (USB будет включён обратно)")
	fmt.Println("================================================")

	// Check uhubctl
	checkUtil("uhubctl")

	// Check root
	checkRoot()

	// Signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main Loop Channel
	done := make(chan bool, 1)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nПолучен сигнал %v. Включаем USB...\n", sig)
		usb_on_all()
		fmt.Println("Скрипт завершён.")
		done <- true
	}()

	// Main Loop
	go func() {
		for {
			usb_off_all()
			time.Sleep(OFF_DURATION)

			usb_on_all()

			if waitAfterOn > 0 {
				fmt.Printf("Ожидание %.0f сек до следующего цикла...\n", waitAfterOn.Seconds())
				time.Sleep(waitAfterOn)
			}
		}
	}()

	// Block main thread while main signal is absent
	<-done
}
