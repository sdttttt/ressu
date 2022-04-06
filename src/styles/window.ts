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

// @ts-ignore
if (import.meta.vitest) {
	// @ts-ignore
	const { it, expect, assert } = import.meta.vitest;

	it("Math.sqrt()", () => {
		expect(Math.sqrt(4)).toBe(2);
		expect(Math.sqrt(144)).toBe(12);
		expect(Math.sqrt(2)).toBe(Math.SQRT2);
	});
	it("JSON", () => {
		const input = {
			foo: "hello",
			bar: "world"
		};

		const output = JSON.stringify(input);

		expect(output).eq('{"foo":"hello","bar":"world"}');
		assert.deepEqual(JSON.parse(output), input, "matches original");
	});
}
