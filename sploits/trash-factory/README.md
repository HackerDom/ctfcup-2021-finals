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
и возможность получения `TOKEN`

# Четвертая уязвимость

Посмотрим на то, как сервис хранит контейнеры, для этого заглянем в [controlpanel/db.go](../../services/trash-factory/cmd/controlpanel/db.go).

```go
	db := DataBase{dbPath: "db/"}
	db.userDBPath = db.dbPath + "users/"
	db.containerDBPath = db.dbPath + "containers/"
```

```go
func (db *DataBase) SaveContainer(tokenKey string, container *models.Container) error {
	data, err := container.Serialize()
	if err != nil {
		return err
	}

	userFolder := db.containerDBPath + tokenKey + "/"
	err = os.MkdirAll(userFolder, os.ModePerm)
	if err != nil {
		return err
	}

	if err := os.WriteFile(userFolder+container.ID, data, 0666); err != nil {
		return err
	}

	user, err := db.GetUser(tokenKey)
	if err != nil {
		return err
	}
	user.ContainersIds = append(user.ContainersIds, container.ID)
	return db.SaveUser(user)
}
```

Контейнеры находятся по следующему пути: `db/containers/{{ userTokenKey }}/{{ containerID}}`, причем `{{containerID}}` это инкрементарный счетчик
[controlpanel/db.go](../../services/trash-factory/cmd/commands/commands.go)
```go
func (cp *ControlPanel) CreateContainer(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializeCreateContainerOp(opBytes)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't deserialize container. %s", err))
	}

	fmt.Printf("Container creating... TOKEN %02x  ARGS %02x\n", tokenKey, opBytes)

	if op.Size > '\x05' || op.Size < '\x01' {
		return nil, errors.New("incorrect container size")
	}

	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()
	count, err := cp.DB.GetContainersCount(tokenKey)
	if err != nil {
		return nil, err
	}
	container := models.Container{
		ID:          strconv.Itoa(count + 1),
		Size:        op.Size,
		Description: op.Description,
	}
	return []byte(container.ID), cp.DB.SaveContainer(tokenKey, &container)
}
```

Если в методе `GetContainerInfo` передать `ContainerID` вида `db/containers/{{ userTokenKey }}/{{ containerID}/../{{ victimTokenKey }}/{{ containerID}}` 