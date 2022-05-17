/**
 *
 * @param delay  mini seconds
 * @returns
 */
export const sleep = async (delay: number) => {
	return new Promise(resolve => {
		setTimeout(resolve, delay);
	});
};


/**
 * It takes two arrays and a function, and calls the function on every combination of the two arrays.
 * @param {T[]} arr1 - The first array to loop through
 * @param {T[]} arr2 - The array that will be looped through.
 * @param func - (a: T, b: T) => void
 */
export const doubleArrayLoopCall = <T> (arr1: T[], arr2:T[], func: (a: T, b: T) => void) => {
	for (let i = 0; i < arr1.length; i++) {
		for (let k = 0; k < arr2.length; k++) {
			func(arr1[i], arr2[k]);
		}
	}
}


/**
 * It takes a URL as a string and returns the root of the URL as a string.
 * @param {string} url - The URL you want to get the root of.
 * @returns The url root of the url passed in.
 */
export const urlRoot = (url: string) => {
	const urlParts = url.split("/");
	return urlParts[0] + "//" + urlParts[2];
}
