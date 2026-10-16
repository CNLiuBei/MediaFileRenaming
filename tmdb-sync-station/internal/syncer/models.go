package syncer

type changedIDsPage struct {
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
	Results    []struct {
		ID int64 `json:"id"`
	} `json:"results"`
}

type mediaDetails struct {
	ID               int64   `json:"id"`
	Title            string  `json:"title"`
	Name             string  `json:"name"`
	OriginalTitle    string  `json:"original_title"`
	OriginalName     string  `json:"original_name"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	FirstAirDate     string  `json:"first_air_date"`
	LastAirDate      string  `json:"last_air_date"`
	OriginalLanguage string  `json:"original_language"`
	Popularity       float64 `json:"popularity"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	Runtime          int     `json:"runtime"`
	Genres           []struct {
		ID int `json:"id"`
	} `json:"genres"`
	OriginCountry       []string `json:"origin_country"`
	ProductionCountries []struct {
		Code string `json:"iso_3166_1"`
	} `json:"production_countries"`
	ProductionCompanies []namedEntity `json:"production_companies"`
	Networks            []namedEntity `json:"networks"`
	Credits             credits       `json:"credits"`
	CreatedBy           []namedEntity `json:"created_by"`
	ExternalIDs         struct {
		IMDbID string `json:"imdb_id"`
	} `json:"external_ids"`
	AlternativeTitles struct {
		Titles  []localizedTitle `json:"titles"`
		Results []localizedTitle `json:"results"`
	} `json:"alternative_titles"`
	BelongsToCollection *struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"belongs_to_collection"`
	Tagline          string          `json:"tagline"`
	Status           string          `json:"status"`
	Homepage         string          `json:"homepage"`
	InProduction     bool            `json:"in_production"`
	Type             string          `json:"type"`
	NumberOfSeasons  int             `json:"number_of_seasons"`
	NumberOfEpisodes int             `json:"number_of_episodes"`
	Seasons          []seasonSummary `json:"seasons"`
}

type namedEntity struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type localizedTitle struct {
	Title    string `json:"title"`
	Name     string `json:"name"`
	Country  string `json:"iso_3166_1"`
	Language string `json:"iso_639_1"`
}
type credits struct {
	Cast []castCredit `json:"cast"`
	Crew []crewCredit `json:"crew"`
}
type castCredit struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	Order       int    `json:"order"`
	ProfilePath string `json:"profile_path"`
}
type crewCredit struct {
	Job  string `json:"job"`
	Name string `json:"name"`
}
type seasonSummary struct {
	Number       int    `json:"season_number"`
	Name         string `json:"name"`
	EpisodeCount int    `json:"episode_count"`
}
type seasonDetails struct {
	Name     string `json:"name"`
	Episodes []struct {
		Number int    `json:"episode_number"`
		Name   string `json:"name"`
	} `json:"episodes"`
}
