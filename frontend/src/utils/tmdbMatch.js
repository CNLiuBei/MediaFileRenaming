export function inferCategory(result, kind) {
  if (kind === "电影") {
    return ({ zh: "华语电影", en: "英语电影", ja: "日语电影", ko: "韩语电影" })[
      result.original_language
    ] || "其他语言电影";
  }
  const countries = result.origin_country || [];
  if (countries.includes("CN")) return "国产剧";
  if (countries.includes("US")) return "美剧";
  if (countries.includes("JP")) return "日剧";
  if (countries.includes("KR")) return "韩剧";
  if (countries.includes("HK")) return "港剧";
  if (countries.includes("TW")) return "台剧";
  if (countries.includes("GB")) return "英剧";
  return "其他地区剧";
}

function normalizeTitle(value) {
  return (value || "").normalize("NFKC").toLocaleLowerCase().replace(/[^\p{L}\p{N}]/gu, "");
}

function similarity(left, right) {
  if (!left || !right) return 0;
  if (left === right) return 1;
  if (left.includes(right) || right.includes(left)) return Math.min(left.length, right.length) / Math.max(left.length, right.length);
  if (left.length < 2 || right.length < 2) return 0;
  const leftPairs = new Set(Array.from({ length: left.length - 1 }, (_, index) => left.slice(index, index + 2)));
  const rightPairs = new Set(Array.from({ length: right.length - 1 }, (_, index) => right.slice(index, index + 2)));
  const shared = [...leftPairs].filter((pair) => rightPairs.has(pair)).length;
  return (2 * shared) / (leftPairs.size + rightPairs.size);
}

export function findBestTmdbMatch(results, query, year = "") {
  const normalizedQuery = normalizeTitle(query);
  if (!normalizedQuery) return null;
  let best = null;
  let bestScore = 0;
  for (const result of results) {
    if (year && result.year && result.year !== year) continue;
    const titleScore = Math.max(
      similarity(normalizedQuery, normalizeTitle(result.title)),
      similarity(normalizedQuery, normalizeTitle(result.original_title)),
    );
    const score = titleScore + (year && result.year === year ? 0.1 : 0);
    if (score > bestScore) {
      best = result;
      bestScore = score;
    }
  }
  const minimumScore = normalizedQuery.length < 5 ? 1 : 0.72;
  return bestScore >= minimumScore ? best : null;
}
