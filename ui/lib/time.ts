import { Interval } from "./api/core/types.pb";

export function showInterval(interval: Interval): string {
  const parts = [];

  if (interval.hours !== "0") {
    parts.push(`${interval.hours}h`);
  }

  if (interval.minutes !== "0" || parts.length > 0) {
    // Show minutes we have Hour, but minute is zero.
    // For example: 1h 0m 23s
    // It's easier to read, without it, it's easy to misread the
    // value "1h 23s" as "1m 23s"
    parts.push(`${interval.minutes}m`);
  }

  if (interval.seconds !== "0" || parts.length > 0) {
    // Same as minute.
    parts.push(`${interval.seconds}s`);
  }

  return parts.join(" ");
}
