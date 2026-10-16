const defaultImageProxy = "https://tmdb.liubei.org";
const imageProxy = (import.meta.env.VITE_TMDB_IMAGE_PROXY || defaultImageProxy).replace(/\/+$/, "");

export function getPosterUrl(posterPath, size = "w342") {
  if (!posterPath) return "";
  const normalizedPath = posterPath
    .split("/")
    .filter(Boolean)
    .map(encodeURIComponent)
    .join("/");

  return normalizedPath ? `${imageProxy}/t/p/${size}/${normalizedPath}` : "";
}

export function getTmdbApiUrl(path) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${imageProxy}${normalizedPath}`;
}
