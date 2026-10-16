import { getTmdbApiUrl } from "./config/tmdb";

async function request(path, options = {}) {
  const response = await fetch(path, options);
  const body = await response.json();
  if (!response.ok) throw new Error(body.error || `请求失败 (${response.status})`);
  return body;
}

export const loadDashboard = () => request("/api/dashboard");
export const loadCatalog = ({ type, q, limit = 25, offset = 0 }) => {
  const params = new URLSearchParams({ type, q, limit, offset });
  return request(`/api/catalog?${params}`);
};
export const loadCatalogDetail = ({ kind, id }) => {
  const params = new URLSearchParams({ kind, id: String(id) });
  return request(`/api/catalog/detail?${params}`);
};

const tmdbSeasonCache = new Map();
const tmdbSeasonRequests = new Map();

export function loadTmdbSeasonDetails({ showId, seasonNumber }) {
  const cacheKey = `${showId}-${seasonNumber}`;
  if (tmdbSeasonCache.has(cacheKey)) return Promise.resolve(tmdbSeasonCache.get(cacheKey));
  if (tmdbSeasonRequests.has(cacheKey)) return tmdbSeasonRequests.get(cacheKey);

  const params = new URLSearchParams({ language: "zh-CN" });
  const requestPromise = fetch(getTmdbApiUrl(`/3/tv/${showId}/season/${seasonNumber}?${params}`))
    .then(async (response) => {
      let body;
      try {
        body = await response.json();
      } catch {
        throw new Error("TMDB 代理返回了无法识别的响应");
      }
      if (!response.ok) throw new Error(body?.status_message || `TMDB 请求失败 (${response.status})`);
      if (!Array.isArray(body?.episodes)) throw new Error("TMDB 返回中没有该季的分集数据");
      const episodes = body.episodes.map((episode) => ({
        episodeNumber: Number(episode.episode_number || 0),
        title: episode.name || "",
        overview: episode.overview || "",
        airDate: episode.air_date || "",
        runtime: Number(episode.runtime || 0),
        voteAverage: Number(episode.vote_average || 0),
        stillPath: episode.still_path || "",
      }));
      tmdbSeasonCache.set(cacheKey, episodes);
      return episodes;
    })
    .finally(() => tmdbSeasonRequests.delete(cacheKey));
  tmdbSeasonRequests.set(cacheKey, requestPromise);
  return requestPromise;
}

export const loadDatabaseHealth = () => request("/api/health");
export const startSync = () => request("/api/sync/start", { method: "POST" });
export const loadSyncStatus = () => request("/api/sync/status");
