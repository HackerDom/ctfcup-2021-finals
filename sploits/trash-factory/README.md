# Первая уязвимость

Обратим внимание на фрагмент кода из файла [web/main.go](../../services/trash-factory/cmd/web/main.go). 
```go
var (
	sessionsStorage = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	client          = NewClient(os.Getenv("CP_ADDR"), "", "")
	pageSize        = 20
)
```
Можно заметить, что сессии инициализируются с помощью переменной окружения `SESSION_KEY`, 
однако в файле [docker-compose.yaml](../../services/trash-factory/docker-compose.yml), в поле переменных окружения, переменная выставлена с опечаткой - `SESSIONS_KEY`, а значит сессии будут проинициализированы пустой строкой. Появляется возможность создавать собственные сессии, которые позволят получить `TOKEN` любого пользователя по `TOKEN_KEY`.

Детали реализации создания сессии по `TOKEN_KEY` можно найти в [эксплойте](1_session_generation/cookie_gen.go).

# Вторая и третья уязвимости
Данные уязвимости несут в себе тривиальную идею атаки на встроенный ГПСЧ go, 
точнее на тот seed, которым он инициализируется, и различаются лишь вектором атаки.

Для второй уязвимости некорректную инициализацию можно найти в файле [web/main.go](../../services/trash-factory/cmd/web/main.go).
```go
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/stat", statHandler)
	rand.Seed(time.Now().Unix()) # <-- SEED INIT
	fmt.Printf("Starting server at port :%s\n", port)
```
А используется в файле [pkg/api/client.go](../../services/trash-factory/pkg/api/client.go).
```go
func (client *Client) CreateUser() (string, error) {
	msg := []byte{commands.CreateUser}

	token := make([]byte, 8)
	binary.LittleEndian.PutUint64(token, rand.Uint64())
	tokenKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(tokenKey, rand.Uint64())

	createUserOp := commands.CreateUserOp{
		Token:    token,
		TokenKey: hex.EncodeToString(tokenKey),
	}
```
Таким образом если подобрать seed, который инициализируется от текущего timestamp в секундах, 
то появляется возможность подобрать `TOKEN` пользователя.

Для третьей уязвимости обратимся к файлу [controlpanel/main.go](../../services/trash-factory/cmd/controlpanel/main.go).
```go
	cp.stats = statistic
	rand.Seed(time.Now().Unix()) <-- Init random
	CreateAdminUser(err, &cp)

	return &cp
}

func CreateAdminUser(err error, cp *ControlPanel) {
	adminTokenKey := fmt.Sprintf("%08x", rand.Uint64()) <-- Usage
	adminToken := fmt.Sprintf("%08x", rand.Uint64())    <-- Usage
	cp.AdminCredentials = &models.User{
		TokenKey:      adminTokenKey,
		Token:         []byte(adminToken),
		ContainersIds: []string{},
	}
```
В следствии некорректной инициализации, появляется возможность подделать токен администратора, 
и возможность получения TOKEN`