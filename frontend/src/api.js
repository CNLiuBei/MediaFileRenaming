async function request(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: {
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...options.headers,
    },
  });
  const data = await response.json();
  if (!response.ok)
    throw new Error(data.error || `请求失败（${response.status}）`);
  return data;
}
export const getStatus = () => request("/api/status");
export const getLogs = () => request("/api/logs?limit=100");
export const browseDirectories = (path = "") =>
  request(`/api/directories?path=${encodeURIComponent(path)}`);
export const saveConfig = (config) =>
  request("/api/config", { method: "PUT", body: JSON.stringify(config) });
export const previewPlan = (overrides) =>
  request("/api/preview", {
    method: "POST",
    body: JSON.stringify({ overrides }),
  });
export const runPlan = (overrides) =>
  request("/api/run", { method: "POST", body: JSON.stringify({ overrides }) });
export const searchTmdb = (payload) =>
  request("/api/tmdb/search", {
    method: "POST",
    body: JSON.stringify(payload),
  });
