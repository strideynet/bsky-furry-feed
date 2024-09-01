export function addSISuffix(number?: number) {
  number = number || 0;

  const suffixes = ["", "K", "M"];
  const order = Math.floor(Math.log10(number) / 3);

  for (let i = 0; i < order; i++) {
    number = number / 1000;
  }

  return `${Math.round(number * 100) / 100}${suffixes[order] || ""}`;
}
