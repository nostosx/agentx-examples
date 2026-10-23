module github.com/nostosx/agentx-examples

go 1.27.3

require (
	github.com/nostosx/agentx v0.0.0-00010101000000-000000000000
	github.com/nostosx/agentx-contrib/filestore v0.0.0-00010101000000-000000000000
	github.com/nostosx/agentx-contrib/provider v0.0.0-00010101000000-000000000000
)

require github.com/nostosx/kitx v1.0.0 // indirect

replace (
	github.com/nostosx/agentx => ../agentx
	github.com/nostosx/agentx-contrib/filestore => ../agentx-contrib/filestore
	github.com/nostosx/agentx-contrib/provider => ../agentx-contrib/provider
)
