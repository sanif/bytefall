package data

// StatsProvider is the interface for domain statistics sources
type StatsProvider interface {
	GetStats() []*DomainStats
}
