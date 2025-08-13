package repository

type Repositories struct {
	User       UserRepository
	Credential CredentialRepository
	// Add future repos: CardRepository, TextRepository, etc.
}
