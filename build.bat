@echo off
chcp 65001 >NUL 2>&1

REM Переходим в директорию где находится bat-файл
cd /d "%~dp0"

echo ═══════════════════════════════════════════════════════════════════════
echo                    КОМПИЛЯЦИЯ CUSTOS - GO VERSION
echo ═══════════════════════════════════════════════════════════════════════
echo.
echo Рабочая директория: %CD%
echo.

REM Проверка наличия Go
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go не установлен!
    echo.
    echo Скачайте Go с официального сайта:
    echo https://go.dev/dl/
    echo.
    echo После установки перезапустите командную строку.
    pause
    exit /b 1
)

echo [1/4] Проверка версии Go...
go version
echo.

REM Проверка наличия rsrc для встраивания манифеста
where rsrc >nul 2>&1
if %errorlevel% neq 0 (
    echo [2/4] Установка rsrc для встраивания манифеста...
    go install github.com/akavel/rsrc@latest
    echo.
) else (
    echo [2/4] rsrc уже установлен
    echo.
)

echo [3/4] Создание файла ресурсов с манифестом...
rsrc -manifest custos.manifest -o rsrc.syso
if %errorlevel% neq 0 (
    echo.
    echo [ERROR] Ошибка создания файла ресурсов!
    pause
    exit /b 1
)
echo.

echo [4/4] Компиляция custos.go...
echo.

REM Компиляция с флагами для уменьшения размера
REM -ldflags "-s -w" убирает отладочную информацию
REM rsrc.syso автоматически встроится с манифестом

go build -ldflags "-s -w" -o custos.exe custos.go

if %errorlevel% neq 0 (
    echo.
    echo [ERROR] Ошибка компиляции!
    pause
    exit /b 1
)

echo.
echo ═══════════════════════════════════════════════════════════════════════
echo                         ГОТОВО!
echo ═══════════════════════════════════════════════════════════════════════
echo.
echo Создан файл:
echo   • custos.exe - готовая программа с правами администратора
echo.
echo Для запуска:
echo   1. Дважды кликните на custos.exe
echo   2. Подтвердите запрос на права администратора (появится автоматически)
echo.
pause
