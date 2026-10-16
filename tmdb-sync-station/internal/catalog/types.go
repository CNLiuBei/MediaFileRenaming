package catalog

import "encoding/json"

type Dashboard struct {
	Catalog CatalogStats `json:"catalog"`
	Sync    SyncStatus   `json:"sync"`
	Recent  []Change     `json:"recent"`
	Daily   []DailyCount `json:"daily"`
	Runs    []SyncRun    `json:"runs"`
}

type CatalogStats struct {
	Titles          int64  `json:"titles"`
	Movies          int64  `json:"movies"`
	Series          int64  `json:"series"`
	Episodes        int64  `json:"episodes"`
	Aliases         int64  `json:"aliases"`
	Size            string `json:"size"`
	Source          string `json:"source"`
	SourceUpdatedAt string `json:"sourceUpdatedAt"`
	ImportedAt      string `json:"importedAt"`
}

type SyncStatus struct {
	State          string  `json:"state"`
	Configured     bool    `json:"configured"`
	AuthConfigured bool    `json:"authConfigured"`
	AutoSync       bool    `json:"autoSync"`
	APIBase        string  `json:"apiBase"`
	RateLimit      float64 `json:"rateLimit"`
	Progress       int     `json:"progress"`
	Current        int64   `json:"current"`
	Total          int64   `json:"total"`
	Stage          string  `json:"stage"`
	LastSync       string  `json:"lastSync"`
	NextSync       string  `json:"nextSync"`
	Description    string  `json:"description"`
}

type Change struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Kind      string `json:"kind"`
	Action    string `json:"action"`
	UpdatedAt string `json:"updatedAt"`
}

type DailyCount struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type SyncRun struct {
	ID         int64  `json:"id"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
	State      string `json:"state"`
	Processed  int64  `json:"processed"`
	Current    int64  `json:"current"`
	Total      int64  `json:"total"`
	Stage      string `json:"stage"`
	Error      string `json:"error"`
}

type CatalogItem struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Kind          string  `json:"kind"`
	OriginalTitle string  `json:"originalTitle"`
	Overview      string  `json:"overview"`
	ReleaseDate   string  `json:"releaseDate"`
	PosterPath    string  `json:"posterPath"`
	Popularity    float64 `json:"popularity"`
	UpdatedAt     string  `json:"updatedAt"`
}

type CatalogSearchResult struct {
	ID               int64           `json:"id"`
	Title            string          `json:"title"`
	OriginalTitle    string          `json:"original_title"`
	EnglishTitle     string          `json:"english_title"`
	MatchedAlias     string          `json:"matched_alias"`
	Year             string          `json:"year"`
	Overview         string          `json:"overview"`
	OriginalLanguage string          `json:"original_language"`
	OriginCountry    json.RawMessage `json:"origin_country"`
	PosterPath       string          `json:"poster_path"`
}

type CatalogDetail struct {
	ID               int64           `json:"id"`
	Kind             string          `json:"kind"`
	Title            string          `json:"title"`
	OriginalTitle    string          `json:"originalTitle"`
	Overview         string          `json:"overview"`
	Tagline          string          `json:"tagline"`
	Year             int             `json:"year"`
	OriginalLanguage string          `json:"originalLanguage"`
	OriginCountries  json.RawMessage `json:"originCountries"`
	PosterPath       string          `json:"posterPath"`
	BackdropPath     string          `json:"backdropPath"`
	Popularity       float64         `json:"popularity"`
	VoteAverage      float64         `json:"voteAverage"`
	VoteCount        int             `json:"voteCount"`
	Runtime          int             `json:"runtime"`
	Director         string          `json:"director"`
	CastMembers      json.RawMessage `json:"castMembers"`
	Networks         json.RawMessage `json:"networks"`
	Companies        json.RawMessage `json:"companies"`
	Status           string          `json:"status"`
	Homepage         string          `json:"homepage"`
	IMDbID           string          `json:"imdbId"`
	SeasonCount      int             `json:"seasonCount"`
	EpisodeCount     int             `json:"episodeCount"`
	Seasons          []SeasonDetail  `json:"seasons"`
	InProduction     bool            `json:"inProduction"`
	CollectionName   string          `json:"collectionName"`
}

type SeasonDetail struct {
	SeasonNumber int             `json:"seasonNumber"`
	Name         string          `json:"name"`
	EpisodeCount int             `json:"episodeCount"`
	Episodes     []EpisodeDetail `json:"episodes"`
}

type EpisodeDetail struct {
	EpisodeNumber int    `json:"episodeNumber"`
	Title         string `json:"title"`
}
