module process

go 1.13

replace github.com/rogerluo410/openai-ws/src/client => ../client

require (
	github.com/panjf2000/ants/v2 v2.4.8
	github.com/rogerluo410/openai-ws/src/client v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
)
