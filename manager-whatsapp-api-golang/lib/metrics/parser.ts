/**
 * Prometheus Exposition Format Parser
 *
 * Parses raw Prometheus metrics text format into structured TypeScript objects.
 * Handles counters, gauges, histograms, and summaries.
 *
 * @example
 * ```ts
 * const raw = await fetch('/metrics').then(r => r.text());
 * const parsed = parsePrometheusMetrics(raw);
 * console.log(parsed.families['http_requests_total']);
 * ```
 *
 * @module lib/metrics/parser
 */

import type {
	HistogramBucket,
	HistogramMetric,
	MetricFamily,
	MetricSample,
	MetricType,
	ParsedMetrics,
} from "@/types/metrics";

/**
 * Regex patterns for parsing Prometheus format
 */
const PATTERNS = {
	/** Matches # HELP metric_name description */
	help: /^#\s*HELP\s+(\w+)\s+(.*)$/,
	/** Matches # TYPE metric_name type */
	type: /^#\s*TYPE\s+(\w+)\s+(\w+)$/,
	/** Matches metric_name{label="value"} value timestamp? */
	sample: /^(\w+)(?:{([^}]*)})?(?:\s+)([+-]?(?:\d+\.?\d*|\d*\.?\d+)(?:[eE][+-]?\d+)?|[+-]?Inf|NaN)(?:\s+(\d+))?$/,
	/** Matches label="value" pairs */
	label: /(\w+)="([^"]*)"/g,
};

/**
 * Parse a single line of Prometheus format
 */
function parseLine(line: string): {
	type: "help" | "type" | "sample" | "empty" | "unknown";
	data?: Record<string, unknown>;
} {
	const trimmed = line.trim();

	if (!trimmed || trimmed.startsWith("##")) {
		return { type: "empty" };
	}

	// Parse HELP comment
	const helpMatch = trimmed.match(PATTERNS.help);
	if (helpMatch) {
		return {
			type: "help",
			data: { name: helpMatch[1], help: helpMatch[2] },
		};
	}

	// Parse TYPE comment
	const typeMatch = trimmed.match(PATTERNS.type);
	if (typeMatch) {
		return {
			type: "type",
			data: { name: typeMatch[1], metricType: typeMatch[2] as MetricType },
		};
	}

	// Parse metric sample
	const sampleMatch = trimmed.match(PATTERNS.sample);
	if (sampleMatch) {
		const [, name, labelsStr, valueStr, timestampStr] = sampleMatch;

		// Parse labels
		const labels: Record<string, string> = {};
		if (labelsStr) {
			let match: RegExpExecArray | null;
			const labelPattern = new RegExp(PATTERNS.label.source, "g");
			while ((match = labelPattern.exec(labelsStr)) !== null) {
				labels[match[1]] = match[2];
			}
		}

		// Parse value
		let value: number;
		if (valueStr === "+Inf" || valueStr === "Inf") {
			value = Number.POSITIVE_INFINITY;
		} else if (valueStr === "-Inf") {
			value = Number.NEGATIVE_INFINITY;
		} else if (valueStr === "NaN") {
			value = Number.NaN;
		} else {
			value = Number.parseFloat(valueStr);
		}

		return {
			type: "sample",
			data: {
				name,
				labels,
				value,
				timestamp: timestampStr ? Number.parseInt(timestampStr, 10) : undefined,
			},
		};
	}

	return { type: "unknown" };
}

/**
 * Get the base metric name without histogram/summary suffixes
 */
function getBaseMetricName(name: string): string {
	// Remove histogram suffixes
	if (name.endsWith("_bucket")) {
		return name.slice(0, -7);
	}
	if (name.endsWith("_sum")) {
		return name.slice(0, -4);
	}
	if (name.endsWith("_count")) {
		return name.slice(0, -6);
	}
	if (name.endsWith("_total")) {
		return name.slice(0, -6);
	}
	return name;
}

/**
 * Compute percentile from histogram buckets using linear interpolation
 */
function computePercentile(
	buckets: HistogramBucket[],
	totalCount: number,
	percentile: number,
): number | undefined {
	if (totalCount === 0 || buckets.length === 0) {
		return undefined;
	}

	const targetCount = totalCount * (percentile / 100);

	// Sort buckets by boundary (le)
	const sortedBuckets = [...buckets].sort((a, b) => {
		const aVal = a.le === "+Inf" ? Number.POSITIVE_INFINITY : Number.parseFloat(a.le);
		const bVal = b.le === "+Inf" ? Number.POSITIVE_INFINITY : Number.parseFloat(b.le);
		return aVal - bVal;
	});

	let prevCount = 0;
	let prevBoundary = 0;

	for (const bucket of sortedBuckets) {
		const boundary =
			bucket.le === "+Inf"
				? Number.POSITIVE_INFINITY
				: Number.parseFloat(bucket.le);

		if (bucket.count >= targetCount) {
			// Linear interpolation
			if (bucket.count === prevCount) {
				return boundary;
			}

			const fraction = (targetCount - prevCount) / (bucket.count - prevCount);
			return prevBoundary + fraction * (boundary - prevBoundary);
		}

		prevCount = bucket.count;
		prevBoundary = boundary;
	}

	// If we get here, return the last non-infinite boundary
	const lastFiniteBucket = sortedBuckets.findLast(
		(b) => b.le !== "+Inf" && Number.isFinite(Number.parseFloat(b.le)),
	);
	return lastFiniteBucket
		? Number.parseFloat(lastFiniteBucket.le)
		: undefined;
}

/**
 * Parse Prometheus exposition format into structured metrics
 *
 * @param text - Raw Prometheus metrics text
 * @returns Parsed metrics with families and samples
 */
export function parsePrometheusMetrics(text: string): ParsedMetrics {
	const lines = text.split("\n");
	const families: Record<string, MetricFamily> = {};
	const parseErrors: string[] = [];

	let currentMetric: {
		name: string;
		help: string;
		type: MetricType;
	} | null = null;

	for (const line of lines) {
		const result = parseLine(line);

		switch (result.type) {
			case "help": {
				const { name, help } = result.data as { name: string; help: string };
				if (!families[name]) {
					families[name] = {
						name,
						help: help || "",
						type: "unknown",
						samples: [],
					};
				} else {
					families[name].help = help || "";
				}
				currentMetric = {
					name,
					help: help || "",
					type: families[name].type,
				};
				break;
			}

			case "type": {
				const { name, metricType } = result.data as {
					name: string;
					metricType: MetricType;
				};
				if (!families[name]) {
					families[name] = {
						name,
						help: "",
						type: metricType,
						samples: [],
					};
				} else {
					families[name].type = metricType;
				}
				if (currentMetric?.name === name) {
					currentMetric.type = metricType;
				}
				break;
			}

			case "sample": {
				const sample = result.data as unknown as MetricSample;
				const baseName = getBaseMetricName(sample.name);

				// Ensure family exists
				if (!families[baseName]) {
					families[baseName] = {
						name: baseName,
						help: "",
						type: "unknown",
						samples: [],
					};
				}

				families[baseName].samples.push(sample);
				break;
			}

			case "unknown": {
				if (line.trim() && !line.trim().startsWith("#")) {
					parseErrors.push(`Failed to parse line: ${line.substring(0, 100)}`);
				}
				break;
			}
		}
	}

	return {
		families,
		timestamp: new Date().toISOString(),
		parseErrors,
	};
}

/**
 * Extract histogram data from parsed metrics
 *
 * @param family - Metric family containing histogram samples
 * @returns Array of histogram metrics with computed percentiles
 */
export function extractHistograms(family: MetricFamily): HistogramMetric[] {
	if (family.type !== "histogram") {
		return [];
	}

	// Group samples by label set (excluding 'le')
	const groups = new Map<string, { buckets: HistogramBucket[]; sum: number; count: number; labels: Record<string, string> }>();

	for (const sample of family.samples) {
		// Create a key from labels excluding 'le'
		const labelsCopy = { ...sample.labels };
		const le = labelsCopy.le;
		delete labelsCopy.le;
		const key = JSON.stringify(labelsCopy);

		if (!groups.has(key)) {
			groups.set(key, { buckets: [], sum: 0, count: 0, labels: labelsCopy });
		}

		const group = groups.get(key)!;

		if (sample.name.endsWith("_bucket") && le !== undefined) {
			group.buckets.push({ le, count: sample.value });
		} else if (sample.name.endsWith("_sum")) {
			group.sum = sample.value;
		} else if (sample.name.endsWith("_count")) {
			group.count = sample.value;
		}
	}

	// Convert to histogram metrics
	const histograms: HistogramMetric[] = [];

	for (const [, group] of groups) {
		const histogram: HistogramMetric = {
			name: family.name,
			labels: group.labels,
			buckets: group.buckets,
			sum: group.sum,
			count: group.count,
		};

		// Compute percentiles
		if (group.count > 0) {
			histogram.p50 = computePercentile(group.buckets, group.count, 50);
			histogram.p90 = computePercentile(group.buckets, group.count, 90);
			histogram.p95 = computePercentile(group.buckets, group.count, 95);
			histogram.p99 = computePercentile(group.buckets, group.count, 99);
		}

		histograms.push(histogram);
	}

	return histograms;
}

/**
 * Get samples filtered by labels
 *
 * @param family - Metric family to filter
 * @param labelFilter - Label key-value pairs to match
 * @returns Filtered samples
 */
export function filterSamplesByLabels(
	family: MetricFamily,
	labelFilter: Record<string, string>,
): MetricSample[] {
	return family.samples.filter((sample) => {
		for (const [key, value] of Object.entries(labelFilter)) {
			if (sample.labels[key] !== value) {
				return false;
			}
		}
		return true;
	});
}

/**
 * Sum all values in a metric family
 *
 * @param family - Metric family
 * @param suffix - Optional suffix to filter (e.g., '_total')
 * @returns Sum of all matching sample values
 */
export function sumMetricValues(
	family: MetricFamily,
	suffix?: string,
): number {
	return family.samples
		.filter((s) => !suffix || s.name.endsWith(suffix))
		.reduce((sum, s) => sum + (Number.isFinite(s.value) ? s.value : 0), 0);
}

/**
 * Get gauge value (single value metric)
 *
 * @param family - Metric family
 * @returns The gauge value or 0 if not found
 */
export function getGaugeValue(family: MetricFamily): number {
	const sample = family.samples.find(
		(s) => s.name === family.name && Object.keys(s.labels).length === 0,
	);
	return sample?.value ?? 0;
}

/**
 * Group samples by a specific label
 *
 * @param family - Metric family
 * @param labelKey - Label key to group by
 * @returns Map of label values to samples
 */
export function groupSamplesByLabel(
	family: MetricFamily,
	labelKey: string,
): Map<string, MetricSample[]> {
	const groups = new Map<string, MetricSample[]>();

	for (const sample of family.samples) {
		const labelValue = sample.labels[labelKey] || "_unknown_";
		if (!groups.has(labelValue)) {
			groups.set(labelValue, []);
		}
		groups.get(labelValue)!.push(sample);
	}

	return groups;
}

/**
 * Extract unique label values from a metric family
 *
 * @param family - Metric family
 * @param labelKey - Label key to extract
 * @returns Array of unique label values
 */
export function getUniqueLabelValues(
	family: MetricFamily,
	labelKey: string,
): string[] {
	const values = new Set<string>();

	for (const sample of family.samples) {
		if (sample.labels[labelKey]) {
			values.add(sample.labels[labelKey]);
		}
	}

	return Array.from(values);
}
