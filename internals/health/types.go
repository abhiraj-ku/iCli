package health

type ServerStats struct {
	PGVersion    string
	DatabaseName string
	DatabaseSize string
	Uptime       string
	InRecovery   bool
}

type ConnectionStats struct {
	TotalConnections int
	MaxConnections   int
	Active           int
}

type CacheStats struct {
	HitRatio  float64
	MissRatio float64
}

type Report struct {
	Server      ServerStats
	Connections ConnectionStats
	Cache       CacheStats
}
