package syncer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"tmdb-sync-station/internal/catalog"
)

func convertDetails(details mediaDetails, kind string) (catalog.SyncTitle, error) {
	title, original, date := details.Title, details.OriginalTitle, details.ReleaseDate
	if kind == "tv" {
		title, original, date = details.Name, details.OriginalName, details.FirstAirDate
	}
	english := ""
	if details.OriginalLanguage == "en" {
		english = original
	}
	aliases := []string{title, original, english}
	for _, alternative := range append(details.AlternativeTitles.Titles, details.AlternativeTitles.Results...) {
		name := alternative.Title
		if name == "" {
			name = alternative.Name
		}
		aliases = append(aliases, name)
		if english == "" && (alternative.Language == "en" || alternative.Country == "US" || alternative.Country == "GB") {
			english = name
		}
	}
	genreIDs := make([]int, 0, len(details.Genres))
	for _, genre := range details.Genres {
		genreIDs = append(genreIDs, genre.ID)
	}
	countries := details.OriginCountry
	if kind == "movie" {
		countries = make([]string, 0, len(details.ProductionCountries))
		for _, country := range details.ProductionCountries {
			if country.Code != "" {
				countries = append(countries, country.Code)
			}
		}
	}
	directors := make([]string, 0)
	if kind == "tv" {
		for _, creator := range details.CreatedBy {
			if creator.Name != "" {
				directors = append(directors, creator.Name)
			}
		}
	} else {
		for _, member := range details.Credits.Crew {
			if member.Job == "Director" && member.Name != "" {
				directors = append(directors, member.Name)
			}
		}
	}
	cast := details.Credits.Cast
	if len(cast) > 20 {
		cast = cast[:20]
	}
	companies, err := json.Marshal(details.ProductionCompanies)
	if err != nil {
		return catalog.SyncTitle{}, fmt.Errorf("编码制作公司失败: %w", err)
	}
	networks, err := json.Marshal(details.Networks)
	if err != nil {
		return catalog.SyncTitle{}, fmt.Errorf("编码电视网失败: %w", err)
	}
	castJSON, err := json.Marshal(cast)
	if err != nil {
		return catalog.SyncTitle{}, fmt.Errorf("编码演职员失败: %w", err)
	}
	genreJSON, _ := json.Marshal(genreIDs)
	countryJSON, _ := json.Marshal(countries)
	year, _ := strconv.Atoi(firstFourDigits(date))
	seasonCount, episodeCount := 0, 0
	if kind == "tv" {
		seasonCount, episodeCount = details.NumberOfSeasons, details.NumberOfEpisodes
		if seasonCount == 0 {
			seasonCount = len(details.Seasons)
		}
	}
	var collectionID *int64
	collectionName := ""
	if details.BelongsToCollection != nil {
		collectionID, collectionName = &details.BelongsToCollection.ID, details.BelongsToCollection.Name
	}
	return catalog.SyncTitle{
		ID: details.ID, Kind: kind, Title: title, OriginalTitle: original, EnglishTitle: english, Overview: details.Overview,
		OriginalLanguage: details.OriginalLanguage, Year: year, Runtime: details.Runtime, VoteCount: details.VoteCount,
		SeasonCount: seasonCount, EpisodeCount: episodeCount, Popularity: details.Popularity, VoteAverage: details.VoteAverage,
		GenreIDs: genreJSON, OriginCountries: countryJSON, CastMembers: castJSON, Networks: networks, Companies: companies,
		Director: strings.Join(directors, ", "), PosterPath: details.PosterPath, BackdropPath: details.BackdropPath,
		IMDbID: details.ExternalIDs.IMDbID, Tagline: details.Tagline, Status: details.Status, Homepage: details.Homepage,
		LastAirDate: details.LastAirDate, InProduction: details.InProduction, ShowType: details.Type,
		CollectionID: collectionID, CollectionName: collectionName, Aliases: aliases,
	}, nil
}

func firstFourDigits(value string) string {
	if len(value) < 4 {
		return ""
	}
	return value[:4]
}
