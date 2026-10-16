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
export function importOfflineCatalog(file, onProgress) {
  return new Promise((resolve, reject) => {
    const upload = new XMLHttpRequest();
    upload.open("POST", "/api/catalog/import");
    upload.setRequestHeader("Content-Type", "application/octet-stream");
    upload.upload.onprogress = (event) => {
      if (event.lengthComputable) onProgress(event.loaded / event.total);
    };
    upload.onerror = () => reject(new Error("上传连接中断，请检查 NAS 网络后重试"));
    upload.onload = () => {
      let result;
      try {
        result = JSON.parse(upload.responseText);
      } catch {
        reject(new Error(`服务器返回无效响应（${upload.status}）`));
        return;
      }
      if (upload.status < 200 || upload.status >= 300) {
        reject(new Error(result.error || `导入失败（${upload.status}）`));
        return;
      }
      resolve(result.catalog);
    };
    upload.send(file);
  });
}
