/**
 * From div percent width compute fixed left value, for center in window.
 * @param percent
 * @returns
 */
export function computeFixedLeftDistanceFromPercent(percent: number): string {
	const windoWidth = window.innerWidth;

	const divWidth = windoWidth * (percent / 100);
	return `${(windoWidth - divWidth) / 2}px`;
}


if (import.meta.vitest) {

	const { it, expect, assert } = import.meta.vitest;


}
