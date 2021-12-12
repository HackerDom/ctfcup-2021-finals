# Нулевая уязвимость

В файле [docker-compose.yml](../../services/5g_shop/docker-compose.yml) в секции инициализации образа postgresql разработчик сервиса, 
видимо, забыл после очередной отладки сервиса, что открывал порты. Надо либо убирать строчку про порты (у контейнеров общая сеть), либо менять пароль ролей в бд.
```yml
    ports:
      - "5432:5432"
```

# Первая уязвимость

Бэкэнд сервиса в качестве базы данных использует postgresql. Так как С++ относительно низкоуровневый язык, то сервис не пользуется никакими ORM, 
но всю работу с базой проделывает через сырые запросы с помощью библиотеки [libpq](https://www.postgresql.org/docs/current/libpq.html). Следовательно,
есть поле для SQL инъекций. При этом почти все места, куда можно было бы вставить инъекцию защищены функцией экранирования PQescapeLiteral. Кроме одного места - 
[метода, используемого при авторизации пользователя](../../services/5g_shop/back/internal/src/services/UsersService.cpp#L98):
```cpp
    auto query = Format(
            HiddenStr("select * from users where login='%s' and password_hash=%s;"),
            login.c_str(),
            escapedPasswordHash.c_str()
    );

    result = PQexec(conn, query.c_str());
```

Сюда можно вставлять практически любой SQL запрос, однако на выходе есть [проверка на "количество и качество"](../../services/5g_shop/back/internal/src/services/UsersService.cpp#L105), которая сильно затрудняет 
извлечение данных через простые запросы:
```cpp
    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        auto msg = std::string(PQresultErrorMessage(result));

        CROW_LOG_ERROR << msg;

        return Result<std::shared_ptr<User>>::ofError("invalid login or password");
    }

    auto user = User::ReadFromPGResult(result);

    if (user->passwordHash != passwordHash || user->login != login) {
        return Result<std::shared_ptr<User>>::ofError("invalid login or password");
    }
```

В то же время, в сервисе есть функционал загрузки картинок, использующихся для "описания" продаваемой техники. Логика проверки того, что загружаемый файл является картинкой, [присутствует](../../services/5g_shop/back/internal/src/main.cpp#L422), однако написана с ошибкой - проверка есть только заголовка и то, только в случае, когда он есть:
```cpp
    for (auto &header: part.headers) {
        if (header.value.first == "Content-Type") {
            if (!header.value.second.starts_with("image")) {
                return response(status::BAD_REQUEST);
            }
        }
        if (header.value.first == "Content-Disposition") {
            filename = header.params["filename"];
        }
    }
```

Поэтому загрузить любой файл не составит труда. К тому же загруженные файлы будут доступны из контейнера postgresql. Поэтому есть вариант использования уязвимости - выполнить произвольный код
в postgresql, нужно только загрузить туда правильный *.so*-файл. Пример реализации можно посмотреть [здесь](./psql-rce)


# Вторая уязвимость

В сервисе используется самописный генератор uuid4 для кук авторизации пользователей. Генератор основан на линейном конгруэнтном методе, который легко взламывается. [UUID.cpp](services/5g_shop/back/internal/src/tools/UUID.cpp#L14)
```cpp
static volatile uint32_t x, a = 1103515245, c = 1013904223;
static bool initialized = false;

static char alpha[] = "0123456789abcdef";

namespace shop {
    uint16_t rand() {
        if (!initialized) {
            initialized = true;

            x = std::time(nullptr);
        }

        x = x * a + c;

        return static_cast<uint16_t>((x & 0xFFFF0000) >> 16);
    }
}
```

Более того, получаемые uuid почти полностью выдают стейт генератора, лишнь меняя полубайты местами и выводя сгенерированные числа "в bit endian'e":
```cpp
int counters[] = {4, 2, 2, 8};

    union {
        uint16_t randvalues[8];
        uint8_t bytes[16];
    };

    randvalues[0] = rand();
    randvalues[1] = rand();
    randvalues[2] = rand();
    randvalues[3] = rand();
    randvalues[4] = rand();
    randvalues[5] = rand();
    randvalues[6] = rand();
    randvalues[7] = rand();

    int p = 0;

    for (auto i = 0; i < 4; ++i) {
        for (int j = 0; j < counters[i]; ++j) {
            uint8_t b = bytes[p++];

            ss << alpha[b & 0x0F] << alpha[(b & 0xF0) >> 4];
        }

        if (i != 3) {
            ss << "-";
        }
    }

    return ss.str();
```

