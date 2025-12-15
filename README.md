# Go Parser 27 Вариант

## Общее описание

Этот проект представляет собой инструмент командной строки, написанный на Go, для преобразования текста из учебного конфигурационного языка в формат TOML. Парсер обрабатывает различные конструкции языка, включая комментарии, числа, массивы, словари, строки и константные выражения, с проверкой синтаксических ошибок и выводом информативных сообщений об ошибках.

## Описание всех функций и настроек

### Поддерживаемые конструкции языка

#### 1. Комментарии
- Однострочные комментарии:
    -- Это однострочный комментарий
  
- Многострочные комментарии:
    =begin
  Это многострочный
  комментарий
  =cut
  

#### 2. Числа
Поддерживается формат чисел с плавающей точкой и экспоненциальной записью:
[+-]?\d+\.?\d*[eE][+-]?\d+

Примеры: 123, -45.67, 1.23e+10, -5.6E-3

#### 3. Массивы
#( значение значение значение ... )

Пример: #( 1 2 "три" 4.5 )

#### 4. Словари
begin
  имя := значение;
  имя := значение;
  имя := значение;
  ...
end


#### 5. Строки
'Это строка'


#### 6. Объявление констант
(define имя значение);


#### 7. Константные выражения
![имя + 1]
![sort(array_value)]
![max(a b c)]


### Поддерживаемые операции и функции

1. Сложение (+)
2. Функция сортировки (sort())
3. Функция максимума (max())

## Команды для сборки проекта и запуска тестов

### Сборка проекта

# Клонирование репозитория
git clone https://github.com/Ijne/GoParserForSomethingUse-less-ful.git
cd GoParserForSomethingUse-less-ful

# Сборка проекта
go build -o parser main.go

# Или сборка с оптимизацией
go build -ldflags="-s -w" -o parser main.go


### Запуск тестов

# Запуск всех тестов
go test ./...

# Запуск тестов с подробным выводом
go test -v ./...

# Запуск тестов с покрытием
go test -cover ./...

# Запуск конкретного тестового файла
go test -v parser_test.go


## Примеры использования

### Пример 1: Базовая конфигурация приложения

Входной файл (config.input):
-- Конфигурация веб-приложения
(define version 1.0);

begin
  app_name := 'MyWebApp';
  version := ![version + 0.1]; -- 1.1
  debug_mode := true;
  
  server := begin
    host := 'localhost';
    port := 8080;
    ssl_enabled := false;
  end;
  
  database := begin
    host := 'db.localhost';
    port := 5432;
    connections := 20;
  end;
  
  features := #( 'auth' 'api' 'logging' );
  timeout := 30.5;
end;


Команда для преобразования:
./parser -input config.input


### Пример 2: Конфигурация игрового движка

Входной файл (game.input):
=begin
Конфигурация игрового движка
Версия 2.1
=cut

(define base_speed 10);
(define player_count 4);

begin
  game_title := 'Space Adventure';
  version := '2.1.0';
  
  graphics := begin
    resolution := #( 1920 1080 );
    fullscreen := true;
    vsync := true;
    fps_limit := 60;
  end;
  
  physics := begin
    gravity := 9.81;
    collision_detection := true;
  end;
  
  players := begin
    count := ![player_count];
    default_speed := ![base_speed * 2]; -- 20
    abilities := #( 'jump' 'shoot' 'run' );
  end;
  
  levels := #( 'forest' 'cave' 'castle' 'boss' );
  sorted_levels := ![sort(levels)];
  
  scores := #( 1500 3200 980 2750 );
  high_score := ![max(scores)]; -- 3200
end;


Команда для преобразования:
./parser -input game.input > game_config.toml


### Использование инструмента

# Базовое использование
./parser -input input_file.txt

# Перенаправление вывода в файл
./parser -input config.input > output.toml

# Просмотр справки
./parser -help


### Ключи командной строки

- -input string: Путь к входному файлу (обязательный параметр)
- -help: Показать справку по использованию

Проект полностью покрыт тестами и обрабатывает все заявленные конструкции учебного конфигурационного языка, включая вложенные структуры и константные выражения.
