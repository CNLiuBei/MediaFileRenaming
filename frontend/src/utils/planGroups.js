export function groupPlanItems(items) {
  const groups = new Map();
  for (const item of items) {
    const normalizedTitle = (item.title || "")
      .trim()
      .replace(/\s+/g, " ")
      .toLocaleLowerCase();
    const isSeries = item.kind === "电视剧";
    const key =
      isSeries && normalizedTitle
        ? `series:${normalizedTitle}`
        : `file:${item.source}`;
    if (!groups.has(key)) {
      groups.set(key, {
        key,
        isSeries,
        title: item.title || "未识别片名",
        items: [],
      });
    }
    groups.get(key).items.push(item);
  }
  return [...groups.values()];
}
