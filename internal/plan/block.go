package plan

type Block struct {
	Worker     string
	Repo       string
	OwnedPaths []string
	Change     string
	Verify     string
}
