package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	shell32              = syscall.NewLazyDLL("shell32.dll")
	user32               = syscall.NewLazyDLL("user32.dll")
	procShellExecute     = shell32.NewProc("ShellExecuteW")
	procSetConsoleMode   = kernel32.NewProc("SetConsoleMode")
	procGetConsoleMode   = kernel32.NewProc("GetConsoleMode")
	procGetSystemMenu    = user32.NewProc("GetSystemMenu")
	procDeleteMenu       = user32.NewProc("DeleteMenu")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procSetWindowLong    = user32.NewProc("SetWindowLongW")
	procGetWindowLong    = user32.NewProc("GetWindowLongW")
	procSetConsoleCtrlHandler = kernel32.NewProc("SetConsoleCtrlHandler")
)

// Глобальный список процессов для отслеживания
var (
	runningProcesses = make(map[*exec.Cmd]bool)
	processMutex     sync.Mutex
)

// Цвета для консоли
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorBold    = "\033[1m"
	ColorMagenta = "\033[35m"
)

// Константы для настройки окна консоли
const (
	SC_CLOSE       = 0xF060
	SC_MINIMIZE    = 0xF020
	SC_MAXIMIZE    = 0xF030
	SC_SIZE        = 0xF000
	MF_BYCOMMAND   = 0x00000000
	WS_SIZEBOX     = 0x00040000
	WS_MAXIMIZEBOX = 0x00010000
)

// Добавление процесса в список отслеживания
func trackProcess(cmd *exec.Cmd) {
	processMutex.Lock()
	runningProcesses[cmd] = true
	processMutex.Unlock()
}

// Удаление процесса из списка отслеживания
func untrackProcess(cmd *exec.Cmd) {
	processMutex.Lock()
	delete(runningProcesses, cmd)
	processMutex.Unlock()
}

// Завершение всех запущенных процессов
func killAllProcesses() {
	processMutex.Lock()
	defer processMutex.Unlock()

	for cmd := range runningProcesses {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
	runningProcesses = make(map[*exec.Cmd]bool)
}

// Обработчик закрытия консоли
func setupCloseHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()
}

// Функция очистки ресурсов
func cleanup() {
	killAllProcesses()
}

// Проверка прав администратора
func isAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}

// Запуск с правами администратора
func runAsAdmin() {
	if isAdmin() {
		return
	}

	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)

	var showCmd int32 = 1 // SW_NORMAL

	ret, _, _ := procShellExecute.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		0,
		uintptr(unsafe.Pointer(cwdPtr)),
		uintptr(showCmd),
	)

	if ret > 32 {
		os.Exit(0)
	}
}

// Настройка консоли (фиксированный размер, отключение изменения размера)
func setupConsole() {
	// Получаем хэндл окна консоли
	consoleWindow, _, _ := procGetConsoleWindow.Call()

	if consoleWindow != 0 {
		// Получаем системное меню
		systemMenu, _, _ := procGetSystemMenu.Call(consoleWindow, 0)

		if systemMenu != 0 {
			// Удаляем кнопку "Развернуть" и возможность изменения размера
			procDeleteMenu.Call(systemMenu, SC_MAXIMIZE, MF_BYCOMMAND)
			procDeleteMenu.Call(systemMenu, SC_SIZE, MF_BYCOMMAND)
		}

		// Получаем текущий стиль окна
		gwlStyle := ^uintptr(0) - 16 + 1 // -16 в uintptr
		style, _, _ := procGetWindowLong.Call(consoleWindow, gwlStyle)

		// Убираем флаги WS_SIZEBOX и WS_MAXIMIZEBOX
		newStyle := style &^ uintptr(WS_SIZEBOX) &^ uintptr(WS_MAXIMIZEBOX)

		// Устанавливаем новый стиль
		procSetWindowLong.Call(consoleWindow, gwlStyle, newStyle)
	}

	// Настраиваем режим консоли для поддержки ANSI цветов
	handle := uintptr(syscall.Stdout)
	var mode uint32
	procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
	procSetConsoleMode.Call(handle, uintptr(mode))
}

// Очистка экрана
func clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// Заголовок
func printHeader() {
	clearScreen()
	fmt.Println(ColorCyan + ColorBold)
	fmt.Println()
	fmt.Println("                  █▀▀ █ █ █▀ ▀█▀ █▀█ █▀")
	fmt.Println("                  █▄▄ █▄█ ▄█  █  █▄█ ▄█")
	fmt.Println()
	fmt.Println("                   sdelano s lubovyu")
	fmt.Println()
	fmt.Println(ColorReset)

	if isAdmin() {
		fmt.Printf("  %s[✓]%s Статус: %sАдминистратор%s\n", ColorGreen, ColorReset, ColorGreen+ColorBold, ColorReset)
	} else {
		fmt.Printf("  %s[✗]%s Статус: %sОтсутствуют права администратора!%s\n", ColorRed, ColorReset, ColorRed+ColorBold, ColorReset)
	}
	fmt.Printf("  %s[i]%s Дата: %s\n", ColorBlue, ColorReset, time.Now().Format("02.01.2006 15:04:05"))
	fmt.Println()
}

// Меню
func printMenu(title string, options []string, showBack bool) {
	fmt.Printf("\n%s%s%s\n\n", ColorYellow+ColorBold, title, ColorReset)

	for i, opt := range options {
		fmt.Printf("  %s[%d]%s ➤ %s\n", ColorCyan+ColorBold, i+1, ColorReset, opt)
	}

	if showBack {
		fmt.Printf("\n  %s[0]%s ← %sНазад%s\n", ColorMagenta+ColorBold, ColorReset, ColorMagenta, ColorReset)
	} else {
		fmt.Printf("\n  %s[0]%s ✖ %sВыход%s\n", ColorRed+ColorBold, ColorReset, ColorRed, ColorReset)
	}
	fmt.Println()
}

// Получение выбора
func getChoice(maxOpt int) int {
	for {
		fmt.Printf("\n%s►%s Выберите опцию [0-%d]: ", ColorGreen+ColorBold, ColorReset, maxOpt)
		var choice int
		_, err := fmt.Scanf("%d\n", &choice)
		if err == nil && choice >= 0 && choice <= maxOpt {
			return choice
		}
		fmt.Printf("\n%s⚠ Ошибка: Введите число от 0 до %d%s\n", ColorRed+ColorBold, maxOpt, ColorReset)
		var discard string
		fmt.Scanln(&discard)
	}
}

// Логирование
func log(msg string, ok bool) {
	t := time.Now().Format("15:04:05")
	if ok {
		fmt.Printf("  %s[%s]%s %s✓%s %s\n", ColorBlue, t, ColorReset, ColorGreen+ColorBold, ColorReset, msg)
	} else {
		fmt.Printf("  %s[%s]%s %s✗%s %s\n", ColorBlue, t, ColorReset, ColorRed+ColorBold, ColorReset, msg)
	}
}

// Открытие папки
func openFolder(path, desc string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log(fmt.Sprintf("Папка не найдена: %s", path), false)
		return false
	}

	cmd := exec.Command("explorer", path)
	err := cmd.Start()
	if err != nil {
		log(fmt.Sprintf("Ошибка: %v", err), false)
		return false
	}

	trackProcess(cmd)
	go func() {
		cmd.Wait()
		untrackProcess(cmd)
	}()

	log(fmt.Sprintf("%s: %s", desc, path), true)
	return true
}

// Выполнение команды
func runCommand(command, desc string) bool {
	cmd := exec.Command("cmd", "/c", "start", command)
	err := cmd.Start()
	if err != nil {
		log(fmt.Sprintf("Ошибка: %v", err), false)
		return false
	}

	trackProcess(cmd)
	go func() {
		cmd.Wait()
		untrackProcess(cmd)
	}()

	log(desc, true)
	return true
}

// Открытие реестра с копированием пути
func openRegistry(path string) bool {
	// Копируем путь в буфер обмена
	cmd := exec.Command("cmd", "/c", fmt.Sprintf("echo %s | clip", path))
	err := cmd.Run()
	if err != nil {
		log(fmt.Sprintf("Ошибка копирования: %v", err), false)
		return false
	}

	// Открываем regedit
	cmdReg := exec.Command("regedit.exe")
	err = cmdReg.Start()
	if err != nil {
		log(fmt.Sprintf("Ошибка: %v", err), false)
		return false
	}

	trackProcess(cmdReg)
	go func() {
		cmdReg.Wait()
		untrackProcess(cmdReg)
	}()

	log(fmt.Sprintf("Путь скопирован: %s", path), true)
	fmt.Printf("%sВставьте путь в regedit (Ctrl+V)%s\n", ColorYellow, ColorReset)
	return true
}

// Пауза
func pause() {
	fmt.Printf("\n%s►%s Нажмите Enter для продолжения...\n", ColorGreen+ColorBold, ColorReset)
	fmt.Scanln()
}

// МЕНЮ РАЗДЕЛЫ

func networkMenu() {
	printHeader()
	fmt.Printf("\n%s═══ СЕТЬ И ИНТЕРНЕТ ═══%s\n\n", ColorCyan+ColorBold, ColorReset)

	runCommand("ms-settings:datausage", "Использование данных")

	fmt.Printf("\n%sЧТО НУЖНО ПРОВЕРИТЬ:%s\n", ColorYellow+ColorBold, ColorReset)
	fmt.Printf("  %s►%s Неизвестные .exe файлы с сетевой активностью\n", ColorRed, ColorReset)
	fmt.Printf("  %s►%s Подозрительные названия процессов\n", ColorRed, ColorReset)
	fmt.Printf("  %s►%s Большой объем переданных данных\n", ColorRed, ColorReset)
	pause()
}

func defenderMenu() {
	printHeader()
	fmt.Printf("\n%s═══ ЗАЩИТА WINDOWS ═══%s\n\n", ColorCyan+ColorBold, ColorReset)

	runCommand("windowsdefender://threat/", "Журнал защиты Windows Defender")

	fmt.Printf("\n%sКЛЮЧЕВЫЕ СЛОВА ДЛЯ ПОИСКА:%s\n", ColorYellow+ColorBold, ColorReset)
	fmt.Printf("  %s►%s undead, melony, ancient, loader\n", ColorRed, ColorReset)
	fmt.Printf("  %s►%s hack, cheat, unturned, bypass\n", ColorRed, ColorReset)
	fmt.Printf("  %s►%s inject, overlay, esp, aimbot\n", ColorRed, ColorReset)
	pause()
}

func foldersMenu() {
	for {
		printHeader()
		printMenu("СИСТЕМНЫЕ ПАПКИ", []string{
			"AppData\\Roaming",
			"AppData\\Local",
			"AppData\\LocalLow",
			"Videos (видео)",
			"Prefetch (запущенные .exe)",
			"Открыть все",
		}, true)

		choice := getChoice(6)
		if choice == 0 {
			break
		}

		appdata := os.Getenv("APPDATA")
		localappdata := os.Getenv("LOCALAPPDATA")
		userprofile := os.Getenv("USERPROFILE")

		switch choice {
		case 1:
			openFolder(appdata, "AppData\\Roaming")
			pause()
		case 2:
			openFolder(localappdata, "AppData\\Local")
			pause()
		case 3:
			openFolder(filepath.Join(userprofile, "AppData", "LocalLow"), "AppData\\LocalLow")
			pause()
		case 4:
			openFolder(filepath.Join(userprofile, "Videos"), "Videos")
			pause()
		case 5:
			openFolder("C:\\Windows\\Prefetch", "Prefetch")
			pause()
		case 6:
			openFolder(appdata, "Roaming")
			openFolder(localappdata, "Local")
			openFolder(filepath.Join(userprofile, "AppData", "LocalLow"), "LocalLow")
			openFolder(filepath.Join(userprofile, "Videos"), "Videos")
			openFolder("C:\\Windows\\Prefetch", "Prefetch")
			pause()
		}
	}
}

func registryMenu() {
	for {
		printHeader()
		printMenu("РЕЕСТР WINDOWS", []string{
			"Открыть regedit",
			"MuiCache (запущенные программы)",
			"AppSwitched (переключения Alt+Tab)",
			"ShowJumpView (JumpList история)",
		}, true)

		choice := getChoice(4)
		if choice == 0 {
			break
		}

		switch choice {
		case 1:
			cmd := exec.Command("regedit.exe")
			err := cmd.Start()
			if err == nil {
				trackProcess(cmd)
				go func() {
					cmd.Wait()
					untrackProcess(cmd)
				}()
				log("Regedit открыт", true)
			} else {
				log(fmt.Sprintf("Ошибка: %v", err), false)
			}
			pause()
		case 2:
			openRegistry(`HKEY_CURRENT_USER\SOFTWARE\Classes\Local Settings\Software\Microsoft\Windows\Shell\MuiCache`)
			pause()
		case 3:
			openRegistry(`HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Explorer\FeatureUsage\AppSwitched`)
			pause()
		case 4:
			openRegistry(`HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Explorer\FeatureUsage\ShowJumpView`)
			pause()
		}
	}
}

func utilitiesMenu() {
	printHeader()
	fmt.Printf("\n%s═══ УТИЛИТЫ ═══%s\n\n", ColorCyan+ColorBold, ColorReset)

	fmt.Printf("  %s[i]%s Открываем ссылки на утилиты для проверки...%s\n\n", ColorBlue, ColorReset, ColorReset)

	runCommand("https://www.voidtools.com/downloads/", "Everything (поиск файлов)")
	runCommand("https://www.nirsoft.net/utils/computer_activity_view.html", "ComputerActivityView")
	runCommand("https://www.nirsoft.net/utils/usb_devices_view.html", "USBDevicesView")
	runCommand("https://privazer.com/en/download-shellbag-analyzer-shellbag-cleaner.php", "ShellBag Analyzer")

	fmt.Printf("\n%sУТИЛИТЫ:%s\n", ColorYellow+ColorBold, ColorReset)
	fmt.Printf("  %s►%s Everything - быстрый поиск файлов на ПК\n", ColorCyan, ColorReset)
	fmt.Printf("  %s►%s ComputerActivityView - активность компьютера\n", ColorCyan, ColorReset)
	fmt.Printf("  %s►%s USBDevicesView - история USB устройств\n", ColorCyan, ColorReset)
	fmt.Printf("  %s►%s ShellBag Analyzer - анализ посещенных папок\n", ColorCyan, ColorReset)
	pause()
}

func steamCheckMenu() {
	printHeader()
	fmt.Printf("\n%s═══ ПРОВЕРКА STEAM АККАУНТА ═══%s\n\n", ColorCyan+ColorBold, ColorReset)

	// Ищем папку Steam в стандартных местах
	possiblePaths := []string{
		`C:\Program Files (x86)\Steam\config`,
		`C:\Program Files\Steam\config`,
	}

	// Также проверяем в других дисках
	drives := []string{"D:", "E:", "F:"}
	for _, drive := range drives {
		possiblePaths = append(possiblePaths, filepath.Join(drive, "Steam", "config"))
		possiblePaths = append(possiblePaths, filepath.Join(drive, "Program Files (x86)", "Steam", "config"))
		possiblePaths = append(possiblePaths, filepath.Join(drive, "Program Files", "Steam", "config"))
	}

	found := false
	for _, steamPath := range possiblePaths {
		if _, err := os.Stat(steamPath); !os.IsNotExist(err) {
			found = true
			if openFolder(steamPath, "Папка Steam\\config") {
				fmt.Printf("\n%sЧТО НУЖНО ПРОВЕРИТЬ:%s\n", ColorYellow+ColorBold, ColorReset)
				fmt.Printf("  %s►%s Конфигурационные файлы Steam\n", ColorRed, ColorReset)
				fmt.Printf("  %s►%s Информация об аккаунтах\n", ColorRed, ColorReset)
				fmt.Printf("  %s►%s Логи и настройки\n", ColorRed, ColorReset)
			}
			break
		}
	}

	if !found {
		log("Папка Steam\\config не найдена в системе", false)
		fmt.Printf("\n%s⚠ Steam может быть не установлен или находится в нестандартной директории%s\n", ColorYellow, ColorReset)
	}

	pause()
}

func unturnedMenu() {
	printHeader()
	fmt.Printf("\n%s═══ UNTURNED ═══%s\n\n", ColorCyan+ColorBold, ColorReset)

	// Ищем папку Unturned в стандартных местах Steam
	possiblePaths := []string{
		`C:\Program Files (x86)\Steam\steamapps\common\Unturned\Screenshots`,
		`C:\Program Files\Steam\steamapps\common\Unturned\Screenshots`,
	}

	// Также проверяем в других дисках
	drives := []string{"D:", "E:", "F:"}
	for _, drive := range drives {
		possiblePaths = append(possiblePaths, filepath.Join(drive, "Steam", "steamapps", "common", "Unturned", "Screenshots"))
		possiblePaths = append(possiblePaths, filepath.Join(drive, "Program Files (x86)", "Steam", "steamapps", "common", "Unturned", "Screenshots"))
		possiblePaths = append(possiblePaths, filepath.Join(drive, "Program Files", "Steam", "steamapps", "common", "Unturned", "Screenshots"))
	}

	found := false
	for _, screenshots := range possiblePaths {
		if _, err := os.Stat(screenshots); !os.IsNotExist(err) {
			found = true
			fmt.Printf("  %s[i]%s Найдено: %s%s%s\n\n", ColorBlue, ColorReset, ColorCyan, screenshots, ColorReset)
			if openFolder(screenshots, "Папка Screenshots Unturned") {
				fmt.Printf("\n%sЧТО НУЖНО ПРОВЕРИТЬ:%s\n", ColorYellow+ColorBold, ColorReset)
				fmt.Printf("  %s►%s UI читов на скриншотах\n", ColorRed, ColorReset)
				fmt.Printf("  %s►%s ESP/Wallhack индикаторы\n", ColorRed, ColorReset)
				fmt.Printf("  %s►%s Overlay меню\n", ColorRed, ColorReset)
				fmt.Printf("  %s►%s Необычные элементы интерфейса\n", ColorRed, ColorReset)
			}
			break
		}
	}

	if !found {
		log("Папка Steam\\steamapps\\common\\Unturned\\Screenshots не найдена в системе", false)
		fmt.Printf("\n%s⚠ Unturned может быть не установлен или находится в нестандартной директории%s\n", ColorYellow, ColorReset)
	}

	pause()
}

func mainMenu() {
	for {
		printHeader()
		printMenu("ГЛАВНОЕ МЕНЮ", []string{
			"Сеть и интернет",
			"Защита Windows",
			"Утилиты",
			"Системные папки",
			"Реестр Windows",
			"Проверка Steam аккаунта",
			"Unturned",
			"Справка (ключевые слова)",
		}, false)

		choice := getChoice(8)

		switch choice {
		case 0:
			fmt.Printf("\n%s✖ Завершение работы...%s\n", ColorCyan+ColorBold, ColorReset)
			time.Sleep(500 * time.Millisecond)
			cleanup()
			return
		case 1:
			networkMenu()
		case 2:
			defenderMenu()
		case 3:
			utilitiesMenu()
		case 4:
			foldersMenu()
		case 5:
			registryMenu()
		case 6:
			steamCheckMenu()
		case 7:
			unturnedMenu()
		case 8:
			printHeader()
			fmt.Printf("\n%s═══ КЛЮЧЕВЫЕ СЛОВА ДЛЯ ПОИСКА ═══%s\n\n", ColorYellow+ColorBold, ColorReset)
			fmt.Printf("  %s►%s undead, melony, ancient, loader, hack, cheat, unturned\n", ColorCyan, ColorReset)
			fmt.Printf("  %s►%s чит, loader, soft, bypass, inject, overlay\n", ColorCyan, ColorReset)
			fmt.Printf("  %s►%s esp, aimbot, wallhack, speedhack, fly\n\n", ColorCyan, ColorReset)
			fmt.Printf("%sПроверяйте файлы, реестр и папки на наличие этих слов%s\n", ColorYellow, ColorReset)
			pause()
		}
	}
}

func main() {
	// Проверка и запрос прав администратора
	runAsAdmin()

	// Настройка обработчика закрытия консоли
	setupCloseHandler()

	// Настройка консоли (цвета, фиксированный размер)
	setupConsole()

	// Запуск главного меню
	mainMenu()

	// Очистка при нормальном выходе
	cleanup()
}
