export interface MetricSample {
  name: string;
  labels: Record<string, string>;
  value: number;
}

export function parsePrometheusText(text: string): MetricSample[] {
  const samples: MetricSample[] = [];

  for (const line of text.split('\n')) {
    if (!line || line.startsWith('#')) continue;

    const match = line.match(/^([a-zA-Z_:][a-zA-Z0-9_:]*)(\{([^}]*)\})?\s+([0-9.eE+-]+)$/);
    if (!match) continue;

    const [, name, , labelStr, valueStr] = match;
    const labels: Record<string, string> = {};

    if (labelStr) {
      for (const pair of labelStr.matchAll(/([a-zA-Z_][a-zA-Z0-9_]*)="([^"]*)"/g)) {
        labels[pair[1]] = pair[2];
      }
    }

    samples.push({ name, labels, value: parseFloat(valueStr) });
  }

  return samples;
}