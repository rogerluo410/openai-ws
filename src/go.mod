module github.com/rogerluo410/openai-ws

go 1.13

replace github.com/rogerluo410/openai-ws/src/client => ./client

replace github.com/rogerluo410/openai-ws/src/process => ./process

require (
	github.com/rogerluo410/openai-ws/src/client v0.0.0-00010101000000-000000000000
	github.com/rogerluo410/openai-ws/src/process v0.0.0-00010101000000-000000000000
)
