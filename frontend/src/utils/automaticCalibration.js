import { searchTmdb } from "../api";
import { findBestTmdbMatch, inferCategory } from "./tmdbMatch";

function getTargetKey(group) {
  const item = group.items[0];
  return `${item.kind}:${(item.title || "").trim().toLocaleLowerCase()}:${item.year || ""}`;
}

export function countCalibrationTargets(groups) {
  return new Set(groups.filter((group) => group.items[0]?.title).map(getTargetKey)).size;
}

export function buildCalibrationOverrides(selections) {
  const overrides = {};
  for (const { group, result } of selections) {
    const kind = group.items[0].kind;
    const metadata = {
      kind,
      category: inferCategory(result, kind),
      title: result.title,
      year: result.year || "",
      tmdb_id: result.id,
      poster_path: result.poster_path || "",
    };
    for (const item of group.items) {
      overrides[item.source] = {
        ...metadata,
        ...(kind === "电视剧"
          ? { season: item.season || 1, episode: item.episode || 1 }
          : {}),
      };
    }
  }
  return overrides;
}

export async function calibrateGroups(groups, onProgress, shouldStop) {
  const candidatesByKey = new Map();
  const selections = [];
  let attempted = 0;
  let unmatched = 0;
  let error = "";
  let cancelled = false;
  const targets = groups.filter((group) => group.items[0]?.title);

  for (const [index, group] of targets.entries()) {
    if (shouldStop()) {
      cancelled = true;
      break;
    }
    const item = group.items[0];
    const targetKey = getTargetKey(group);
    onProgress(`正在校准 ${index + 1}/${targets.length}：${group.title}`);
    if (!candidatesByKey.has(targetKey)) {
      attempted += 1;
      try {
        const response = await searchTmdb({
          source: item.source,
          kind: item.kind,
          query: item.title,
          year: item.year || "",
        });
        candidatesByKey.set(targetKey, response.results);
      } catch (cause) {
        error = cause.message;
        break;
      }
    }
    const result = findBestTmdbMatch(candidatesByKey.get(targetKey), item.title, item.year || "");
    if (result) selections.push({ group, result });
    else unmatched += 1;
  }
  if (!cancelled && shouldStop()) cancelled = true;
  return { selections, attempted, unmatched, error, cancelled };
}
