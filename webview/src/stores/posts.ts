import { toast } from "~/lib/notifications";
import { Posts, type SortOrder } from "~/mygo";
import { app as state, setAppState as setState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";
import { closeSnapshot } from "~/stores/snapshot-actions";

/** How many posts a page of the list holds. */
const PAGE_SIZE = 100;

/** How long to wait for typing to settle before searching titles. */
const SEARCH_DEBOUNCE_MS = 300;

const aggregateFeedId = 0;

let searchTimer: ReturnType<typeof setTimeout> | undefined;

/** Loads a page of the selected feed's posts, replacing or appending them. */
export async function loadPosts(offset = 0): Promise<void> {
  const feedId = state.current.feedId ?? aggregateFeedId;

  setState("loadingPosts", true);
  try {
    const page = await Posts.list({
      feedId,
      unreadOnly: state.unreadOnly,
      starredOnly: state.starredOnly,
      search: state.postSearch,
      offset,
      limit: PAGE_SIZE,
      sortOrder: state.postSortOrder,
    });
    if (offset === 0) setState("posts", page.posts);
    else setState("posts", (posts) => [...posts, ...page.posts]);
    setState({ postTotal: page.total, postUnread: page.unread });
  } catch (error) {
    toast.error(`无法读取文章：${messageOf(error)}`);
  } finally {
    setState("loadingPosts", false);
  }
}

/** Loads the next page of posts. */
export async function loadMorePosts(): Promise<void> {
  if (state.loadingPosts) return;
  await loadPosts(state.posts.length);
}

/** Moves the native article list selection from a page shortcut. */
export async function selectNeighbourPost(delta: 1 | -1, onlyUnread = false): Promise<void> {
  await Posts.selectNeighbour(delta, onlyUnread);
}

/** Searches post titles after the user stops typing. */
export function setPostSearch(value: string): void {
  clearTimeout(searchTimer);
  const query = value.trim();
  if (query === "") {
    setState("postSearch", "");
    void loadPosts(0);
    return;
  }

  searchTimer = setTimeout(() => {
    setState("postSearch", query);
    void loadPosts(0);
  }, SEARCH_DEBOUNCE_MS);
}

/** Shows the title search field, or clears the active search while hiding it. */
export async function togglePostSearch(): Promise<void> {
  if (!state.postSearchOpen) {
    setState("postSearchOpen", true);
    return;
  }

  clearTimeout(searchTimer);
  const hadSearch = state.postSearch !== "";
  setState({ postSearchOpen: false, postSearch: "" });
  if (hadSearch) await loadPosts(0);
}

/** Opens an article image in Go's child image viewer window. */
export async function openImage(source: string, alt = ""): Promise<void> {
  try {
    await Posts.openImage(source, alt);
  } catch (error) {
    toast.error(`无法打开图片：${messageOf(error)}`);
  }
}

/** Sorts posts oldest or newest first by reloading the result from Go. */
export async function setPostSortOrder(sortOrder: SortOrder): Promise<void> {
  if (state.postSortOrder === sortOrder) return;
  setState("postSortOrder", sortOrder);
  await loadPosts(0);
}

/** Filters the list by unread or starred posts. */
export async function setPostFilter(filter: {
  unreadOnly?: boolean;
  starredOnly?: boolean;
}): Promise<void> {
  setState({
    unreadOnly: filter.unreadOnly ?? state.unreadOnly,
    starredOnly: filter.starredOnly ?? state.starredOnly,
  });
  await loadPosts(0);
}

/** Opens a post, and marks it read when the preferences ask for that. The
 * feed id is required in the aggregate list; callers can omit it when the
 * selected feed is unambiguous. */
export async function selectPost(guid: string, feedId?: number): Promise<void> {
  const selectedFeedId = feedId ?? state.posts.find((post) => post.guid === guid)?.feedId;
  if (selectedFeedId === undefined) return;

  closeSnapshot();
  setState("current", { postFeedId: selectedFeedId, guid });
  setState("loadingPost", true);
  try {
    const post = await Posts.get(selectedFeedId, guid);
    setState("post", post);
    if (state.prefs.markReadOnOpen && !post.read) {
      await Posts.markRead(selectedFeedId, guid, true);
      setReadLocally(selectedFeedId, guid, true);
    }
  } catch (error) {
    toast.error(`无法打开文章：${messageOf(error)}`);
  } finally {
    setState("loadingPost", false);
  }
}

/** Marks the open post read or unread. */
export async function setOpenPostRead(read: boolean): Promise<void> {
  const post = state.post;
  if (!post) return;

  try {
    await Posts.markRead(post.feedId, post.guid, read);
    setReadLocally(post.feedId, post.guid, read);
    setState("post", "read", read);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Marks every post of a feed read, or every stored post in aggregate mode. */
export async function markAllRead(feedId: number | null = state.current.feedId): Promise<void> {
  try {
    await Posts.markAllRead(feedId ?? aggregateFeedId);
    if (feedId === null || feedId === state.current.feedId) await loadPosts(0);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Stars or unstars a post. The feed id is required in the aggregate list. */
export async function toggleStar(guid: string, feedId?: number): Promise<void> {
  const selectedFeedId = feedId ?? state.posts.find((post) => post.guid === guid)?.feedId;
  if (selectedFeedId === undefined) return;

  try {
    const updated = await Posts.toggleStar(selectedFeedId, guid);
    setState(
      "posts",
      (post) => post.feedId === selectedFeedId && post.guid === guid,
      "starred",
      updated.starred,
    );
    if (state.post?.guid === guid) setState("post", "starred", updated.starred);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

function setReadLocally(feedId: number, guid: string, read: boolean): void {
  const previous = state.posts.find((post) => post.feedId === feedId && post.guid === guid)?.read;
  setState("posts", (post) => post.feedId === feedId && post.guid === guid, "read", read);
  if (previous === undefined || previous === read) return;
  setState("postUnread", (unread) => Math.max(read ? unread - 1 : unread + 1, 0));
}
