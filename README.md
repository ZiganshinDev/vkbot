# scheduleVKBot Golang

Чат-Бот, написанный на Golang, для ВКонтакте. Имеет функционал, предоставляющий расписание занятий для студентов НИУ МГСУ.

Создан как проект, чтобы научиться работать с SQL (использована база данных SQLite. Запросы написаны на чистом SQL без использования ORM), также разобраться в CI/CD процессах, GitHubActions и хостингах. 

## Подробнее: 

### 1. Причины создания бота:
·Университет не предоставляет удобный способ получения обновленного расписания занятий. Таким образом, бот может стать альтернативным и удобным способом получения расписания занятий для студентов, которым требуется более быстрый и удобный доступ к этой информации.

·Студентам часто бывает трудно запомнить свое расписание или не могут найти его в нужный момент. Бот может стать удобным и быстрым способом для получения расписания занятий, без необходимости запоминать его или искать на других ресурсах, помимо мессенджера.

·Наличие бота может значительно упростить и улучшить жизнь студентов. Это может повысить удобство использования и повысить охват пользователей, что приведет к увеличению числа пользователей и возможному расширению функционала вашего бота в будущем.

### 2. Функциональность бота:
·Чат-Бот работает на платформе ВКонтакте

1)Пользователь отправляет запрос в бот с указанием номера группы и дня недели, на который он хочет получить расписание.
2)Бот обрабатывает запрос и выдает пользователю расписание занятий на указанный день.

### 3. Процесс разработки бота и интересные моменты разработки:
·Регистрация и настройка сообщества бота на платформе ВКонтакте.

·Создание определенного алгоритма обработки запросов пользователя. Написание программного кода для обработки запросов от пользователей.
Для сохраняния чистой архитектуры проекта и следованию принципам SOLID, алгоритм разработки был: 

1) Инициализация конфига
2) Инициализация логгирования
3) Описание структуры базы данных и её методов
4) Инициализация базы данных (в данном случае SQLite)
5) Описание интерфейса базы данных
6) Описание поведения Чат-Бота
7) Инициализация Чат-Бота

·Кодовая база разделена по модулям (пакетам) в соответсивии golang-standards/project-layout

·Определение структуры данных, используемых для хранения информации о расписании занятий.
Была использована база данных SQLite для хранения данных пользователей и расписания занятий, поскольку она легковесная и с ней не появляется необходимость в Docker-контейнерах и их оркестрации.
База данных имеет 2 схемы: schedule с полями lesson_id INTEGER PRIMARY KEY, institute TEXT, course INTEGER, group_number INTEGER, lesson_name TEXT, lesson_type TEXT, date_range TEXT, day TEXT, audience TEXT, lesson_number INTEGER, week TEX  и users с полями user_id INTEGER PRIMARY KEY, institute TEXT, course INTEGER, group_number INTEGER, peer_id INTEGER, week TEXT.

Получение расписания для конкретного пользователя через JOIN, с сравнением полей схем и peer_id пользователя ВКонтакте.

## Примеры кода:

### main.go файл - отправная точка:
![Imgur](https://i.imgur.com/AO5RJzk.png)

### Реализация интерфейса базы данных в месте использования:
![Imgur](https://i.imgur.com/VLbx7Kl.png)

### Часть конфига для размещения (деплоя):
![Imgur](https://i.imgur.com/UYzKt5I.png)

### Некоторые из методов базы данных:
![Imgur](https://i.imgur.com/QooZJeq.png)

![Imgur](https://i.imgur.com/vVyhBxB.png)

### 4. Трудности в процессе разработки:
·Следование SOLID-принципам и принципам чистой архитектуры, для уменьшение технического долга проекта.

·Установка необходимых инстументов. Напимер GCC и G++ компиляторы на Windows, для запуска инструментов вне UNIX-like систем. 

·Внесение данных в базу из неудобных для парсинга форматов(PDF).

·Выбор хостинга и дистрибутива для приложения (приложение было поднято на Ubuntu 22.04 из-за glibc-библиотеки).

·Деплой на удаленный хостинг, связанный с малым опытом в разработке.

### 5. Результаты:
·Размещение и тестирование Чат-Бота на платформе ВКонтакте.

·Улучшение коммуникации между студентами и университетом. Бот может стать дополнительным каналом связи между студентами и университетом, что может помочь улучшить коммуникацию и облегчить процесс получения информации.

·Удобство использования бота и возможность быстрого доступа к расписанию занятий. 
Чат-Бот может предоставлять студентам расписание, не выходя из приложения ВКонтакте, что может увеличить удобство получения информации.

### 6. Возможные улучшения:
·Уменьшение технического долга, рефакторинг.

·Добавление новых функций, например, напоминаний о ближайших занятиях или уведомлений об изменениях в расписании.

·Работа с другими форматами начальных данных.

·Написание API для неавтризированных пользователей.

# Процесс деплоя через GitHubActions:

![Imgur](https://i.imgur.com/dmoynIf.png)

# Демонстрация:

![Imgur](https://i.imgur.com/dQkfkRT.png)

# Проект с открытым исходным кодом

·https://github.com/ZiganshinDev/vkbot