# golang_match
Matching players into teams


## Testing
```
git clone https://github.com/starnuik/golang_match.git
cd golang_match
docker-compose -f compose.test.yaml pull
docker-compose -f compose.test.yaml up (-d)
go test ./... -v -count 1 -p 1
go test -benchmem -run ^\$ -bench Benchmark ./...
```

## Deployment
```
git clone https://github.com/starnuik/golang_match.git
cd golang_match
mv example.env .env
docker-compose -f compose.full.yaml pull
docker-compose -f compose.full.yaml build
docker-compose -f compose.full.yaml up (-d)

# sends a lot of requests automatically
USERS_PER_SECOND="10" ENDPOINT_URL="http://localhost/api/users" go run ./cmd/fake_load/
```

## Design
A user is represented as a 2d point with skill and latency being the points coordinates.
This 2d space is split into a grid, with each cell representing a small range of skill and latency.
When a user is added to the queue, they are put into one of the cells of the grid.
The basic matching algorithm walks over every cell and returns groups that are of the match size.
The priority algorithm searches in a square around the iterated cell for users that have been in the waiting queue longer than a specified limit, merges these priority users with the current cell, sorts them by descending wait time, then returns groups based on the same principle as in the basic algorithm.

## Дизайн
Пользователь представлен в виде точки в двумерной системе координат, с осями skill и latency.
Данная система координат разбита на клетки. Каждая клетка является представлением небольшого промежутка значений skill и latency.
При добавлении пользователя в очередь, он помещается в одну из корзин.
Базовый алгоритм обходит все корзины и возвращает из них группы размером match size.
Алгоритм с приоритетом при обходе корзин также ищет в увеличенном радиусе пользователей, время ожидания которых превысило определенный soft limut, и добавляет их к корзине базового алгоритма, предварительно отсортировав по убыванию времени ожидания.
