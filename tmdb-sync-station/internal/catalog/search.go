package catalog

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func (s *Store) SearchCatalog(ctx context.Context, kind, query, year string) ([]CatalogSearchResult, error) {
	dbKind := map[string]string{"电影": "movie", "电视剧": "tv"}[kind]
	key := catalogSearchKey(query)
	if dbKind == "" || key == "" || len([]rune(query)) > 120 {
		return nil, fmt.Errorf("影视类型或搜索词无效")
	}
	var yearValue int
	if len(year) == 4 {
		parsedYear, err := strconv.Atoi(year)
		if err == nil {
			yearValue = parsedYear
		}
	}
	rows, err := s.db.QueryContext(ctx, `WITH matched_ids AS MATERIALIZED (
		SELECT a.tmdb_id FROM aliases a WHERE a.kind=$1 AND
			(a.name_key=$2 OR (length($2)>=2 AND a.name_key LIKE $2 || '%'))
		UNION
		SELECT t.tmdb_id FROM titles t WHERE t.kind=$1 AND length($3)>=3 AND
			(t.title || ' ' || t.original_title || ' ' || t.english_title) ILIKE '%' || $3 || '%'
	)
	SELECT t.tmdb_id,t.title,t.original_title,t.english_title,
		COALESCE(alias.name,''),COALESCE(t.year::text,''),t.overview,t.original_language,
		t.origin_countries,t.poster_path
	FROM matched_ids m JOIN titles t ON t.tmdb_id=m.tmdb_id AND t.kind=$1
	LEFT JOIN LATERAL (
		SELECT a.name FROM aliases a WHERE a.tmdb_id=t.tmdb_id AND a.kind=t.kind
			AND (a.name_key=$2 OR (length($2)>=2 AND a.name_key LIKE $2 || '%'))
		ORDER BY CASE WHEN a.name_key=$2 THEN 0 WHEN a.name_key LIKE $2 || '%' THEN 1 ELSE 2 END,
			length(a.name_key),a.name LIMIT 1
	) alias ON true
	ORDER BY CASE WHEN t.year=$4 AND $4>0 THEN 0 ELSE 1 END,t.popularity DESC
	LIMIT 40`, dbKind, key, strings.TrimSpace(query), yearValue)
	if err != nil {
		return nil, fmt.Errorf("查询同步片库失败: %w", err)
	}
	defer rows.Close()
	results := make([]CatalogSearchResult, 0)
	for rows.Next() {
		var item CatalogSearchResult
		var originCountry []byte
		if err := rows.Scan(&item.ID, &item.Title, &item.OriginalTitle, &item.EnglishTitle,
			&item.MatchedAlias, &item.Year, &item.Overview, &item.OriginalLanguage,
			&originCountry, &item.PosterPath); err != nil {
			return nil, err
		}
		item.OriginCountry = jsonOrEmptyArray(originCountry)
		results = append(results, item)
	}
	return results, rows.Err()
}

func catalogSearchKey(value string) string {
	var key strings.Builder
	for _, char := range norm.NFKC.String(strings.ToLower(strings.TrimSpace(value))) {
		if unicode.IsPunct(char) || unicode.IsSpace(char) {
			continue
		}
		key.WriteRune(char)
	}
	return key.String()
}
