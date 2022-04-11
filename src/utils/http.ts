import { getClient, Response, ResponseType } from "@tauri-apps/api/http";

/**
 * request client.
 */
const client = getClient();

/**
 * FROM url get rss context text.
 * @param url 
 * @returns 
 */
export async function fetchRSSText(url: string): Promise<string> {
    const response: Response<string> = await (await client).get(url, { responseType: ResponseType.Text });
    return response.data;
}